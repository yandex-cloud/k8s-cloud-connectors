// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s-connectors/pkg/util"
)

var (
	serverSerial = big.NewInt(1984)
	caSerial     = big.NewInt(2020)
	organization = "connectors.cloud.yandex.com"
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

	if err := createCertificates(log, serviceName, namespaceName); err != nil {
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

	if err := signCertificate(client); err != nil {
		log.Error(err, "unable to sign certificate")
		os.Exit(1)
	}

	if err := createSecret(client); err != nil {
		log.Error(err, "unable to create secret with certificate")
		os.Exit(1)
	}

	if err := patchConfig(client); err != nil {
		log.Error(err, "unable to patch config")
		os.Exit(1)
	}
}

func createCertificates(log logr.Logger, service, namespace string) error {
	var caPEM, serverCertPEM, serverPrivateKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: caSerial,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate RSA private key: %v", err)
	}
	log.Info("CA private key created")

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("unable to create self-signed certificate: %v", err)
	}
	log.Info("self-signed CA certificate created")

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return fmt.Errorf("unable to PEM encode self-signed certificate: %v", err)
	}

	dnsNames := []string{
		service,
		service + "." + namespace,
		service + "." + namespace + ".svc",
	}
	commonName := service + "." + namespace + ".svc"

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		// TODO (covariance) read about magic number ang generate it safely
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate RSA server secret key: %v", err)
	}
	log.Info("server private key created")

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("unable to create server certificate: %v", err)
	}
	log.Info("server certificate created")

	// PEM encode the server cert and key
	serverCertPEM = new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err != nil {
		return fmt.Errorf("unable to PEM encode public key: %v", err)
	}

	serverPrivateKeyPEM = new(bytes.Buffer)
	err = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if err != nil {
		return fmt.Errorf("unable to PEM encode secret key: %v", err)
	}

	if err := ioutil.WriteFile("/etc/webhook/certs/tls.crt", serverCertPEM.Bytes(), 0600); err != nil {
		return fmt.Errorf("unable to write public key: %v", err)
	}
	if err := ioutil.WriteFile("/etc/webhook/certs/tls.key", serverPrivateKeyPEM.Bytes(), 0600); err != nil {
		return fmt.Errorf("unable to write secret key: %v", err)
	}

	return nil
}

func signCertificate(cl *kubernetes.Clientset) error {

	return nil
}

func createSecret(cl *kubernetes.Clientset) error {
	return nil
}

func patchConfig(cl *kubernetes.Clientset) error {
	return nil
}
