package zmiddlewares

import (
	"github.com/google/uuid"
	"net/http"
)

const (
	RequestIDHeader = "X-Request-ID"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return requestIDMiddleware(next)
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set(RequestIDHeader, requestID)
		}

		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

func Logger() Middleware {
	return LoggingMiddleware
}
