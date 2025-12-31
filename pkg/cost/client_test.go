package cost
package cost

import (
	"context"
	"testing"
	"time"
)

func TestEstimateMonthlyCost(t *testing.T) {
	tests := []struct {
		name       string
		hourlyCost float64
		want       float64
	}{
		{
			name:       "basic calculation",
			hourlyCost: 1.0,
			want:       730.0,
		},
		{
			name:       "fractional cost",
			hourlyCost: 2.5,
			want:       1825.0,
		},
		{
			name:       "zero cost",
			hourlyCost: 0.0,
			want:       0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateMonthlyCost(tt.hourlyCost)
			if got != tt.want {
				t.Errorf("EstimateMonthlyCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "48 hours",
			duration: 48 * time.Hour,
			want:     "2d",
		},
		{
			name:     "24 hours",
			duration: 24 * time.Hour,
			want:     "1d",
		},
		{
			name:     "36 hours",
			duration: 36 * time.Hour,
			want:     "36h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	endpoint := "http://test:9003"
	timeout := 30 * time.Second

	client := NewClient(endpoint, timeout)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.endpoint != endpoint {
		t.Errorf("endpoint = %v, want %v", client.endpoint, endpoint)
	}

	if client.httpClient.Timeout != timeout {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, timeout)
	}
}

func TestClient_HealthCheck(t *testing.T) {
	// This test would require a mock HTTP server
	// Placeholder for now
	t.Skip("Requires mock HTTP server")
}
