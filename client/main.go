package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Load CA certificate
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal("Failed to read ca.crt:", err)
	}
	certPool.AppendCertsFromPEM(caCert)

	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatal("Failed to load client certificate/key:", err)
	}

	// TLS config with client certs
	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
	}

	// HTTP client with custom transport
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	for {
		resp, err := client.Get("https://127.0.0.1:5555/")
		if err != nil {
			log.Println("HTTP request error:", err)
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Printf("Response: %s", string(body))
			resp.Body.Close()
		}
		time.Sleep(3 * time.Second)
	}
}
