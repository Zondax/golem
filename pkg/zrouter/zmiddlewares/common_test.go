package zmiddlewares

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRoutePattern(t *testing.T) {
	r := chi.NewRouter()

	routePattern := "/test/{param}"
	r.Get(routePattern, func(w http.ResponseWriter, r *http.Request) {
		routePattern := GetRoutePattern(r)

		assert.Equal(t, routePattern, "/test/{param}", "The returned route pattern should match the one configured in the router.")
	})

	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "The expected status code should be 200 OK.")
}

func TestGetRequestBody(t *testing.T) {
	bodyContent := "test body content"
	req, err := http.NewRequest("POST", "/test", bytes.NewBufferString(bodyContent))
	assert.NoError(t, err)

	bodyBytes, err := getRequestBody(req)
	assert.NoError(t, err)
	assert.Equal(t, bodyContent, string(bodyBytes))

	// Verify the request body can be read again
	secondReadBytes, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, bodyContent, string(secondReadBytes))
}

func TestGenerateBodyHash(t *testing.T) {
	testContent := "test content"
	expectedHash := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	hash := generateBodyHash([]byte(testContent))
	assert.Equal(t, expectedHash, hash)
}
