package zmiddlewares

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/zrouter/domain"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zcache"
)

func setupTest() (context.Context, metrics.TaskMetrics) {
	// Initialize logger
	logger.InitLogger(logger.Config{
		Level: "debug",
	})
	ctx := context.Background()

	// Initialize metrics
	ms := metrics.NewTaskMetrics("test", "test", "test_app")
	return ctx, ms
}

func TestCacheMiddleware(t *testing.T) {
	ctx, ms := setupTest()

	expectedCacheKey := "zrouter_cache.GET:/api/cached-path"
	r := chi.NewRouter()
	mockCache := new(zcache.MockZCache)

	cacheConfig := domain.CacheConfig{
		Paths: map[string]time.Duration{
			"/api/cached-path": 5 * time.Minute,
		},
	}

	errs := RegisterRequestMetrics(ms)
	assert.Empty(t, errs)

	r.Use(CacheMiddleware(ms, mockCache, cacheConfig))

	cachedResponseBody := []byte("Test!")

	// Setup route handler
	r.Get("/api/cached-path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(cachedResponseBody)
	})

	// Setup mock expectations
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Once()
	mockCache.On("Set", mock.Anything, expectedCacheKey, cachedResponseBody, 5*time.Minute).Return(nil).Once()
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]byte)
		*arg = cachedResponseBody
	}).Once()

	// First request (cache miss)
	req := httptest.NewRequest("GET", "/api/cached-path", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(cachedResponseBody), rec.Body.String())

	// Second request (cache hit)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, string(cachedResponseBody), rec2.Body.String())

	mockCache.AssertExpectations(t)
}

func TestCacheMiddlewareWithRequestBody(t *testing.T) {
	ctx, ms := setupTest()

	r := chi.NewRouter()
	mockCache := new(zcache.MockZCache)

	cacheConfig := domain.CacheConfig{
		Paths: map[string]time.Duration{
			"/api/post-path": 5 * time.Minute,
		},
	}

	r.Use(CacheMiddleware(ms, mockCache, cacheConfig))

	// Setup route handler
	r.Post("/api/post-path", func(w http.ResponseWriter, r *http.Request) {
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
	expectedCacheKey := fmt.Sprintf("zrouter_cache.POST:/api/post-path.body:%s", hashedBody)
	expectedResponse := []byte("Received: Request Body Content")

	// Setup mock expectations
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Once()
	mockCache.On("Set", mock.Anything, expectedCacheKey, expectedResponse, 5*time.Minute).Return(nil).Once()
	mockCache.On("Get", mock.Anything, expectedCacheKey, mock.AnythingOfType("*[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]byte)
		*arg = expectedResponse
	}).Once()

	// First request (cache miss)
	req := httptest.NewRequest("POST", "/api/post-path", bytes.NewBuffer(requestBody))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(expectedResponse), rec.Body.String())

	// Second request (cache hit)
	req = httptest.NewRequest("POST", "/api/post-path", bytes.NewBuffer(requestBody))
	req = req.WithContext(ctx)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, string(expectedResponse), rec2.Body.String())

	mockCache.AssertExpectations(t)
}
