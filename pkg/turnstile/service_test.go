package turnstile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func createTestConfig(endpoint string) Config {
	return Config{
		SecretKey: "test-secret",
		Endpoint:  endpoint,
	}
}

func createSuccessResponse() verifyResponse {
	return verifyResponse{
		Success: true,
	}
}

func createFailureResponse(errorCodes ...string) verifyResponse {
	return verifyResponse{
		Success:    false,
		ErrorCodes: errorCodes,
	}
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, response verifyResponse) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	require.NoError(t, err)
}

func TestConstants(t *testing.T) {
	t.Run("WhenFieldConstants_ShouldHaveCorrectValues", func(t *testing.T) {
		assert.Equal(t, "secret", FieldSecret)
		assert.Equal(t, "response", FieldResponse)
	})

	t.Run("WhenHeaderConstants_ShouldHaveCorrectValues", func(t *testing.T) {
		assert.Equal(t, "Content-Type", HeaderContentType)
	})

	t.Run("WhenDefaultTimeout_ShouldBe30Seconds", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, DefaultTimeout)
	})
}

func TestNewService(t *testing.T) {
	t.Run("WhenValidConfig_ShouldCreateService", func(t *testing.T) {
		config := Config{
			SecretKey: "test-secret",
			Endpoint:  "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		}

		svc := NewService(config)
		assert.NotNil(t, svc)
	})

	t.Run("WhenNoHTTPClient_ShouldUseDefault", func(t *testing.T) {
		config := Config{
			SecretKey: "test-secret",
			Endpoint:  "https://example.com",
		}

		svc := NewService(config).(*service)
		assert.NotNil(t, svc.httpClient)
		assert.Equal(t, DefaultTimeout, svc.config.Timeout)
	})

	t.Run("WhenCustomTimeout_ShouldUseCustomTimeout", func(t *testing.T) {
		config := Config{
			SecretKey: "test-secret",
			Endpoint:  "https://example.com",
			Timeout:   10 * time.Second,
		}

		svc := NewService(config).(*service)
		assert.NotNil(t, svc.httpClient)
		assert.Equal(t, 10*time.Second, svc.config.Timeout)
	})

	t.Run("WhenStandardHTTPClient_ShouldAcceptIt", func(t *testing.T) {
		// Test that we can use the standard Go http.Client directly
		standardClient := &http.Client{
			Timeout: 15 * time.Second,
		}

		config := Config{
			SecretKey:  "test-secret",
			Endpoint:   "https://example.com",
			HTTPClient: standardClient,
		}

		svc := NewService(config).(*service)
		assert.Equal(t, standardClient, svc.httpClient)
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Run("WhenCalled_ShouldReturnValidDefaults", func(t *testing.T) {
		config := DefaultConfig()

		assert.Equal(t, DefaultTimeout, config.Timeout)
		assert.NotNil(t, config.HTTPClient)
		assert.IsType(t, &http.Client{}, config.HTTPClient)
	})
}

func TestService_Verify(t *testing.T) {
	t.Run("WhenValidToken_ShouldSucceed", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.Header.Get(HeaderContentType), "multipart/form-data")

			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)
			assert.Equal(t, "test-secret", r.FormValue(FieldSecret))
			assert.Equal(t, "valid-token", r.FormValue(FieldResponse))

			writeJSONResponse(t, w, createSuccessResponse())
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.NoError(t, err)
	})

	t.Run("WhenInvalidToken_ShouldReturnError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, createFailureResponse("invalid-input-response", "timeout-or-duplicate"))
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "invalid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "turnstile verification failed")
		assert.Contains(t, err.Error(), "invalid-input-response")
		assert.Contains(t, err.Error(), "timeout-or-duplicate")
	})

	t.Run("WhenServerError_ShouldReturnError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code: 500")
	})

	t.Run("WhenInvalidJSON_ShouldReturnError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("invalid json"))
			require.NoError(t, err)
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse response")
	})

	t.Run("WhenNetworkError_ShouldReturnError", func(t *testing.T) {
		// Arrange
		config := createTestConfig("http://non-existent-server.example.com")
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})

	t.Run("WhenContextCanceled_ShouldReturnError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			writeJSONResponse(t, w, createSuccessResponse())
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Act
		err := svc.Verify(ctx, "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("WhenInvalidEndpoint_ShouldReturnError", func(t *testing.T) {
		// Arrange
		config := createTestConfig("invalid-url")
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})

	t.Run("WhenEmptyToken_ShouldStillWork", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)
			assert.Equal(t, "", r.FormValue(FieldResponse))

			writeJSONResponse(t, w, createFailureResponse("missing-input-response"))
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "turnstile verification failed")
	})

	t.Run("WhenEmptySecretKey_ShouldStillWork", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)
			assert.Equal(t, "", r.FormValue(FieldSecret))

			writeJSONResponse(t, w, createFailureResponse("missing-input-secret"))
		})

		config := Config{
			SecretKey: "",
			Endpoint:  server.URL,
		}
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "turnstile verification failed")
	})

	t.Run("WhenResponseBodyReadFails_ShouldReturnError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("partial"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			if hijacker, ok := w.(http.Hijacker); ok {
				conn, _, _ := hijacker.Hijack()
				_ = conn.Close()
			}
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.True(t,
			err.Error() == "failed to read response body: unexpected EOF" ||
				err.Error() == "failed to parse response: unexpected end of JSON input" ||
				err.Error() == "failed to read response body: EOF" ||
				err.Error() == "failed to parse response: EOF",
			"Expected read or parse error, got: %s", err.Error())
	})

	t.Run("WhenMultipleErrorCodes_ShouldIncludeAllInError", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, createFailureResponse("invalid-input-response", "timeout-or-duplicate", "bad-request"))
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "invalid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "turnstile verification failed")
		assert.Contains(t, err.Error(), "invalid-input-response")
		assert.Contains(t, err.Error(), "timeout-or-duplicate")
		assert.Contains(t, err.Error(), "bad-request")
	})

	t.Run("WhenSuccessWithErrorCodes_ShouldSucceed", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			response := verifyResponse{
				Success:    true,
				ErrorCodes: []string{"some-warning"}, // Success=true should take precedence
			}
			writeJSONResponse(t, w, response)
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.NoError(t, err)
	})
}

func TestService_Verify_EdgeCases(t *testing.T) {
	t.Run("WhenMalformedEndpointURL_ShouldReturnError", func(t *testing.T) {
		// Arrange - Use a malformed URL that will cause http.NewRequestWithContext to fail
		config := Config{
			SecretKey: "test-secret",
			Endpoint:  "://invalid-url", // Malformed URL
		}
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("WhenVeryLongToken_ShouldStillWork", func(t *testing.T) {
		// Arrange - Test with a very long token to ensure multipart handling works
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)

			// Verify the long token was received correctly
			token := r.FormValue(FieldResponse)
			assert.Len(t, token, 10000) // Verify it's the expected length

			writeJSONResponse(t, w, createSuccessResponse())
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Create a very long token (10KB)
		longToken := string(make([]byte, 10000))
		for i := range longToken {
			longToken = longToken[:i] + "a" + longToken[i+1:]
		}

		// Act
		err := svc.Verify(context.Background(), longToken)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("WhenSpecialCharactersInToken_ShouldHandleCorrectly", func(t *testing.T) {
		// Arrange - Test with special characters that might cause multipart encoding issues
		specialToken := "token-with-special-chars-!@#$%^&*()_+-=[]{}|;:,.<>?"

		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)

			// Verify special characters were preserved
			receivedToken := r.FormValue(FieldResponse)
			assert.Equal(t, specialToken, receivedToken)

			writeJSONResponse(t, w, createSuccessResponse())
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), specialToken)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("WhenUnicodeCharactersInToken_ShouldHandleCorrectly", func(t *testing.T) {
		// Arrange - Test with Unicode characters
		unicodeToken := "token-with-unicode-🚀-测试-🔒"

		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)

			// Verify Unicode characters were preserved
			receivedToken := r.FormValue(FieldResponse)
			assert.Equal(t, unicodeToken, receivedToken)

			writeJSONResponse(t, w, createSuccessResponse())
		})

		config := createTestConfig(server.URL)
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), unicodeToken)

		// Assert
		assert.NoError(t, err)
	})
}

func TestService_Verify_WithCustomHTTPClient(t *testing.T) {
	t.Run("WhenCustomHTTPClient_ShouldUseIt", func(t *testing.T) {
		// Arrange
		server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, createSuccessResponse())
		})

		customClient := &http.Client{
			Timeout: 5 * time.Second,
		}

		config := Config{
			SecretKey:  "test-secret",
			Endpoint:   server.URL,
			HTTPClient: customClient,
		}
		svc := NewService(config)

		// Act
		err := svc.Verify(context.Background(), "valid-token")

		// Assert
		assert.NoError(t, err)
		// Verify the custom client was used by checking the service's internal state
		serviceCasted := svc.(*service)
		assert.Equal(t, customClient, serviceCasted.httpClient)
	})
}
