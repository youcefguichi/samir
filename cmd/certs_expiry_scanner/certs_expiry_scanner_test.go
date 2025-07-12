package main


import (
    "crypto/x509"
    "testing"
    "time"
)

func TestGetCertificateStatus(t *testing.T) {
    tests := []struct {
        daysLeft int
        want     string
    }{
        {-1, "EXPIRED"},
        {0, "URGENT"},
        {29, "URGENT"},
        {30, "WARNING"},
        {89, "WARNING"},
        {90, "OKAY"},
        {365, "OKAY"},
    }
    for _, tt := range tests {
        got := GetCertificateStatus(tt.daysLeft)
        if got != tt.want {
            t.Errorf("GetCertificateStatus(%d) = %q, want %q", tt.daysLeft, got, tt.want)
        }
    }
}

func TestExpiresIn(t *testing.T) {
    now := time.Now()
    cert := &x509.Certificate{
        NotAfter: now.Add(48 * time.Hour),
    }

	// TODO: this case is failling, iy is less by one day always
    days := ExpiresIn(cert)
    if days != 2 {
        t.Errorf("ExpiresIn: got %d, want 2", days)
    }
    
    cert.NotAfter = now.Add(-24 * time.Hour)
    days = ExpiresIn(cert)
    if days != -1 {
        t.Errorf("ExpiresIn: got %d, want -1", days)
    }
}
