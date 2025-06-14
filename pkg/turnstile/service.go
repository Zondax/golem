package turnstile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Constants for Turnstile API fields and headers
const (
	// Form field names for Turnstile API
	FieldSecret   = "secret"
	FieldResponse = "response"

	// HTTP headers
	HeaderContentType = "Content-Type"

	// Default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second
)

// Service defines the interface for Turnstile verification
type Service interface {
	// Verify validates a Turnstile token against the configured endpoint
	Verify(ctx context.Context, token string) error
}

// Config holds the required configuration for the Turnstile service
type Config struct {
	SecretKey string
	Endpoint  string
	// HTTPClient allows dependency injection of HTTP client
	// If nil, a default client with reasonable timeouts will be used
	HTTPClient *http.Client
	// Timeout for HTTP requests (default: 30 seconds)
	Timeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Timeout: DefaultTimeout,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

type service struct {
	config     Config
	httpClient *http.Client
}

// NewService creates a new instance of the Turnstile verification service
func NewService(config Config) Service {
	// Apply defaults if not provided
	if config.HTTPClient == nil {
		if config.Timeout == 0 {
			config.Timeout = DefaultTimeout
		}
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &service{
		config:     config,
		httpClient: config.HTTPClient,
	}
}

// verifyResponse represents the response structure from the Turnstile API
type verifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify implements the Service interface by validating a Turnstile token.
// It sends a multipart form request to the configured endpoint with the secret key
// and token, then processes the response to determine if the verification was successful.
func (s *service) Verify(ctx context.Context, token string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField(FieldSecret, s.config.SecretKey); err != nil {
		return fmt.Errorf("failed to write secret key: %w", err)
	}

	if err := writer.WriteField(FieldResponse, token); err != nil {
		return fmt.Errorf("failed to write response token: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.Endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(HeaderContentType, writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var result verifyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("turnstile verification failed: %v", result.ErrorCodes)
	}

	return nil
}
