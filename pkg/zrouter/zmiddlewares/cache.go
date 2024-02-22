package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"runtime/debug"
	"time"
)

const (
	cacheKeyPrefix = "zrouter_cache"
)

func CacheMiddleware(cache zcache.ZCache, config domain.CacheConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			fullURL := constructFullURL(r)

			if ttl, found := config.Paths[path]; found {
				key := constructCacheKey(fullURL)

				if tryServeFromCache(w, r, cache, key) {
					return
				}

				rw := &responseWriter{ResponseWriter: w}
				next.ServeHTTP(rw, r) // Important: This line needs to be BEFORE setting the cache.
				cacheResponseIfNeeded(rw, r, cache, key, ttl)
			}
		})
	}
}

func constructFullURL(r *http.Request) string {
	fullURL := r.URL.Path
	if queryString := r.URL.RawQuery; queryString != "" {
		fullURL += "?" + queryString
	}
	return fullURL
}

func constructCacheKey(fullURL string) string {
	return fmt.Sprintf("%s:%s", cacheKeyPrefix, fullURL)
}

func tryServeFromCache(w http.ResponseWriter, r *http.Request, cache zcache.ZCache, key string) bool {
	var cachedResponse []byte
	err := cache.Get(r.Context(), key, &cachedResponse)
	if err == nil && cachedResponse != nil {
		w.Header().Set(domain.ContentTypeHeader, domain.ContentTypeApplicationJSON)
		_, _ = w.Write(cachedResponse)
		requestID := r.Header.Get(RequestIDHeader)

		logger.Log().Debugf(r.Context(), "[Cache] request_id: %s - Method: %s - URL: %s | Status: %d - Response Body: %s",
			requestID, r.Method, r.URL.String(), http.StatusOK, string(cachedResponse))
		return true
	}
	return false
}

func cacheResponseIfNeeded(rw *responseWriter, r *http.Request, cache zcache.ZCache, key string, ttl time.Duration) {
	if rw.status != http.StatusOK {
		return
	}

	responseBody := rw.Body()
	if err := cache.Set(r.Context(), key, responseBody, ttl); err != nil {
		logger.Log().Errorf(r.Context(), "Internal error when setting cache response: %v\n%s", err, debug.Stack())
	}
}
