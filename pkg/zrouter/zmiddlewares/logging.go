package zmiddlewares

import (
	"bytes"
	"github.com/zondax/golem/pkg/logger"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := &bytes.Buffer{}

		rw := &responseWriter{
			ResponseWriter: w,
			body:           buffer,
		}

		start := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		requestID := r.Header.Get(RequestIDHeader)
		ctx := r.Context()

		logger.Log().Debugf(ctx, "request_id: %s - Method: %s - URL: %s | Status: %d - Duration: %s - Response Body: %s",
			requestID, r.Method, r.URL.String(), rw.status, duration, rw.Body())
	})
}
