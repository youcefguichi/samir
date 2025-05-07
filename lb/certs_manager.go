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

// CreateCA generates a Certificate Authority (CA) certificate and private key.
func (cm *certManager) CreateCA(caCertPath, caKeyPath string) {
	// Generate private key for the CA
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate CA private key: %v", err)
	}

	// Create the CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		log.Fatalf("Failed to create CA certificate: %v", err)
	}

	// Save the CA certificate to a file
	caCertFile, err := os.Create(caCertPath)
	if err != nil {
		log.Fatalf("Failed to create CA certificate file: %v", err)
	}
	defer caCertFile.Close()
	pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

	// Save the CA private key to a file
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		log.Fatalf("Failed to create CA key file: %v", err)
	}
	defer caKeyFile.Close()
	caKeyBytes, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		log.Fatalf("Failed to marshal CA private key: %v", err)
	}
	pem.Encode(caKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyBytes})

	// Store CA in the struct
	cm.caCert, err = x509.ParseCertificate(caCertBytes)
	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}
	cm.caKey = caKey
}

// CreateClientCert generates a client certificate signed by the CA.
func (cm *certManager) CreateClientCert(clientCertPath, clientKeyPath string) {
	if cm.caCert == nil || cm.caKey == nil {
		log.Fatal("CA certificate and key must be created first")
	}

	// Generate private key for the client
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate client private key: %v", err)
	}

	// Create the client certificate template
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

	// Sign the client certificate with the CA
	clientCertBytes, err := x509.CreateCertificate(rand.Reader, clientTemplate, cm.caCert, &clientKey.PublicKey, cm.caKey)
	if err != nil {
		log.Fatalf("Failed to create client certificate: %v", err)
	}

	// Save the client certificate to a file
	clientCertFile, err := os.Create(clientCertPath)
	if err != nil {
		log.Fatalf("Failed to create client certificate file: %v", err)
	}
	defer clientCertFile.Close()
	pem.Encode(clientCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes})

	// Save the client private key to a file
	clientKeyFile, err := os.Create(clientKeyPath)
	if err != nil {
		log.Fatalf("Failed to create client key file: %v", err)
	}
	defer clientKeyFile.Close()
	// Save the client private key to a file
	clientKeyBytes, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		log.Fatalf("Failed to marshal client private key: %v", err)
	}
	pem.Encode(clientKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyBytes})
}
