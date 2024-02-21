package zmiddlewares

import (
	"bytes"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := &bytes.Buffer{}

		mrw := &responseWriter{
			ResponseWriter: w,
			body:           buffer,
		}

		start := time.Now()
		next.ServeHTTP(mrw, r)
		duration := time.Since(start)
		requestID := r.Header.Get(RequestIDHeader)

		zap.S().Debugf("request_id: %s - Method: %s - URL: %s | Status: %d - Duration: %s - Response Body: %s",
			requestID, r.Method, r.URL.String(), mrw.status, duration, mrw.Body())
	})
}
