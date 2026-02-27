package sradmin

import (
	"crypto/tls"
	"ktea/config"
	"net/http"
	"testing"
)

func TestCreateHttpClient_TLSConfig(t *testing.T) {
	tests := []struct {
		name     string
		registry *config.SchemaRegistryConfig
		validate func(*testing.T, *tls.Config)
	}{
		{
			name: "skip verify enabled",
			registry: &config.SchemaRegistryConfig{
				Url:      "https://localhost:8081",
				Username: "user",
				Password: "pass",
				TLSConfig: config.TLSConfig{
					SkipVerify: true,
				},
			},
			validate: func(t *testing.T, cfg *tls.Config) {
				if !cfg.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be true")
				}
			},
		},
		{
			name: "skip verify disabled",
			registry: &config.SchemaRegistryConfig{
				Url:      "https://localhost:8081",
				Username: "user",
				Password: "pass",
				TLSConfig: config.TLSConfig{
					SkipVerify: false,
				},
			},
			validate: func(t *testing.T, cfg *tls.Config) {
				if cfg.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createHttpClient(tt.registry)
			transport := client.Transport.(roundTripperWithAuth)
			httpTransport := transport.baseTransport.(*http.Transport)
			tt.validate(t, httpTransport.TLSClientConfig)
		})
	}
}
