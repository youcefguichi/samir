package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func main() {
    log.SetFlags(log.Lshortfile)

    cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        log.Println(err)
        return
    }

    config := &tls.Config{Certificates: []tls.Certificate{cer}}
    ln, err := tls.Listen("tcp", ":5556", config) 
    if err != nil {
        log.Println(err)
        return
    }
    defer ln.Close()

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
    r := bufio.NewReader(conn)
    for {
        msg, err := r.ReadString('\n')
        if err != nil {
            log.Println(err)
            return
        }
        
		fmt.Println("--- Layer 7 ---")
        println(msg)
		println(remoteAddr)
        fmt.Println("--- Layer 7 ---")

        n, err := conn.Write([]byte("hello from layer 7\n"))
        if err != nil {
            log.Println(n, err)
            return
        }
    }
}