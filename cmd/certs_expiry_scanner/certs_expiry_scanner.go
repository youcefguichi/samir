package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	//"math"
	//"runtime"
	context "context"
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

func expiresIn(cert *x509.Certificate) int {

	daysLeft := time.Until(cert.NotAfter).Hours() / 24
	return int(daysLeft)
}

func CheckCertsforAllPeers(endpoints []string, timeoutDurationSeconds int, insecureskip bool) {

	var wg sync.WaitGroup
	resultCh := make(chan CertResult, len(endpoints))

	t := configureCertExpiryReportTable()

	for _, endpoint := range endpoints {

		wg.Add(1)

		go func(ep string) {
			defer wg.Done()

			ctx := context.Background()
			response := DialTlsPeer(ctx, timeoutDurationSeconds, ep, insecureskip)

			if response.Err != nil {
				resultCh <- CertResult{
					Endpoint: ep,
					Err:      response.Err,
				}
				return
			}

			certs := response.conn.ConnectionState().PeerCertificates

			for _, cert := range certs {

				daysLeft := expiresIn(cert)

				resultCh <- CertResult{
					Endpoint:  ep,
					NotBefore: cert.NotBefore,
					NotAfter:  cert.NotAfter,
					DaysLeft:  daysLeft,
					Status:    certStatus(daysLeft),
					Err:       nil,
				}

			}

		}(endpoint)

	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		if result.Err != nil {
			fmt.Printf("X %s — %s\n", result.Endpoint, result.Err)
			continue
		}
		t.AppendRow(table.Row{result.Endpoint, result.NotBefore, result.NotAfter, result.DaysLeft, result.Status})
	}

	fmt.Println(t.Render())
}

func configureCertExpiryReportTable() table.Writer {

	t := table.NewWriter()
	t.SetTitle("SSL Check: Who is Expiring Soon?")
	t.SetAutoIndex(true)
	t.Style().Format.Header = text.FormatTitle
	t.AppendHeader(table.Row{"Domain", "Valid From", "Expires In", "Days Left", "Status"})

	t.SortBy([]table.SortBy{
		{Name: "Expires In", Mode: table.Asc},
	})

	return t
}

func certStatus(daysLeft int) string {
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
