package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Print request details
	fmt.Printf("Method: %s, URL: %s, Header: %v\n", r.Method, r.URL, r.Header)
	// Respond to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request received"))
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}