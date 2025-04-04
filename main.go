package main

import (
	"fmt"
	"net/http"
	"os"
)

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello, World!</h1>")
	fmt.Println("Received request:", r.Method, r.URL.Path)
	fmt.Println("Received request: TLS", r.TLS)
	fmt.Println("Received request: header", r.Header)
	fmt.Println("Received request: cookies", r.Cookies())
	fmt.Println("Received request: content length", r.ContentLength)
	fmt.Println("Received request: cookies", r.Form)
	fmt.Println("Received request: body", r.GetBody)
	fmt.Println("Received request: host", r.Host)
	fmt.Println("Received request: method", r.Method)
	fmt.Println("Received request: multupart", r.MultipartForm)
	fmt.Println("Received request: pattern", r.Pattern)
	fmt.Println("Received request: post form", r.PostForm)
	fmt.Println("Received request: proto", r.Proto)
	fmt.Println("Received request: proto Mahor", r.ProtoMajor)
	fmt.Println("Received request: MulriForm", r.MultipartForm)
	fmt.Println("Received request: remote address", r.RemoteAddr)
	fmt.Println("Received request: request URI", r.RequestURI)
	fmt.Println("Received request: response", r.Response)
	fmt.Println("Received request: trailer", r.Trailer)
	fmt.Println("Received request: encoding", r.TransferEncoding)
	
	
}

func main() {

	http.HandleFunc("/", root)

	port := ":8080"
	err := http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server started on localhost: %s\n", port)

}
