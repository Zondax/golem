package zmiddlewares

import (
	"github.com/go-chi/chi/v5/middleware"
)

func RequestID() Middleware {
	return middleware.RequestID
}
