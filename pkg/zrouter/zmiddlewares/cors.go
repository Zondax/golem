package zmiddlewares

import (
	"github.com/go-chi/cors"
	"net/http"
)

type CorsOptions struct {
	AllowedOrigins     []string
	AllowOriginFunc    func(r *http.Request, origin string) bool
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAge             int
	OptionsPassthrough bool
	Debug              bool
}

func (co CorsOptions) toChiOptions() cors.Options {
	return cors.Options{
		AllowedOrigins:     co.AllowedOrigins,
		AllowOriginFunc:    co.AllowOriginFunc,
		AllowedMethods:     co.AllowedMethods,
		AllowedHeaders:     co.AllowedHeaders,
		ExposedHeaders:     co.ExposedHeaders,
		AllowCredentials:   co.AllowCredentials,
		MaxAge:             co.MaxAge,
		OptionsPassthrough: co.OptionsPassthrough,
		Debug:              co.Debug,
	}
}

func DefaultCors() Middleware {
	corsMiddleware := cors.New(cors.Options{})
	return corsMiddleware.Handler
}

func Cors(options CorsOptions) Middleware {
	corsMiddleware := cors.New(options.toChiOptions())
	return corsMiddleware.Handler
}
