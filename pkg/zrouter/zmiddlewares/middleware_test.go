package zmiddlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zondax/golem/pkg/logger"
)

func TestResponseWriterDoesNotBufferUnlessAsked(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	n, err := rw.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", rec.Body.String())
	assert.Nil(t, rw.Body())
	assert.Equal(t, int64(5), rw.written)
}

func TestResponseWriterBuffersWhenBodySet(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, body: &bytes.Buffer{}}

	_, err := rw.Write([]byte("cached"))
	require.NoError(t, err)
	assert.Equal(t, "cached", rec.Body.String())
	assert.Equal(t, []byte("cached"), rw.Body())
}

func TestRequestIDDoesNotWrapResponseWriter(t *testing.T) {
	var saw http.ResponseWriter
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = w
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	RequestID()(next).ServeHTTP(rec, req)

	assert.Equal(t, rec, saw)
	assert.Equal(t, "ok", rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader))
}

func TestLoggingMiddlewareDoesNotCaptureBodyAtInfo(t *testing.T) {
	logger.InitLogger(logger.Config{Level: "info"})

	payload := bytes.Repeat([]byte("x"), 4096)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/big", nil)
	rec := httptest.NewRecorder()
	LoggingMiddleware(LoggingMiddlewareOptions{})(next).ServeHTTP(rec, req)

	assert.Equal(t, payload, rec.Body.Bytes())
}

func TestLoggingMiddlewareCapturesBodyAtDebug(t *testing.T) {
	logger.InitLogger(logger.Config{Level: "debug"})
	t.Cleanup(func() { logger.InitLogger(logger.Config{Level: "info"}) })

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("debug-body"))
	})

	req := httptest.NewRequest(http.MethodGet, "/dbg", nil)
	rec := httptest.NewRecorder()
	LoggingMiddleware(LoggingMiddlewareOptions{})(next).ServeHTTP(rec, req)

	assert.Equal(t, "debug-body", rec.Body.String())
}
