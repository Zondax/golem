package zmiddlewares

import (
	"encoding/json"
	"fmt"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

const (
	internalErrorCode = "internal_error"
)

func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorf("Internal error: %v\n%s", err, debug.Stack())
				message := fmt.Sprintf("An internal error occurred: %v", err)
				apiError := domain.NewAPIErrorResponse(http.StatusInternalServerError, internalErrorCode, message)

				w.Header().Set(domain.ContentTypeHeader, domain.ContentTypeJSON)
				w.WriteHeader(apiError.HTTPStatus)
				_ = json.NewEncoder(w).Encode(apiError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
