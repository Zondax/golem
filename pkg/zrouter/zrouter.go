package zrouter

import (
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/zrouter/zmiddlewares"
	"go.uber.org/zap"
	"net/http"
)

const (
	defaultAddress = ":8080"
)

type ZRouter interface {
	Routes
	Run(addr ...string) error
}

type Routes interface {
	GET(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	POST(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	PUT(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	PATCH(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	DELETE(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	Route(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	Group(prefix string) Routes
	Use(middlewares ...zmiddlewares.Middleware) Routes
	NoRoute(handler HandlerFunc)
}

type zrouter struct {
	router        *chi.Mux
	middlewares   []zmiddlewares.Middleware
	metricsServer metrics.TaskMetrics
	appName       string
}

func New(appName string, metricsServer metrics.TaskMetrics) ZRouter {
	zr := &zrouter{
		router:        chi.NewRouter(),
		metricsServer: metricsServer,
		appName:       appName,
	}
	return zr
}

func (r *zrouter) SetDefaultMiddlewares() {
	r.Use(zmiddlewares.ErrorHandlerMiddleware)
	r.Use(zmiddlewares.RequestID())
	if err := zmiddlewares.RegisterRequestMetrics(r.appName, r.metricsServer); err != nil {
		zap.S().With("err", err).Error("Error registering metrics")
	}

	r.Use(zmiddlewares.RequestMetrics(r.appName, r.metricsServer))
}

func (r *zrouter) Group(prefix string) Routes {
	newRouter := &zrouter{
		router: chi.NewRouter(),
	}

	r.router.Group(func(groupRouter chi.Router) {
		groupRouter.Mount(prefix, newRouter.router)
	})

	return newRouter
}

func (r *zrouter) Run(addr ...string) error {
	address := defaultAddress
	if len(addr) > 0 {
		address = addr[0]
	}
	return http.ListenAndServe(address, r.router)
}

func (r *zrouter) applyMiddlewares(handler http.HandlerFunc, middlewares ...zmiddlewares.Middleware) http.Handler {
	var wrappedHandler http.Handler = handler
	for _, mw := range middlewares {
		wrappedHandler = mw(wrappedHandler)
	}
	return wrappedHandler
}

func (r *zrouter) Method(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	chiHandler := getChiHandler(handler)
	finalHandler := r.applyMiddlewares(chiHandler, middlewares...)
	r.router.Method(method, path, finalHandler)
	return r
}

func (r *zrouter) GET(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(http.MethodGet, path, handler, middlewares...)
	return r
}

func (r *zrouter) POST(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(http.MethodPost, path, handler, middlewares...)
	return r
}

func (r *zrouter) PUT(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(http.MethodPut, path, handler, middlewares...)
	return r
}

func (r *zrouter) PATCH(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(http.MethodPatch, path, handler, middlewares...)
	return r
}

func (r *zrouter) DELETE(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(http.MethodDelete, path, handler, middlewares...)
	return r
}

func (r *zrouter) Route(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	r.Method(method, path, handler, middlewares...)
	return r
}

func (r *zrouter) NoRoute(handler HandlerFunc) {
	r.router.NotFound(getChiHandler(handler))
}

func (r *zrouter) Use(middlewares ...zmiddlewares.Middleware) Routes {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}
