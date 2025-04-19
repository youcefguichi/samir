package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"gopkg.in/yaml.v3"
)


type lb struct {
   config config
   currentBackend int
}

type config struct {
	Host           string   `yaml:"host"`
	Port           string   `yaml:"port"`
	Protocol       string   `yaml:"protocol"`
	CertLocation   string   `yaml:"cert_location"`
	KeyLocation    string   `yaml:"key_location"`
	BackendServers []string `yaml:"backend_servers"`
}


func (l *lb) LoadConfig(configPath string){

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("Error reading config file: \n", err)
	}

	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("Error unmarshalling config file: \n", err)
	}

	l.config = cfg
}

func (l *lb) start() {
	http.HandleFunc("/", l.RequestsHandler)
    
	// lodd the configuration file
	// start the load balancer
	log.Printf("Load balancer is running on %s using protocol %s\n", l.config.Host, l.config.Protocol)
	err := http.ListenAndServeTLS(l.config.Host, l.config.CertLocation, l.config.KeyLocation, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}


func (l *lb) RequestsHandler(w http.ResponseWriter, req *http.Request) {
	for {

		backend, err := l.getNextBackend()
		if err != nil {
			log.Println("Error getting next backend:", err)
		}

		backend = fmt.Sprintf("%s%s", backend, req.URL.Path)
		proxyReq, err := http.NewRequest(req.Method, backend, req.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		for key, values := range req.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Println("Error sending request to backend:", err)
			continue
		}
		defer resp.Body.Close()

		// Copy the response from the backend to the client
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)

		// Copy the backend response body to the client
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println("Error writing response body:", err)
		}

		break
	}
}

func (l *lb) getNextBackend() (string, error) {
	// implement logic to get the next backend server (round-robin)

	if l.currentBackend >= len(l.config.BackendServers) {
		l.currentBackend = 0
	}

	backend := l.config.BackendServers[l.currentBackend]

	l.currentBackend++

	return backend, nil
}
