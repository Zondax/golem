package zmiddlewares

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimit(t *testing.T) {
	r := chi.NewRouter()

	r.Use(RateLimit(1))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}
