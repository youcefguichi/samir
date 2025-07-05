package main

import (
	"crypto/tls"
	"crypto/x509"
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

type Waf struct {
	AllowedIPs                []string            `yaml:"allowedIPs"`
	AllowedPorts              []string            `yaml:"allowedPorts"`
}

type config struct {
	Host           string   `yaml:"host"`
	Port           string   `yaml:"port"`
	Protocol       string   `yaml:"protocol"`
	CaLocation     string   `yaml:"ca_location"`
	CertLocation   string   `yaml:"cert_location"`
	KeyLocation    string   `yaml:"key_location"`
	BackendServers []string `yaml:"backend_servers"`
	Waf            Waf      `yaml:"waf"`
}

func (l *lb) LoadConfig(configPath string) {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting working directory: \n", err)
	}
	data, err := os.ReadFile(wd + configPath)
	if err != nil {
		log.Fatal("Error reading config file: \n", err)
	}

	var cfg config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("Error unmarshalling config file: \n", err)
	}

	fmt.Println(cfg)

	l.config = cfg
}

func (l *lb) start() {

	http.HandleFunc("/", l.applyWAF(l.RequestsHandler))
	log.Printf("Load balancer is running on %s using protocol %s\n", l.config.Host, l.config.Protocol)

	caCert, _ := os.ReadFile(l.config.CaLocation)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      l.config.Host,
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(l.config.CertLocation, l.config.KeyLocation)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// ask client to present client cert
	// if the client does not present a cert, return 403
	// if authenticaions succedded do the authorization
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

func isIPAllowed(ip string, cdir []string) bool {
	found := false
	for _, cdir := range cdir {

		if !strings.Contains(cdir, "/") && cdir == ip {
			found = true
			break
		}

		_, netIP, err := net.ParseCIDR(cdir)

		if err != nil {
			log.Fatal("Error parsing CIDR:", err)
		}

		if netIP.Contains(net.ParseIP(ip)) {
			found = true
			break
		}
	}

	return found
}

func isPortAllowed(port string, ports []string) bool {
	found := false
	for _, p := range ports {

		if p == port {
			found = true
			break
		}

		if p == "*" {
			found = true
			break
		}
	}
	return found
}


func (l *lb) applyWAF(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		ip, port, err := net.SplitHostPort(req.RemoteAddr)

		if err != nil {
			log.Println("Error getting client IP:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// WAF check

		if !isPortAllowed(port, l.config.Waf.AllowedPorts) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !isIPAllowed(ip, l.config.Waf.AllowedIPs) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		handler(w, req)

	}

}
