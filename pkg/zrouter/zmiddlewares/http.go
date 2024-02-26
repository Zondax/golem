package zmiddlewares

import (
	"github.com/google/uuid"
	"github.com/zondax/golem/pkg/logger"
	"net/http"
)

const (
	RequestIDHeader = "X-Request-ID"
)

func RequestID() Middleware {
	return requestIDMiddleware
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set(RequestIDHeader, requestID)
		}

		w.Header().Set(RequestIDHeader, requestID)
		rw := &responseWriter{ResponseWriter: w}
		ctx := logger.ContextWithLogger(r.Context(), logger.NewLogger(logger.Field{
			Key:   logger.RequestIDKey,
			Value: requestID,
		}))

		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func Logger(options LoggingMiddlewareOptions) Middleware {
	return LoggingMiddleware(options)
}
