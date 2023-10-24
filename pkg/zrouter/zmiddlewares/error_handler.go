package zmiddlewares

import (
	"encoding/json"
	"fmt"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Error("Internal error: %v\n%s", err, debug.Stack())
				message := fmt.Sprintf("An internal error occurred: %v", err)
				apiError := domain.NewAPIErrorResponse(http.StatusInternalServerError, "internal_error", message)
				w.WriteHeader(apiError.HTTPStatus)
				_ = json.NewEncoder(w).Encode(apiError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
