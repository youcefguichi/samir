package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"time"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	//"math"
)

func RetreiveCertificateFromPeer(endpoint string, insecureSkip bool) ([]*x509.Certificate, error) {

	con, err := tls.Dial("tcp", endpoint, &tls.Config{InsecureSkipVerify: insecureSkip})
	if err != nil {
		return nil, errors.New("failed to connect to endpoint: " + endpoint + " " + err.Error())
	}

	certs := con.ConnectionState().PeerCertificates

	return certs, nil

}

func expiresIn(cert *x509.Certificate) int {

	daysLeft := time.Until(cert.NotAfter).Hours() / 24
	return int(daysLeft)
}

func CheckCertsforAllPeers(endpoints []string, insecureskip bool) {

	t := configureCertExpiryReportTable()

	for _, endpoint := range endpoints {

		certs, err := RetreiveCertificateFromPeer(endpoint, insecureskip)

		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, cert := range certs {
			t.AppendRow(table.Row{endpoint, cert.NotBefore, cert.NotAfter, expiresIn(cert), certStatus(expiresIn(cert))})
		}

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
