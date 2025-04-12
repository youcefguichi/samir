package main

import (
	"fmt"
	//"io"
	"log"
	"net/http"
)

type lb struct {
	host           string
	port           string // 5556
	address        string // host:port
	protocol       string // http or https
	certLocation   string
	keyLocation    string
	backendServers []string
	currentBackend int
}

func NewLoadBalancer(host, port, protocol, certLocation, keyLocation string, backendServers []string) *lb {
	return &lb{
		host:           host,
		port:           port,
		address:        fmt.Sprintf("%s:%s", host, port),
		protocol:       protocol,
		certLocation:   certLocation,
		keyLocation:    keyLocation,
		backendServers: backendServers,
		currentBackend: 0,
	}
}

func (l *lb) start() {
	// start the load balancer
	log.Printf("Load balancer is running on %s using protocol %s\n", l.address, l.protocol)
	err := http.ListenAndServeTLS(l.address, l.certLocation, l.keyLocation, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// func (l *lb) sendRequestsToBackends () {
//    for {
//       backend, err := l.getNextBackend("round-robin")
//       if err != nil {
//           log.Println("Error getting next backend:", err)
//           continue
//       }

//       // if connections succeded, Send request to the backend server

//       break

//    }
// }

func (l *lb) getNextBackend() (string, error) {
	// implement logic to get the next backend server (round-robin)

	if l.currentBackend >= len(l.backendServers) {
		l.currentBackend = 0
	}

	backend := l.backendServers[l.currentBackend]

	l.currentBackend++

	return backend, nil
}

// func HelloServer(w http.ResponseWriter, req *http.Request) {
//     backends := []string{
//         "http://localhost:8080",
//         "http://localhost:8081",
//         "http://localhost:8082",
//     }
//     backendURL := backends[0] // Select the first backend for now// Replace with your backend server URL

//     // Create a new request to forward to the backend

//     proxyReq, err := http.NewRequest(req.Method, backendURL+req.URL.Path, req.Body)
//     if err != nil {
//         http.Error(w, "Failed to create request", http.StatusInternalServerError)
//         return
//     }

//     // Copy headers from the original request
//     for key, values := range req.Header {
//         for _, value := range values {
//             proxyReq.Header.Add(key, value)
//         }
//     }

//     // Perform the request to the backend
//     client := &http.Client{}
//     resp, err := client.Do(proxyReq)
//     if err != nil {
//         http.Error(w, "Failed to reach backend server", http.StatusBadGateway)
//         return
//     }
//     defer resp.Body.Close()

//     // Copy the response from the backend to the client
//     for key, values := range resp.Header {
//         for _, value := range values {
//             w.Header().Add(key, value)
//         }
//     }

//     w.WriteHeader(resp.StatusCode)

//     // Copy the backend response body to the client
//     _, err = io.Copy(w, resp.Body)
//     if err != nil {
//         log.Println("Error writing response body:", err)
//     }
// }

// func main() {
//     http.HandleFunc("/", HelloServer)
//     err := http.ListenAndServeTLS(":5556", "server.crt", "server.key", nil)
//     if err != nil {
//         log.Fatal("ListenAndServe: ", err)
//     }
// }
