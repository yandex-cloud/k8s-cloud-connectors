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
	"time"

	"github.com/go-logr/logr"
	certificates "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

func main() {
	var secretName string
	var serviceName string
	var namespaceName string
	var webhooks argList
	var debug bool
	flag.StringVar(&secretName, "secret", "secret", "Secret to place cert information")
	flag.StringVar(&serviceName, "service", "webhook-service", "Service that is an entrypoint for webhooks")
	flag.StringVar(&namespaceName, "namespace", "default", "Namespace of the service")
	flag.Var(&webhooks, "webhooks", "Names of webhook configurations to be patched")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging for this connector certifier.")
	flag.Parse()

	log, err := util.NewZaprLogger(debug)
	if err != nil {
		fmt.Printf("unable to set up logger: %v", err)
		os.Exit(1)
	}

	csr, err := createCertificates(log, serviceName, namespaceName)
	if err != nil {
		log.Error(err, "unable to create certificate")
		os.Exit(1)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		log.Error(err, "unable to get kubernetes config")
		os.Exit(1)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "unable to create kubernetes client from config")
		os.Exit(1)
	}

	cert, err := signCertificate(log, client, serviceName, namespaceName, csr)
	if err != nil {
		log.Error(err, "unable to sign certificate")
		os.Exit(1)
	}

	if err := createSecret(log, client, namespaceName, secretName, csr, cert); err != nil {
		log.Error(err, "unable to create secret with certificate")
		os.Exit(1)
	}
}

func createCertificates(log logr.Logger, service, namespace string) ([]byte, error) {
	tmpdir, err := ioutil.TempDir("", "cert_tmp_*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temporary directory: %v", err)
	}

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

	genRSACmd := exec.Command("openssl", "genrsa",
		"-out", "server-key.pem",
		"2048",
	)
	genRSACmd.Dir = tmpdir
	if err := genRSACmd.Run(); err != nil {
		return nil, fmt.Errorf("unable to generate RSA key: %v", err)
	}
	log.Info("RSA key generated")

	createReq := exec.Command("openssl", "req",
		"-new",
		"-key", "server-key.pem",
		"-subj", "/CN="+service+"."+namespace+".svc",
		"-config", "csr.conf",
		"-out", "server.csr",
	)
	createReq.Dir = tmpdir
	if err := createReq.Run(); err != nil {
		return nil, fmt.Errorf("unable to create server csr: %v", err)
	}
	log.Info("server CSR created")

	csr, err := ioutil.ReadFile(tmpdir + "/server.csr")
	if err != nil {
		return nil, fmt.Errorf("unable to read created CSR: %v", err)
	}

	return csr, nil
}

// TODO (covariance) write logging instead of comments
func signCertificate(_ logr.Logger, cl *kubernetes.Clientset, namespace, service string, csrBytes []byte) (
	[]byte, error,
) {
	csrName := service + "." + namespace + ".csr"

	csrClient := cl.CertificatesV1beta1().CertificateSigningRequests()
	csrTypeMeta := metav1.TypeMeta{
		Kind:       "CertificateSigningRequest",
		APIVersion: "certificates.k8s.io/v1beta1",
	}

	// We delete old CSR
	if err := csrClient.Delete(
		context.TODO(),
		csrName,
		metav1.DeleteOptions{
			TypeMeta: csrTypeMeta,
		}); err != nil {
		// If it is present, and something failed, we exit with error
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to delete previous CSR %v", err)
		}
	} else {
		// Otherwise, deletion was successful, so we need to wait for its completion
		for {
			time.Sleep(time.Second)
			_, err := csrClient.Get(context.TODO(), csrName, metav1.GetOptions{
				TypeMeta: csrTypeMeta,
			})
			if err != nil {
				if errors.IsNotFound(err) {
					break
				}
				return nil, fmt.Errorf("unable to get CSR: %v", err)
			}
		}
	}

	// We create new CSR
	csr, err := csrClient.Create(context.TODO(), &certificates.CertificateSigningRequest{
		TypeMeta: csrTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      csrName,
			Namespace: namespace,
		},
		Spec: certificates.CertificateSigningRequestSpec{
			Request: csrBytes,
			Usages: []certificates.KeyUsage{
				certificates.UsageDigitalSignature,
				certificates.UsageKeyEncipherment,
				certificates.UsageServerAuth,
			},
			Groups: []string{"system:authenticated"},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create CSR: %v", err)
	}

	// Again, need to wait some time for operation to finish
	for {
		time.Sleep(time.Second)
		_, err := csrClient.Get(context.TODO(), csrName, metav1.GetOptions{
			TypeMeta: csrTypeMeta,
		})
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("unable to get CSR: %v", err)
		}
		break
	}

	// Then we try to approve CSR
	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:               certificates.CertificateApproved,
		LastUpdateTime:     metav1.Now(),
	})

	if _, err := csrClient.UpdateApproval(context.TODO(), csr, metav1.UpdateOptions{
		TypeMeta: csrTypeMeta,
	}); err != nil {
		return nil, fmt.Errorf("unable to approve CSR: %v", err)
	}

	// And again, wait for changes to propagate
	var cert []byte
	for {
		time.Sleep(time.Second)
		res, err := csrClient.Get(context.TODO(), csrName, metav1.GetOptions{
			TypeMeta: csrTypeMeta,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to get CSR: %v", err)
		}
		if res.Status.Certificate != nil && len(res.Status.Certificate) != 0 {
			cert = res.Status.Certificate
			break
		}
	}

	return cert, nil
}

func createSecret(_ logr.Logger, cl *kubernetes.Clientset, namespace, secret string, key, cert []byte) error {
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

	return nil
}
