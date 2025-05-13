package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

type certManager struct {
	caCert *x509.Certificate
	caKey  *ecdsa.PrivateKey
}

type Users struct {

}


func (cm *certManager) CreateCA(caCertPath, caKeyPath string) {

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Fatalf("Failed to generate CA private key: %v", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"YOUCEF.G"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	signedCA, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		log.Fatalf("Failed to create CA certificate: %v", err)
	}

	caCertFile, err := os.Create(caCertPath)

	if err != nil {
		log.Fatalf("Failed to create CA certificate file: %v", err)
	}

	pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCA})
	defer caCertFile.Close()

	caKeyFile, err := os.Create(caKeyPath)

	if err != nil {
		log.Fatalf("Failed to create CA key file: %v", err)
	}

	caKeyBytes, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		log.Fatalf("Failed to marshal CA private key: %v", err)
	}

	pem.Encode(caKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyBytes})
	defer caKeyFile.Close()

	cm.caCert, err = x509.ParseCertificate(signedCA)

	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}

	cm.caKey = caKey
}

func (cm *certManager) CreateClientCert(clientCertPath, clientKeyPath string) {

	if cm.caCert == nil || cm.caKey == nil {
		log.Fatal("CA certificate and key must be created first")
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate client private key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Client"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	signedClientCert, err := x509.CreateCertificate(rand.Reader, clientTemplate, cm.caCert, &clientKey.PublicKey, cm.caKey)
	if err != nil {
		log.Fatalf("Failed to create client certificate: %v", err)
	}

	clientCertFile, err := os.Create(clientCertPath)
	if err != nil {
		log.Fatalf("Failed to create client certificate file: %v", err)
	}

	pem.Encode(clientCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedClientCert})
	defer clientCertFile.Close()

	clientKeyFile, err := os.Create(clientKeyPath)
	if err != nil {
		log.Fatalf("Failed to create client key file: %v", err)
	}

	clientKeyBytes, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		log.Fatalf("Failed to marshal client private key: %v", err)
	}

	pem.Encode(clientKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyBytes})
	defer clientKeyFile.Close()
}



