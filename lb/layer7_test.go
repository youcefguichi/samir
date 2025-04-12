package main

import (
	//"fmt"
	"reflect"
	"testing"
)

func TestGetNextBackend(t *testing.T) {

	tests := []struct {
		name              string
		requestNumber     int
		availableBackends   []string
		trafficServingBackends []string
	}{
		{
			name:              "number of requests: 1",
			requestNumber:     1,
			availableBackends:   []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1"},
		},
		{
			name:              "number of requests: 2",
			requestNumber:     2,
			availableBackends:   []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2"},
		},
		{
			name:              "number of requests: 3",
			requestNumber:     3,
			availableBackends:   []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2", "backend1"},
		},
		{
			name:              "number of requests: 4",
			requestNumber:     4,
			availableBackends:   []string{"backend1", "backend2"},
			trafficServingBackends: []string{"backend1", "backend2", "backend1", "backend2"},
		},
		{
			name:              "number of requests: 4",
			requestNumber:     4,
			availableBackends:   []string{"backend1", "backend2", "backend3"},
			trafficServingBackends: []string{"backend1", "backend2", "backend3", "backend1"},
		},
	}

	for _, test := range tests {
		lb := NewLoadBalancer("localhost", "5556", "https", "server.crt", "server.key", test.availableBackends)
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
