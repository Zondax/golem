package zmiddlewares

import (
	"bytes"
	"github.com/zondax/golem/pkg/logger"
	"net/http"
	"time"
)

type LoggingMiddlewareOptions struct {
	ExcludePaths []string
}

func LoggingMiddleware(options LoggingMiddlewareOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range options.ExcludePaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			buffer := &bytes.Buffer{}

			rw := &responseWriter{
				ResponseWriter: w,
				body:           buffer,
			}

			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)
			ctx := r.Context()

			log := logger.GetLoggerFromContext(ctx)

			if log.IsDebugEnabled() {
				log.Debugf("Method: %s - URL: %s | Status: %d - Duration: %s - Response Body: %s",
					r.Method, r.URL.String(), rw.status, duration, string(rw.Body()))
				return
			}

			log.Infof("Method: %s - URL: %s | Status: %d - Duration: %s",
				r.Method, r.URL.String(), rw.status, duration)
		})
	}
}
