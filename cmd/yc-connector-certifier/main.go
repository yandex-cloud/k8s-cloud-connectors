// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	certificates "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s-connectors/pkg/util"
)

type argList []string

func (r *argList) String() string {
	return fmt.Sprintf("%v", *r)
}

func (r *argList) Set(val string) error {
	*r = append(*r, val)
	return nil
}

func logAndExitOnError(log logr.Logger, err error, msg string) {
	if err != nil {
		log.Error(err, msg)
		os.Exit(1)
	}
}

const (
	serverKeyFile = "server-key.pem"
	serverCSRFile = "server.csr"
	CSRConfigFile = "csr.conf"
)

const CSRConfigTemplate = `[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = %s
DNS.2 = %s
DNS.3 = %s
`

const (
	kubernetesPollInterval = time.Second
	certifierTotalTimeout  = 5 * time.Minute
)

func main() {
	var secretName string
	var serviceName string
	var namespaceName string
	var mutatingWebhooks argList
	var validatingWebhooks argList
	var debug bool
	flag.StringVar(&secretName, "secret", "secret", "Secret to place cert information")
	flag.StringVar(&serviceName, "service", "webhook-service", "Service that is an entrypoint for webhooks")
	flag.StringVar(&namespaceName, "namespace", "default", "Namespace of the service")
	flag.Var(&mutatingWebhooks, "mw", "Names of webhook configurations to be patched")
	flag.Var(&validatingWebhooks, "vw", "Names of webhook configurations to be patched")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging for this connector certifier.")
	flag.Parse()

	log, err := util.NewZaprLogger(debug)
	if err != nil {
		fmt.Printf("unable to set up logger: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), certifierTotalTimeout)
	defer cancel()

	tmpdir, err := ioutil.TempDir("", "cert_tmp_*")
	logAndExitOnError(log, err, "unable to create temporary directory")
	defer func() { _ = os.RemoveAll(tmpdir) }()

	key, err := createSecretKey(log, tmpdir)
	logAndExitOnError(log, err, "unable to generate secret key")

	csr, err := createCertificates(log, tmpdir, serviceName, namespaceName)
	logAndExitOnError(log, err, "unable to create certificate")

	config, err := ctrl.GetConfig()
	logAndExitOnError(log, err, "unable to get kubernetes config")

	client, err := kubernetes.NewForConfig(config)
	logAndExitOnError(log, err, "unable to create kubernetes client from config")

	cert, err := signCertificate(ctx, log, client, serviceName, namespaceName, csr)
	logAndExitOnError(log, err, "unable to sign certificate")

	logAndExitOnError(
		log,
		createSecret(ctx, log, client, namespaceName, secretName, key, cert),
		"unable to create secret with certificate",
	)

	for _, mutatingWebhook := range mutatingWebhooks {
		log.Info("patching mutating webhook: " + mutatingWebhook)
		logAndExitOnError(
			log,
			patchMutatingConfig(ctx, client, mutatingWebhook, namespaceName, cert),
			"unable to patch config",
		)
	}

	for _, validatingWebhook := range validatingWebhooks {
		log.Info("patching validating webhook: " + validatingWebhook)
		logAndExitOnError(
			log,
			patchValidatingConfig(ctx, client, validatingWebhook, namespaceName, cert),
			"unable to patch config",
		)
	}
}

func createSecretKey(log logr.Logger, tmpdir string) ([]byte, error) {
	genRSACmd := exec.Command("openssl", "genrsa",
		"-out", serverKeyFile,
		"2048",
	)
	genRSACmd.Dir = tmpdir
	if err := genRSACmd.Run(); err != nil {
		return nil, fmt.Errorf("unable to generate RSA key: %v", err)
	}
	log.Info("RSA key generated")

	return ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, serverKeyFile)))
}

func createCertificates(log logr.Logger, tmpdir, service, namespace string) ([]byte, error) {
	csrConf := fmt.Sprintf(CSRConfigTemplate, service, service+"."+namespace, service+"."+namespace+".svc")

	if err := ioutil.WriteFile(filepath.Clean(filepath.Join(tmpdir, CSRConfigFile)), []byte(csrConf), 0600); err != nil {
		return nil, fmt.Errorf("unable to create CSR configuration file: %v", err)
	}
	log.Info("certificate configuration created")

	createReq := exec.Command("openssl", "req",
		"-new",
		"-key", serverKeyFile,
		"-subj", "/CN="+service+"."+namespace+".svc",
		"-config", CSRConfigFile,
		"-out", serverCSRFile,
	) // #nosec G204
	createReq.Dir = tmpdir
	if err := createReq.Run(); err != nil {
		return nil, fmt.Errorf("unable to create server CSR: %v", err)
	}
	log.Info("server CSR created")

	csr, err := ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, serverCSRFile)))
	if err != nil {
		return nil, fmt.Errorf("unable to read created CSR: %v", err)
	}

	return csr, nil
}

func createCSR(
	ctx context.Context,
	cl v1beta1.CertificateSigningRequestInterface,
	name,
	namespace string,
	bytes []byte,
) (*certificates.CertificateSigningRequest, error) {
	return cl.Create(
		ctx,
		&certificates.CertificateSigningRequest{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: certificates.CertificateSigningRequestSpec{
				Request: bytes,
				Usages: []certificates.KeyUsage{
					certificates.UsageDigitalSignature,
					certificates.UsageKeyEncipherment,
					certificates.UsageServerAuth,
				},
				Groups: []string{"system:authenticated"},
			},
		},
		metav1.CreateOptions{},
	)
}

func waitForDeletion(
	ctx context.Context,
	log logr.Logger,
	getter func(innerCtx context.Context) (*certificates.CertificateSigningRequest, error),
) error {
	log.Info("old CSR found, waiting for its deletion to be completed")
	for {
		if _, err := getter(ctx); err != nil {
			if errors.IsNotFound(err) {
				break
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kubernetesPollInterval):
			log.Info("deletion is not completed, waiting")
		}
	}
	log.Info("deletion is completed")
	return nil
}

func waitForCreation(
	ctx context.Context,
	log logr.Logger,
	getter func(innerCtx context.Context) (*certificates.CertificateSigningRequest, error),
) error {
	log.Info("waiting for CSR creation")
	for {
		_, err := getter(ctx)
		if err == nil {
			break
		}
		if !errors.IsNotFound(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kubernetesPollInterval):
			log.Info("creation is not completed, waiting")
		}
	}
	log.Info("creation is completed")
	return nil
}

func signCertificate(
	ctx context.Context,
	log logr.Logger,
	cl *kubernetes.Clientset,
	namespace,
	service string,
	csrBytes []byte,
) ([]byte, error) {
	csrName := service + "." + namespace + ".csr"
	csrClient := cl.CertificatesV1beta1().CertificateSigningRequests()

	getCsr := func(innerCtx context.Context) (*certificates.CertificateSigningRequest, error) {
		return csrClient.Get(innerCtx, csrName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		})
	}

	if err := csrClient.Delete(
		ctx,
		csrName,
		metav1.DeleteOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		},
	); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to delete previous CSR %v", err)
		}
		log.Info("old CSR not found")
	} else if err := waitForDeletion(ctx, log, getCsr); err != nil {
		return nil, fmt.Errorf("error while waiting for old CSR deletion: %v", err)
	}

	log.Info("creating new CSR")
	csr, err := createCSR(ctx, csrClient, csrName, namespace, csrBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to create CSR: %v", err)
	}

	if err := waitForCreation(ctx, log, getCsr); err != nil {
		return nil, fmt.Errorf("error while waiting for CSR creation: %v", err)
	}

	log.Info("approving CSR")
	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:           certificates.CertificateApproved,
		LastUpdateTime: metav1.Now(),
	})

	if _, err := csrClient.UpdateApproval(ctx, csr, metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1beta1",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to approve CSR: %v", err)
	}

	log.Info("waiting for CSR to be approved")
	var cert []byte
	for {
		res, err := getCsr(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while waiting for CSR approval: %v", err)
		}
		if res.Status.Certificate != nil && len(res.Status.Certificate) != 0 {
			cert = res.Status.Certificate
			log.Info("CSR is approved")
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for CSR approval interrupted: %v", ctx.Err())
		case <-time.After(kubernetesPollInterval):
			log.Info("CSR approval is not completed, waiting")
		}
	}

	return cert, nil
}

func createSecret(
	ctx context.Context,
	log logr.Logger,
	cl *kubernetes.Clientset,
	namespace,
	secret string,
	key,
	cert []byte,
) error {
	secretTypeMeta := metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}

	if _, err := cl.CoreV1().Secrets(namespace).Create(ctx, &v1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"tls.key": key,
			"tls.crt": cert,
		},
	}, metav1.CreateOptions{
		TypeMeta: secretTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to create secret: %v", err)
	}
	log.Info("secret with key and cert successfully created")

	return nil
}

func patchMutatingConfig(
	ctx context.Context,
	cl *kubernetes.Clientset,
	webhook,
	namespace string,
	caBundle []byte,
) error {
	confClient := cl.AdmissionregistrationV1().MutatingWebhookConfigurations()
	confTypeMeta := metav1.TypeMeta{
		Kind:       "MutatingWebhookConfiguration",
		APIVersion: "admissionregistration.k8s.io/v1",
	}

	conf, err := confClient.Get(ctx, webhook, metav1.GetOptions{
		TypeMeta: confTypeMeta,
	})
	if err != nil {
		return fmt.Errorf("unable to get webhook configuration: %v", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.Service.Namespace = namespace
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(ctx, conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %v", err)
	}

	return nil
}

func patchValidatingConfig(
	ctx context.Context,
	cl *kubernetes.Clientset,
	webhook,
	namespace string,
	caBundle []byte,
) error {
	confClient := cl.AdmissionregistrationV1().ValidatingWebhookConfigurations()
	confTypeMeta := metav1.TypeMeta{
		Kind:       "ValidatingWebhookConfiguration",
		APIVersion: "admissionregistration.k8s.io/v1",
	}

	conf, err := confClient.Get(ctx, webhook, metav1.GetOptions{
		TypeMeta: confTypeMeta,
	})
	if err != nil {
		return fmt.Errorf("unable to get webhook configuration: %v", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.Service.Namespace = namespace
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(ctx, conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %v", err)
	}

	return nil
}
