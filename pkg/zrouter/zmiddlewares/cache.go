package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"time"
)

const (
	cacheKeyPrefix = "zrouter_cache"
)

func CacheMiddleware(cache zcache.ZCache, config domain.CacheConfig, logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			fullURL := constructFullURL(r)

			if ttl, found := config.Paths[path]; found {
				key := constructCacheKey(fullURL)

				if tryServeFromCache(w, r, cache, key) {
					return
				}

				mrw := &metricsResponseWriter{ResponseWriter: w}
				next.ServeHTTP(mrw, r) // Important: This line needs to be BEFORE setting the cache.
				cacheResponseIfNeeded(mrw, r, cache, key, ttl, logger)
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
		return true
	}
	return false
}

func cacheResponseIfNeeded(mrw *metricsResponseWriter, r *http.Request, cache zcache.ZCache, key string, ttl time.Duration, logger *zap.SugaredLogger) {
	if mrw.status != http.StatusOK {
		return
	}

	responseBody := mrw.Body()
	if err := cache.Set(r.Context(), key, responseBody, ttl); err != nil {
		logger.Errorf("Internal error when setting cache response: %v\n%s", err, debug.Stack())
	}
}
