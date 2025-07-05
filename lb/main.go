package main

func main() {
	proxy := lb{}
	proxy.LoadConfig("config.yaml")
	proxy.start()
}

// import (
// 	"log"
// 	"os"
// )

// func main() {
// 	// Ensure the "certs" folder exists
// 	err := os.MkdirAll("certs", os.ModePerm)
// 	if err != nil {
// 		log.Fatalf("Failed to create certs folder: %v", err)
// 	}

// 	cm := &certManager{}

// 	// Create CA
// 	cm.CreateCA("certs/ca.crt", "certs/ca.key")

// 	// Create a client certificate signed by the CA
// 	cm.CreateClientCert("certs/client.crt", "certs/client.key")
// }
