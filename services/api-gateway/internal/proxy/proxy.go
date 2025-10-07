package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-gateway/internal/config"
)

// Proxy represents a reverse proxy for backend services.
type Proxy struct {
	client  *http.Client
	config  *config.ServiceConfig
	service string
}

// New creates a new proxy instance.
func New(service string, cfg *config.ServiceConfig) *Proxy {
	return &Proxy{
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		config:  cfg,
		service: service,
	}
}

// ForwardRequest forwards an HTTP request to the backend service.
func (p *Proxy) ForwardRequest(ctx context.Context, r *http.Request, path string) (*http.Response, error) {
	// Build target URL
	targetURL := p.config.URL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create new request
	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set service-specific headers
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("X-Gateway-Service", p.service)

	// Forward request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request to %s: %w", p.service, err)
	}

	return resp, nil
}

// ForwardResponse forwards the response back to the client.
func (p *Proxy) ForwardResponse(w http.ResponseWriter, resp *http.Response) error {
	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	if resp.Body != nil {
		defer resp.Body.Close()
		_, err := io.Copy(w, resp.Body)
		return err
	}

	return nil
}

// HealthCheck checks if the backend service is healthy.
func (p *Proxy) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.URL+"/healthz", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed for %s: %w", p.service, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed for %s: status %d", p.service, resp.StatusCode)
	}

	return nil
}
