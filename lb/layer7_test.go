package main

import (
	//"fmt"
	"reflect"
	"testing"
)

func TestGetNextBackend(t *testing.T) {

	tests := []struct {
		name                   string
		requestNumber          int
		availableBackends      []string
		trafficServingBackends []string
	}{
		{
			name:                   "number of requests: 1",
			requestNumber:          1,
			availableBackends:      []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1"},
		},
		{
			name:                   "number of requests: 2",
			requestNumber:          2,
			availableBackends:      []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2"},
		},
		{
			name:                   "number of requests: 3",
			requestNumber:          3,
			availableBackends:      []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2", "backend1"},
		},
		{
			name:                   "number of requests: 4",
			requestNumber:          4,
			availableBackends:      []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2", "backend1", "backend2"},
		},
		{
			name:                   "number of requests: 4",
			requestNumber:          4,
			availableBackends:      []string{"backend1", "backend2", "backend3"},
			trafficServingBackends: []string{"backend1", "backend2", "backend3", "backend1"},
		},
	}

	for _, test := range tests {
		lb := lb{}
		lb.config.BackendServers = test.availableBackends
		lb.currentBackend = 0
		var backends []string
		for i := 0; i < test.requestNumber; i++ {

			backend, err := lb.getNextBackend()

			if err != nil {
				t.Errorf("Error get ting next backend: %v", err)
			}
			backends = append(backends, backend)
		}

		if !reflect.DeepEqual(backends, test.trafficServingBackends) {
			t.Errorf("Expected %s, got %s", test.trafficServingBackends, backends)
		}

	}

}

func TestIsIPAllowed(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		allowed  []string
		expected bool
	}{
		{
			name:     "one CIDR",
			ip:       "10.10.1.1",
			allowed:  []string{"10.10.1.0/24"},
			expected: true,
		},
		{
			name:     "match with single IP",
			ip:       "10.10.2.1",
			allowed:  []string{"10.10.2.1"},
			expected: true,
		},
		{
			name:     "match with single ip written in CIDR format",
			ip:       "10.10.1.1",
			allowed:  []string{"10.10.1.1/32"},
			expected: true,
		},
		{
			name:     "deny ip not in CIDR",
			ip:       "10.10.2.1",
			allowed:  []string{"10.10.1.0/24"},
			expected: false,
		},
		{
			name:     "deny ip not in CIDR",
			ip:       "10.10.2.1",
			allowed:  []string{"10.10.1.1/24", "10.10.3.0/24"},
			expected: false,
		},
		{
			name:     "allow ip multiple CIDR",
			ip:       "10.10.2.1",
			allowed:  []string{"10.10.1.1/24", "10.10.2.0/24"},
			expected: true,
		},
	}

	for _, test := range tests {
		result := isIPAllowed(test.ip, test.allowed)
		if result != test.expected {
			t.Errorf("Test %s failed: expected %v, got %v", test.name, test.expected, result)
		}
	}

}
