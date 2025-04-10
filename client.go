package main

import (
    "log"
    "crypto/tls"
	"crypto/x509"
	 "io/ioutil"
)

func main() {
    log.SetFlags(log.Lshortfile)


	// Load server's certificate
    certPool := x509.NewCertPool()
    cert, err := ioutil.ReadFile("server.crt")
    if err != nil {
        log.Fatal("Failed to read server.crt:", err)
    }
    certPool.AppendCertsFromPEM(cert)

    conf := &tls.Config{
        RootCAs: certPool, // Use the server's certificate
    }


    // conf := &tls.Config{
    //      //InsecureSkipVerify: true,
    // }

    conn, err := tls.Dial("tcp", "127.0.0.1:5555", conf)
    if err != nil {
        log.Println(err)
		log.Println("i'm here first error")
        return
    }
    defer conn.Close()

    n, err := conn.Write([]byte("hello\n"))
    if err != nil {
        log.Println(n, err)
		log.Println("i'm here second error")
        return
    }

    buf := make([]byte, 100)
    n, err = conn.Read(buf)
    if err != nil {
        log.Println(n, err)
		log.Println("i'm here third error")
        return
    }

    println(string(buf[:n]))
}