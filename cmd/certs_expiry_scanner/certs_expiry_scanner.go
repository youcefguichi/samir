package main

import (
	"crypto/tls"
	"crypto/x509"
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

type Response struct {
	Certificates []*x509.Certificate
	Err          error
}

func RetreiveCertificateFromPeer(ctx context.Context, endpoint string, insecureSkip bool) (*Response, error) {

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	resultCh := make(chan Response)

	go func() {
		con, err := tls.Dial("tcp", endpoint, &tls.Config{InsecureSkipVerify: insecureSkip})
		// if err != nil {
		// 	return nil, errors.New("failed to connect to endpoint: " + endpoint + " " + err.Error())
		// }

		certs := con.ConnectionState().PeerCertificates
		resultCh <- Response{
			Certificates: certs,
			Err:          err,
		}

	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-resultCh:
		return &resp, nil
	}

}

func expiresIn(cert *x509.Certificate) int {

	daysLeft := time.Until(cert.NotAfter).Hours() / 24
	return int(daysLeft)
}

func CheckCertsforAllPeers(endpoints []string, insecureskip bool) {

	var wg sync.WaitGroup
	resultCh := make(chan CertResult, len(endpoints))

	t := configureCertExpiryReportTable()

	for _, endpoint := range endpoints {

		wg.Add(1)

		go func(ep string) {
			defer wg.Done()

			ctx := context.Background()
			certs, err := RetreiveCertificateFromPeer(ctx, ep, insecureskip)

			if err != nil {
				resultCh <- CertResult{
					Endpoint: ep,
					Err:      err,
				}
				return
			}

			for _, cert := range certs.Certificates {

				daysLeft := expiresIn(cert)

				if certs.Err != nil {
					resultCh <- CertResult{
						Endpoint: ep,
						Err:      certs.Err,
					}
					return
				}

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
	t.SetTitle("🔐 SSL Check: Who is Expiring Soon?")
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
