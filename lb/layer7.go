package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type lb struct {
	config         config
	currentBackend int
}

type WafRule struct {
	AllowIPs []string `yaml:"allowedIPs"`
}

type config struct {
	Host           string    `yaml:"host"`
	Port           string    `yaml:"port"`
	Protocol       string    `yaml:"protocol"`
	CertLocation   string    `yaml:"cert_location"`
	KeyLocation    string    `yaml:"key_location"`
	BackendServers []string  `yaml:"backend_servers"`
	Waf            []WafRule `yaml:"waf"`
}

func (l *lb) LoadConfig(configPath string) {

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

	http.HandleFunc("/", l.applyWAF(l.RequestsHandler))
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

		fmt.Println("I am here")
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

func isIPAllowed(ip string, cdir string) bool {

	// if cidr is a single IP, check if it matches
	if strings.Contains(cdir, "/") {
		_, netIP, err := net.ParseCIDR(cdir)
		if err != nil {
			log.Fatal("Error parsing CIDR:", err)
		}

		if !netIP.Contains(net.ParseIP(ip)) {
			return false
		}

	} else {
		// if cidr is a single IP, check if it matches
		if ip != cdir {
			return false
		}
	}

	return true

}

func (l *lb) applyWAF(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		// Check if the request IP is allowed
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Println("Error getting client IP:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, rule := range l.config.Waf {
			if !isIPAllowed(ip, rule.AllowIPs[0]) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		handler(w, req)

	}

}
