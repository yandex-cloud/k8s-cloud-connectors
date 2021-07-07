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
	typed "k8s.io/client-go/kubernetes/typed/core/v1"
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

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,resourceNames=kubernetes.io/*,verbs=approve

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

	if err := execute(log, secretName, serviceName, namespaceName, mutatingWebhooks, validatingWebhooks); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}

func execute(log logr.Logger, secretName, serviceName, namespaceName string, mutating, validating []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), certifierTotalTimeout)
	defer cancel()

	tmpdir, err := ioutil.TempDir("", "cert_tmp_*")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpdir) }()

	key, err := createSecretKey(log, tmpdir)
	if err != nil {
		return fmt.Errorf("unable to generate secret key: %w", err)
	}

	csr, err := createCertificates(log, tmpdir, serviceName, namespaceName)
	if err != nil {
		return fmt.Errorf("unable to create certificate: %w", err)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("unable to get kubernetes config: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes client from config: %w", err)
	}

	cert, err := signCertificate(ctx, log, client, serviceName, namespaceName, csr)
	if err != nil {
		return fmt.Errorf("unable to sign certificate: %w", err)
	}

	if err := putKeyAndCertToSecret(
		ctx,
		log,
		client.CoreV1().Secrets(namespaceName),
		namespaceName,
		secretName,
		key,
		cert,
	); err != nil {
		return fmt.Errorf("unable to create secret with certificate: %w", err)
	}

	for _, webhook := range mutating {
		log.Info("patching mutating webhook: " + webhook)
		if err := patchMutatingConfig(ctx, client, webhook, namespaceName, cert); err != nil {
			return fmt.Errorf("unable to patch config: %w", err)
		}
	}

	for _, webhook := range validating {
		log.Info("patching validating webhook: " + webhook)
		if err := patchValidatingConfig(ctx, client, webhook, namespaceName, cert); err != nil {
			return fmt.Errorf("unable to patch config: %w", err)
		}
	}

	return nil
}

func createSecretKey(log logr.Logger, tmpdir string) ([]byte, error) {
	genRSACmd := exec.Command("openssl", "genrsa",
		"-out", serverKeyFile,
		"2048",
	)
	genRSACmd.Dir = tmpdir
	if err := genRSACmd.Run(); err != nil {
		return nil, fmt.Errorf("unable to generate RSA key: %w", err)
	}
	log.Info("RSA key generated")

	return ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, serverKeyFile)))
}

func createCertificates(log logr.Logger, tmpdir, service, namespace string) ([]byte, error) {
	csrConf := fmt.Sprintf(CSRConfigTemplate, service, service+"."+namespace, service+"."+namespace+".svc")

	if err := ioutil.WriteFile(filepath.Clean(filepath.Join(tmpdir, CSRConfigFile)), []byte(csrConf), 0600); //nolint:gomnd
	err != nil {
		return nil, fmt.Errorf("unable to create CSR configuration file: %w", err)
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
		return nil, fmt.Errorf("unable to create server CSR: %w", err)
	}
	log.Info("server CSR created")

	csr, err := ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, serverCSRFile)))
	if err != nil {
		return nil, fmt.Errorf("unable to read created CSR: %w", err)
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
	csrClient v1beta1.CertificateSigningRequestInterface,
	csrName string,
) error {
	log.Info("old CSR found, waiting for its deletion to be completed")
	for {
		if _, err := csrClient.Get(ctx, csrName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		}); err != nil {
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
	csrClient v1beta1.CertificateSigningRequestInterface,
	csrName string,
) error {
	log.Info("waiting for CSR creation")
	for {
		_, err := csrClient.Get(ctx, csrName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		})
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
			return nil, fmt.Errorf("unable to delete previous CSR %w", err)
		}
		log.Info("old CSR not found")
	} else if err := waitForDeletion(ctx, log, csrClient, csrName); err != nil {
		return nil, fmt.Errorf("error while waiting for old CSR deletion: %w", err)
	}

	log.Info("creating new CSR")
	csr, err := createCSR(ctx, csrClient, csrName, namespace, csrBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to create CSR: %w", err)
	}

	if err := waitForCreation(ctx, log, csrClient, csrName); err != nil {
		return nil, fmt.Errorf("error while waiting for CSR creation: %w", err)
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
		return nil, fmt.Errorf("unable to approve CSR: %w", err)
	}

	log.Info("waiting for CSR to be approved")
	var cert []byte
	for {
		res, err := csrClient.Get(ctx, csrName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error while waiting for CSR approval: %w", err)
		}
		if res.Status.Certificate != nil && len(res.Status.Certificate) != 0 {
			cert = res.Status.Certificate
			log.Info("CSR is approved")
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for CSR approval interrupted: %w", ctx.Err())
		case <-time.After(kubernetesPollInterval):
			log.Info("CSR approval is not completed, waiting")
		}
	}

	return cert, nil
}

func putKeyAndCertToSecret(
	ctx context.Context,
	log logr.Logger,
	si typed.SecretInterface,
	namespace,
	secret string,
	key,
	cert []byte,
) error {
	secretType := metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}

	sec := v1.Secret{
		TypeMeta: secretType,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"tls.key": key,
			"tls.crt": cert,
		},
	}

	if _, err := si.Get(ctx, secret, metav1.GetOptions{
		TypeMeta: secretType,
	}); err != nil {
		if errors.IsNotFound(err) {
			log.Info("secret does not exist, creating")
			return createSecret(ctx, log, si, &sec)
		}
		return fmt.Errorf("unable to get secret: %w", err)
	}

	log.Info("secret already exists, updating")
	return updateSecret(ctx, log, si, &sec)
}

func updateSecret(
	ctx context.Context,
	log logr.Logger,
	si typed.SecretInterface,
	sec *v1.Secret,
) error {
	if _, err := si.Update(ctx, sec, metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}); err != nil {
		return fmt.Errorf("unable to update secret: %w", err)
	}
	log.Info("secret with key and cert successfully updated")

	return nil
}

func createSecret(
	ctx context.Context,
	log logr.Logger,
	si typed.SecretInterface,
	sec *v1.Secret,
) error {
	if _, err := si.Create(ctx, sec, metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}); err != nil {
		return fmt.Errorf("unable to create secret: %w", err)
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
		return fmt.Errorf("unable to get webhook configuration: %w", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.Service.Namespace = namespace
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(ctx, conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %w", err)
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
		return fmt.Errorf("unable to get webhook configuration: %w", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.Service.Namespace = namespace
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(ctx, conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %w", err)
	}

	return nil
}
