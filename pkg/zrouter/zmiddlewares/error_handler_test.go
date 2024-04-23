package zmiddlewares

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	logger.InitLogger(logger.Config{})
	r := chi.NewRouter()
	r.Use(ErrorHandlerMiddleware())

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("Some unexpected error")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var apiError domain.APIError
	err := json.NewDecoder(rec.Body).Decode(&apiError)
	assert.NoError(t, err)
	assert.Equal(t, "internal_error", apiError.ErrorCode)
	assert.Contains(t, apiError.Message, "Some unexpected error")
}
