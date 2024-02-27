package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
)

const (
	cacheKeyPrefix    = "zrouter_cache"
	cacheSetsMetric   = "cache_sets"
	cacheHitsMetric   = "cache_hits"
	cacheMissesMetric = "cache_misses"
)

type CacheProcessedPath struct {
	Regex *regexp.Regexp
	TTL   time.Duration
}

func CacheMiddleware(metricServer metrics.TaskMetrics, cache zcache.ZCache, config domain.CacheConfig) func(next http.Handler) http.Handler {
	processedPaths := processCachePaths(config.Paths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			fullURL := constructFullURL(r)

			for _, pPath := range processedPaths {
				if pPath.Regex.MatchString(path) {
					key := constructCacheKey(fullURL)

					if tryServeFromCache(w, r, cache, key, metricServer) {
						return
					}

					rw := &responseWriter{ResponseWriter: w}
					next.ServeHTTP(rw, r) // Important: this line needs to be BEFORE setting the cache.
					cacheResponseIfNeeded(rw, r, cache, key, pPath.TTL, metricServer)
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

func tryServeFromCache(w http.ResponseWriter, r *http.Request, cache zcache.ZCache, key string, metricServer metrics.TaskMetrics) bool {
	var cachedResponse []byte
	err := cache.Get(r.Context(), key, &cachedResponse)
	if err == nil && cachedResponse != nil {
		w.Header().Set(domain.ContentTypeHeader, domain.ContentTypeApplicationJSON)
		_, _ = w.Write(cachedResponse)

		if err = metricServer.IncrementMetric(cacheHitsMetric, r.URL.Path); err != nil {
			logger.GetLoggerFromContext(r.Context()).Errorf("Error incrementing cache_hits metric: %v", err)
		}

		return true
	}

	if err = metricServer.IncrementMetric(cacheMissesMetric, r.URL.Path); err != nil {
		logger.GetLoggerFromContext(r.Context()).Errorf("Error incrementing cache_misses metric: %v", err)
	}

	return false
}

func cacheResponseIfNeeded(rw *responseWriter, r *http.Request, cache zcache.ZCache, key string, ttl time.Duration, metricServer metrics.TaskMetrics) {
	if rw.status != http.StatusOK {
		return
	}

	responseBody := rw.Body()
	if err := cache.Set(r.Context(), key, responseBody, ttl); err != nil {
		logger.GetLoggerFromContext(r.Context()).Errorf("Internal error when setting cache response: %v\n%s", err, debug.Stack())
		return
	}

	if err := metricServer.IncrementMetric(cacheSetsMetric, r.URL.Path); err != nil {
		logger.GetLoggerFromContext(r.Context()).Errorf("Error incrementing cache_sets metric: %v", err)
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
