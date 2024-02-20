package zmiddlewares

import (
	"bytes"
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

type metricsResponseWriter struct {
	http.ResponseWriter
	status  int
	written int64
	body    *bytes.Buffer
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.status = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

func (mrw *metricsResponseWriter) Write(p []byte) (int, error) {
	if mrw.body == nil {
		mrw.body = new(bytes.Buffer)
	}
	mrw.body.Write(p)
	n, err := mrw.ResponseWriter.Write(p)
	mrw.written += int64(n)
	return n, err
}

func (mrw *metricsResponseWriter) Body() []byte {
	if mrw.body != nil {
		return mrw.body.Bytes()
	}
	return nil
}
