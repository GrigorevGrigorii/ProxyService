package handlers

import (
	"fmt"
	"proxy-service/internal/config"
	"testing"
)

func TestAllowedToProxy(t *testing.T) {
	// config that allows only requests to "mock" service with "GET" method and "/mock" path
	var services = map[string]config.Service{
		"mock": {
			Name: "mock",
			Targets: []config.Target{
				{
					Path:   "/mock",
					Method: "GET",
				},
			},
		},
	}

	handlers := Handlers{
		Services: services,
	}

	tests := []struct {
		service string
		method  config.HTTPMethod
		path    string
		want    bool
	}{
		{
			service: "mock",
			method:  config.MethodGet,
			path:    "/mock",
			want:    true,
		},
		{
			service: "unknown", // not allowed by config
			method:  config.MethodGet,
			path:    "/mock",
			want:    false,
		},
		{
			service: "mock",
			method:  config.MethodPost, // not allowed by config
			path:    "/mock",
			want:    false,
		},
		{
			service: "mock",
			method:  config.MethodGet,
			path:    "/another/path", // not allowed by config
			want:    false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := handlers.allowedToProxy(tt.service, tt.method, tt.path)
			if got != tt.want {
				t.Errorf("For service=%v, method=%v, path=%v got %t, but expected %t", tt.service, tt.method, tt.path, got, tt.want)
			}
		})
	}
}
