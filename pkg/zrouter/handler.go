package zrouter

import (
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
)

type HandlerFunc func(ctx Context) (domain.ServiceResponse, error)

func NotFoundHandler(_ Context) (domain.ServiceResponse, error) {
	msg := "Route not found"
	return domain.NewErrorNotFound(msg), nil
}

func getChiHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adaptedContext := &chiContextAdapter{ctx: w, req: r}

		serviceResponse, err := handler(adaptedContext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if serviceResponse != nil {
			body, err := serviceResponse.ResponseBytes()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			contentType := serviceResponse.Header().Get(domain.ContentTypeHeader)
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(serviceResponse.Status())
			_, _ = w.Write(body)
		}
	}
}
