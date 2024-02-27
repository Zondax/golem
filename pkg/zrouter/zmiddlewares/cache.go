package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
)

const (
	cacheKeyPrefix = "zrouter_cache"
)

type CacheProcessedPath struct {
	Regex *regexp.Regexp
	TTL   time.Duration
}

func CacheMiddleware(cache zcache.ZCache, config domain.CacheConfig) func(next http.Handler) http.Handler {
	processedPaths := processCachePaths(config.Paths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			fullURL := constructFullURL(r)

			for _, pPath := range processedPaths {
				if pPath.Regex.MatchString(path) {
					key := constructCacheKey(fullURL)

					if tryServeFromCache(w, r, cache, key) {
						return
					}

					rw := &responseWriter{ResponseWriter: w}
					next.ServeHTTP(rw, r) // Important: this line needs to be BEFORE setting the cache.
					cacheResponseIfNeeded(rw, r, cache, key, pPath.TTL)
					return
				}
			}

			next.ServeHTTP(w, r)
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

		logger.GetLoggerFromContext(r.Context()).Debugf("[Cache] Method: %s - URL: %s | Status: %d - Response Body: %s",
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
		logger.GetLoggerFromContext(r.Context()).Errorf("Internal error when setting cache response: %v\n%s", err, debug.Stack())
	}
}

func ParseCacheConfigPaths(paths map[string]string) (domain.CacheConfig, error) {
	parsedPaths := make(map[string]time.Duration)

	for path, ttlStr := range paths {
		ttl, err := time.ParseDuration(ttlStr)
		if err != nil {
			return domain.CacheConfig{}, fmt.Errorf("error parsing TTL duration for the path %s: %w", path, err)
		}
		parsedPaths[path] = ttl
	}

	return domain.CacheConfig{Paths: parsedPaths}, nil
}

func processCachePaths(paths map[string]time.Duration) []CacheProcessedPath {
	var processedPaths []CacheProcessedPath
	for path, ttl := range paths {
		processedPaths = append(processedPaths, CacheProcessedPath{
			Regex: PathToRegexp(path),
			TTL:   ttl,
		})
	}
	return processedPaths
}
