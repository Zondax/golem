package zrouter

import (
	"encoding/json"
	"errors"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
)

type HandlerFunc func(ctx Context) (domain.ServiceResponse, error)

func NotFoundHandler(_ Context) (domain.ServiceResponse, error) {
	msg := "Route not found"
	return domain.NewErrorNotFound(msg), nil
}

func ToHandlerFunc(h http.Handler) HandlerFunc {
	return func(ctx Context) (domain.ServiceResponse, error) {
		chiCtx, ok := ctx.(*chiContextAdapter)
		if !ok {
			return nil, errors.New("context provided is not a *chiContextAdapter")
		}

		h.ServeHTTP(chiCtx.ctx, chiCtx.req)
		return nil, nil
	}
}

func getChiHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adaptedContext := &chiContextAdapter{ctx: w, req: r}

		serviceResponse, err := handler(adaptedContext)
		if err != nil {
			handleError(w, err)
			return
		}

		handleServiceResponse(w, serviceResponse)
	}
}

func handleError(w http.ResponseWriter, err error) {
	var apiErr *domain.APIError

	if errors.As(err, &apiErr) {
		writeAPIErrorResponse(w, apiErr)
		return
	}

	writeInternalServerError(w)
}

func handleServiceResponse(w http.ResponseWriter, serviceResponse domain.ServiceResponse) {
	if serviceResponse == nil {
		return
	}

	body, err := serviceResponse.ResponseBytes()
	if err != nil {
		http.Error(w, "Failed to process response.", http.StatusInternalServerError)
		return
	}

	contentType := serviceResponse.Header().Get(domain.ContentTypeHeader)
	w.Header().Set(domain.ContentTypeHeader, contentType)
	w.WriteHeader(serviceResponse.Status())
	_, _ = w.Write(body)
}

func writeAPIErrorResponse(w http.ResponseWriter, apiErr *domain.APIError) {
	w.Header().Set(domain.ContentTypeHeader, domain.ContentTypeJSON)
	w.WriteHeader(apiErr.HTTPStatus)
	responseBody, _ := json.Marshal(apiErr)
	_, _ = w.Write(responseBody)
}

func writeInternalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
