package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CertResult struct {
	Endpoint  string
	NotBefore time.Time
	NotAfter  time.Time
	DaysLeft  int
	Status    string
	Err       error
}

type TlsResponse struct {
	conn *tls.Conn
	Err  error
}

type Config struct {
	Endpoints          []string `json:"endpoints"`
	Timeout            int      `json:"timeout"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
	CheckInterval      int      `json:"check_interval"`
}

var (
	CertExpiresIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "certificate_expires_in",
			Help: "Number of days until the certificate expires.",
		},
		[]string{
			"domain",
			"valid_from",
			"expires_at",
			"common_name",
			"status"},
	)
)

func CheckCertsExpirationAndExportItAsPrometheusMetrics() {

	prometheus.MustRegister(CertExpiresIn)
	configLocation := os.Getenv("CONFIG_LOCATION")
	
	config := ParseConfigFromJson(configLocation)

	go func() {
		ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf("Performing periodic certificate check (every %d seconds)...\n", config.CheckInterval)
			CheckCertsforAllPeers(config.Endpoints, config.Timeout, config.InsecureSkipVerify)
		}
	}()

	StartCertsMetricsServer()
}

func CheckCertsforAllPeers(endpoints []string, timeoutDurationSeconds int, insecureskip bool) {

	var wg sync.WaitGroup

	CertExpiresIn.Reset()

	for _, endpoint := range endpoints {

		wg.Add(1)

		go func(ep string) {

			defer wg.Done()
			ctx := context.Background()
			response := DialTlsPeer(ctx, timeoutDurationSeconds, ep, insecureskip)

			if response.Err != nil {
				log.Printf("couldn't connect to %s: %v \n", ep, response.Err)
				return
			}

			certs := response.conn.ConnectionState().PeerCertificates

			for _, cert := range certs {

				daysLeft := ExpiresIn(cert)
				CertExpiresIn.WithLabelValues(
					ep,
					cert.NotBefore.Format(time.RFC3339),
					cert.NotAfter.Format(time.RFC3339),
					cert.Subject.CommonName,
					GetCertificateStatus(daysLeft)).Set(float64(daysLeft))

			}

		}(endpoint)

	}

	wg.Wait()

}

func DialTlsPeer(ctx context.Context, timeoutDurationSeconds int, endpoint string, insecureSkip bool) TlsResponse {

	timeout := time.Duration(timeoutDurationSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh := make(chan TlsResponse)

	go func() {
		conn, err := tls.Dial("tcp", endpoint, &tls.Config{InsecureSkipVerify: insecureSkip})
		if err != nil {
			resultCh <- TlsResponse{
				conn: nil,
				Err:  err,
			}
			return
		}

		resultCh <- TlsResponse{
			conn: conn,
			Err:  nil,
		}

	}()

	select {

	case <-ctx.Done():
		return TlsResponse{
			conn: nil,
			Err:  errors.New("timeout while dialing TLS peer"),
		}
	case resp := <-resultCh:
		return resp
	}

}

func ParseConfigFromJson(configLocation string) Config {

	data, err := os.ReadFile(configLocation)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config

	err = json.Unmarshal(data, &config)

	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	return config

}

func GetCertificateStatus(daysLeft int) string {
	switch {
	case daysLeft < 0:
		return "EXPIRED"
	case daysLeft < 30:
		return "URGENT"
	case daysLeft < 90:
		return "WARNING"
	default:
		return "OKAY"
	}
}

func ExpiresIn(cert *x509.Certificate) int {
	daysLeft := time.Until(cert.NotAfter).Hours() / 24
	return int(daysLeft)
}

func StartCertsMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting HTTP server on 0.0.0.0:3005, metrics are serverd at this endpoint /metrics")
	if err := http.ListenAndServe("0.0.0.0:3005", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
