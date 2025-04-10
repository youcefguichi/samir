package main

import (
	"io"
	"log"
	"net"
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Just a plain TCP listener
	ln, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
        
		go forwardConnection(conn)
	}
}

func forwardConnection(conn net.Conn) {
	defer conn.Close()

	targetConn, err := net.Dial("tcp", "localhost:5556")
	if err != nil {
		log.Println("Failed to connect to backend:", err)
		return
	}
	defer targetConn.Close()

	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}
