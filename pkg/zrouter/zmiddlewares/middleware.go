package zmiddlewares

import (
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

type metricsResponseWriter struct {
	http.ResponseWriter
	status  int
	written int64
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.status = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

func (mrw *metricsResponseWriter) Write(p []byte) (int, error) {
	n, err := mrw.ResponseWriter.Write(p)
	mrw.written += int64(n)
	return n, err
}
