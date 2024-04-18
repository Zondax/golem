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

func TestGetRoutePatternIncludingSubrouters(t *testing.T) {
	r := chi.NewRouter()
	subRouter := chi.NewRouter()

	// Configure a test route on the subrouter
	subRoutePattern := "/sub/{subParam}"
	subRouter.Get(subRoutePattern, func(w http.ResponseWriter, r *http.Request) {
		routePattern := GetRoutePattern(r)
		assert.Equal(t, "/test/sub/{subParam}", routePattern, "The returned route pattern should match the subrouter pattern.")
	})

	// Mount the subrouter onto a specific path of the main router
	r.Mount("/test", subRouter)

	// Test request for the subrouter route
	reqSub := httptest.NewRequest("GET", "/test/sub/456", nil)
	wSub := httptest.NewRecorder()
	r.ServeHTTP(wSub, reqSub)
	assert.Equal(t, http.StatusOK, wSub.Code, "The expected status code for subrouter should be 200 OK.")

	// Configure a test route on the main router
	mainRoutePattern := "/main/{mainParam}"
	r.Get(mainRoutePattern, func(w http.ResponseWriter, r *http.Request) {
		routePattern := GetRoutePattern(r)
		assert.Equal(t, "/main/{mainParam}", routePattern, "The returned route pattern should match the main router pattern.")
	})

	// Test request for the main router route
	reqMain := httptest.NewRequest("GET", "/main/123", nil)
	wMain := httptest.NewRecorder()
	r.ServeHTTP(wMain, reqMain)
	assert.Equal(t, http.StatusOK, wMain.Code, "The expected status code for main router should be 200 OK.")

	// Test request for the subrouter route when the path is undefined
	reqSub = httptest.NewRequest("GET", "/test/undefinedRoute", nil)
	wSub = httptest.NewRecorder()
	r.ServeHTTP(wSub, reqSub)
	assert.Equal(t, http.StatusNotFound, wSub.Code, "The expected status code for an undefined route should be 404 Not Found.")
	assert.Equal(t, notDefinedPath, GetRoutePattern(reqSub))
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
