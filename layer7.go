package main

import (
    // "fmt"
    // "io"
    "net/http"
    "log"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
    backendURL := "http://localhost:8080" // Replace with your backend server URL

    // Create a new request to forward to the backend
    proxyReq, err := http.NewRequest(req.Method, backendURL+req.URL.Path, req.Body)
    if err != nil {
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return
    }

    // Copy headers from the original request
    for key, values := range req.Header {
        for _, value := range values {
            proxyReq.Header.Add(key, value)
        }
    }

    // Perform the request to the backend
    client := &http.Client{}
    resp, err := client.Do(proxyReq)
    if err != nil {
        http.Error(w, "Failed to reach backend server", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    // Copy the response from the backend to the client
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    w.WriteHeader(resp.StatusCode)
    _, err = w.Write([]byte{})
    if err != nil {
        log.Println("Error writing response:", err)
    }
}

func main() {
    http.HandleFunc("/", HelloServer)
    err := http.ListenAndServeTLS(":5556", "server.crt", "server.key", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}


