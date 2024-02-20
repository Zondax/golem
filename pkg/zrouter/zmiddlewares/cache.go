package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

const (
	cacheKeyPrefix = "zrouter_cache"
)

func CacheMiddleware(cache zcache.ZCache, config domain.CacheConfig, logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if ttl, found := config.Paths[path]; found {
				key := fmt.Sprintf("%s:%s", cacheKeyPrefix, path)

				var cachedResponse []byte
				err := cache.Get(r.Context(), key, &cachedResponse)
				if err == nil && cachedResponse != nil {
					_, _ = w.Write(cachedResponse)
					return
				}

				mrw := &metricsResponseWriter{ResponseWriter: w}
				next.ServeHTTP(mrw, r)

				if mrw.status == http.StatusOK {
					responseBody := mrw.Body()
					if err = cache.Set(r.Context(), key, responseBody, ttl); err != nil {
						logger.Errorf("Internal error when setting cache response: %v\n%s", err, debug.Stack())
					}
				}
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
