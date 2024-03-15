package zmiddlewares

import (
	"bytes"
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zcache"
)

func TestCacheMiddleware(t *testing.T) {
	expectedCacheKey := "GET.zrouter_cache:/cached-path"
	r := chi.NewRouter()
	mockCache := new(zcache.MockZCache)
	logger.InitLogger(logger.Config{})
	cacheConfig := domain.CacheConfig{Paths: map[string]time.Duration{
		"/cached-path": 5 * time.Minute,
	}}

	r.Use(CacheMiddleware(metrics.NewTaskMetrics("", "", "appName"), mockCache, cacheConfig))

	// Simulate a response that should be cached
	r.Get("/cached-path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Test!"))
	})

	cachedResponseBody := []byte("Test!")

	// Setup the mock for the first request (cache miss)
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Once()
	mockCache.On("Set", mock.Anything, expectedCacheKey, cachedResponseBody, 5*time.Minute).Return(nil).Once()

	// Setup the mock for the second request (cache hit)
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]byte) // Get the argument where the cached response will be stored
		*arg = cachedResponseBody    // Simulate the cached response
	})

	// Perform the first request: the response should be generated and cached
	req := httptest.NewRequest("GET", "/cached-path", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Test!", rec.Body.String())

	// Verify that the cache mock was invoked correctly
	mockCache.AssertExpectations(t)

	// Perform the second request: the response should be served from the cache
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "Test!", rec2.Body.String())

	// Verify that the cache mock was invoked correctly for the second request
	mockCache.AssertExpectations(t)
}

func TestCacheMiddlewareWithRequestBody(t *testing.T) {
	r := chi.NewRouter()
	mockCache := new(zcache.MockZCache)
	mockMetrics := metrics.NewTaskMetrics("", "", "appname")
	logger.InitLogger(logger.Config{})
	cacheConfig := domain.CacheConfig{
		Paths: map[string]time.Duration{
			"/post-path": 5 * time.Minute,
		},
	}

	r.Use(CacheMiddleware(mockMetrics, mockCache, cacheConfig))

	r.Post("/post-path", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		response := "Received: " + string(bodyBytes)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	})

	requestBody := []byte("Request Body Content")
	hashedBody := generateBodyHash(requestBody)
	expectedCacheKey := fmt.Sprintf("POST.zrouter_cache:/post-path.body:%s", hashedBody)

	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Once()
	mockCache.On("Set", mock.Anything, expectedCacheKey, mock.Anything, 5*time.Minute).Return(nil).Once()

	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]byte)
		*arg = []byte("Received: Request Body Content")
	}).Once()

	req := httptest.NewRequest("POST", "/post-path", bytes.NewBuffer(requestBody))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Received: Request Body Content", rec.Body.String())
	req.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "Received: Request Body Content", rec2.Body.String())
	mockCache.AssertExpectations(t)
}
