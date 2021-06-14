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

	cert, err := signCertificate(log, client, serviceName, namespaceName, csr)
	logAndExitOnError(log, err, "unable to sign certificate")

	logAndExitOnError(
		log,
		createSecret(log, client, namespaceName, secretName, key, cert),
		"unable to create secret with certificate",
	)

	for _, mutatingWebhook := range mutatingWebhooks {
		log.Info("patching mutating webhook: " + mutatingWebhook)
		logAndExitOnError(
			log,
			patchMutatingConfig(client, mutatingWebhook, cert),
			"unable to patch config",
		)
	}

	for _, validatingWebhook := range validatingWebhooks {
		log.Info("patching validating webhook: " + validatingWebhook)
		logAndExitOnError(
			log,
			patchValidatingConfig(client, validatingWebhook, cert),
			"unable to patch config",
		)
	}
}

func createSecretKey(log logr.Logger, tmpdir string) ([]byte, error) {
	genRSACmd := exec.Command("openssl", "genrsa",
		"-out", "server-key.pem",
		"2048",
	)
	genRSACmd.Dir = tmpdir
	if err := genRSACmd.Run(); err != nil {
		return nil, fmt.Errorf("unable to generate RSA key: %v", err)
	}
	log.Info("RSA key generated")

	return ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, "server-key.pem")))
}

func createCertificates(log logr.Logger, tmpdir, service, namespace string) ([]byte, error) {
	csrConf := fmt.Sprintf("[req]\n" +
		"req_extensions = v3_req\n" +
		"distinguished_name = req_distinguished_name\n" +
		"[req_distinguished_name]\n" +
		"[ v3_req ]\n" +
		"basicConstraints = CA:FALSE\n" +
		"keyUsage = nonRepudiation, digitalSignature, keyEncipherment\n" +
		"extendedKeyUsage = serverAuth\n" +
		"subjectAltName = @alt_names\n" +
		"[alt_names]\n" +
		"DNS.1 = " + service + "\n" +
		"DNS.2 = " + service + "." + namespace + "\n" +
		"DNS.3 = " + service + "." + namespace + ".svc\n")

	if err := ioutil.WriteFile(tmpdir+"/csr.conf", []byte(csrConf), 0600); err != nil {
		return nil, fmt.Errorf("unable to create csr configuration file: %v", err)
	}
	log.Info("certificate configuration created")

	createReq := exec.Command("openssl", "req",
		"-new",
		"-key", "server-key.pem",
		"-subj", "/CN="+service+"."+namespace+".svc",
		"-config", "csr.conf",
		"-out", "server.csr",
	) // #nosec G204
	createReq.Dir = tmpdir
	if err := createReq.Run(); err != nil {
		return nil, fmt.Errorf("unable to create server csr: %v", err)
	}
	log.Info("server CSR created")

	csr, err := ioutil.ReadFile(filepath.Clean(filepath.Join(tmpdir, "server.csr")))
	if err != nil {
		return nil, fmt.Errorf("unable to read created CSR: %v", err)
	}

	return csr, nil
}

func createCSR(
	cl v1beta1.CertificateSigningRequestInterface,
	name,
	namespace string,
	bytes []byte,
) (*certificates.CertificateSigningRequest, error) {
	return cl.Create(
		context.TODO(),
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

func signCertificate(
	log logr.Logger,
	cl *kubernetes.Clientset,
	namespace,
	service string,
	csrBytes []byte,
) ([]byte, error) {
	csrName := service + "." + namespace + ".csr"
	csrClient := cl.CertificatesV1beta1().CertificateSigningRequests()

	getCsr := func() (*certificates.CertificateSigningRequest, error) {
		return csrClient.Get(context.TODO(), csrName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CertificateSigningRequest",
				APIVersion: "certificates.k8s.io/v1beta1",
			},
		})
	}

	if err := csrClient.Delete(
		context.TODO(),
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
	} else {
		log.Info("old CSR found, waiting for its deletion to be completed")
		for {
			time.Sleep(time.Second)
			if _, err := getCsr(); err != nil {
				if errors.IsNotFound(err) {
					log.Info("deletion is completed")
					break
				}
				return nil, fmt.Errorf("unable wait for old CSR deletion: %v", err)
			}
			log.Info("deletion is not completed, waiting")
		}
	}

	log.Info("creating new CSR")
	csr, err := createCSR(csrClient, csrName, namespace, csrBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to create CSR: %v", err)
	}

	log.Info("waiting for CSR creation")
	for {
		time.Sleep(time.Second)
		if _, err := getCsr(); err != nil {
			if errors.IsNotFound(err) {
				log.Info("creation is not completed, waiting")
				continue
			}
			return nil, fmt.Errorf("unable to get CSR: %v", err)
		}
		log.Info("creation is completed")
		break
	}

	log.Info("approving CSR")
	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:           certificates.CertificateApproved,
		LastUpdateTime: metav1.Now(),
	})

	if _, err := csrClient.UpdateApproval(context.TODO(), csr, metav1.UpdateOptions{
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
		time.Sleep(time.Second)
		res, err := getCsr()
		if err != nil {
			return nil, fmt.Errorf("unable to get CSR: %v", err)
		}
		if res.Status.Certificate != nil && len(res.Status.Certificate) != 0 {
			cert = res.Status.Certificate
			log.Info("CSR is approved")
			break
		}
		log.Info("CSR approval is not completed, waiting")
	}

	return cert, nil
}

func createSecret(log logr.Logger, cl *kubernetes.Clientset, namespace, secret string, key, cert []byte) error {
	secretTypeMeta := metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}

	if _, err := cl.CoreV1().Secrets(namespace).Create(context.TODO(), &v1.Secret{
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

func patchMutatingConfig(cl *kubernetes.Clientset, webhook string, caBundle []byte) error {
	confClient := cl.AdmissionregistrationV1().MutatingWebhookConfigurations()
	confTypeMeta := metav1.TypeMeta{
		Kind:       "MutatingWebhookConfiguration",
		APIVersion: "admissionregistration.k8s.io/v1",
	}

	conf, err := confClient.Get(context.TODO(), webhook, metav1.GetOptions{
		TypeMeta: confTypeMeta,
	})
	if err != nil {
		return fmt.Errorf("unable to get webhook configuration: %v", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(context.TODO(), conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %v", err)
	}

	return nil
}

func patchValidatingConfig(cl *kubernetes.Clientset, webhook string, caBundle []byte) error {
	confClient := cl.AdmissionregistrationV1().ValidatingWebhookConfigurations()
	confTypeMeta := metav1.TypeMeta{
		Kind:       "ValidatingWebhookConfiguration",
		APIVersion: "admissionregistration.k8s.io/v1",
	}

	conf, err := confClient.Get(context.TODO(), webhook, metav1.GetOptions{
		TypeMeta: confTypeMeta,
	})
	if err != nil {
		return fmt.Errorf("unable to get webhook configuration: %v", err)
	}

	for i := range conf.Webhooks {
		conf.Webhooks[i].ClientConfig.CABundle = caBundle
	}

	if _, err := confClient.Update(context.TODO(), conf, metav1.UpdateOptions{
		TypeMeta: confTypeMeta,
	}); err != nil {
		return fmt.Errorf("unable to update webhook configuration: %v", err)
	}

	return nil
}
