package main

import (
	"log"
	"os"
)

func main() {
	err := os.MkdirAll("certs", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create certs folder: %v", err)
	}

	cm := &certManager{}


	if os.Args[1] == "--ca" {
		cm.CreateCA("certs/"+ os.Args[2], "certs/"+ os.Args[3])
		return
	}

	if os.Args[5] == "server" || os.Args[5] == "client" {
		cm.CreateClientCert("certs/"+ os.Args[6], "certs/"+ os.Args[7])
		return
	}
}
