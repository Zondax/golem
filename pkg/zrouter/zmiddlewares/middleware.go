package zmiddlewares

import (
	"bytes"
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

type responseWriter struct {
	http.ResponseWriter
	status  int
	written int64
	body    *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	// Only copy the payload when a middleware asked for it (HTTP cache, debug logs).
	if rw.body != nil {
		rw.body.Write(p)
	}
	n, err := rw.ResponseWriter.Write(p)
	rw.written += int64(n)
	return n, err
}

func (rw *responseWriter) Body() []byte {
	if rw.body != nil {
		return rw.body.Bytes()
	}
	return nil
}
