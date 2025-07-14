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
	"net"
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

	caCertPath := os.Args[2]
	caKeyPath := os.Args[4]

	// Check if CA cert and key exist
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Fatalf("CA certificate not found at %s", caCertPath)
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		log.Fatalf("CA key not found at %s", caKeyPath)
	}

	// Load CA certificate
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}
	block, _ := pem.Decode(caCertPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		log.Fatalf("Failed to decode CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}

	// Load CA private key
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		log.Fatalf("Failed to read CA key: %v", err)
	}
	block, _ = pem.Decode(caKeyPEM)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		log.Fatalf("Failed to decode CA key PEM")
	}
	caKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse CA private key: %v", err)
	}

	cm.caCert = caCert
	cm.caKey = caKey
	
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate client private key: %v", err)
	}

	var extKeyUsage []x509.ExtKeyUsage
	
	if os.Args[5] == "server" {
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	
	if os.Args[5] == "client" {
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"YOUCEF.G"},
			CommonName:   "localhost",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: extKeyUsage,
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
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
