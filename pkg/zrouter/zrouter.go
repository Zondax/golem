package zrouter

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/zmiddlewares"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultAddress        = ":8080"
	defaultTimeOut        = 240000
	uptimeMetricName      = "uptime"
	appVersionMetric      = "app_version"
	appRevisionMetric     = "app_revision"
	defaultUpdateInterval = 5 * time.Minute

	// metrics

	topNRequestMetricDefault = 5
)

type JWTUsageMetricsConfig struct {
	RemoteCache       zcache.RemoteCache
	Enable            bool
	TokenDetailsTTL   time.Duration
	UsageMetricTTL    time.Duration
	UpdateInterval    time.Duration
	TopNRequestMetric int
}

type SystemMetrics struct {
	Enable         bool
	UpdateInterval time.Duration
}

type Config struct {
	ReadTimeOut           time.Duration
	WriteTimeOut          time.Duration
	Logger                *logger.Logger
	SystemMetrics         SystemMetrics
	EnableRequestID       bool
	JWTUsageMetricsConfig JWTUsageMetricsConfig
	AppVersion            string
	AppRevision           string
}

func (c *Config) setDefaultValues() {
	if c.ReadTimeOut == 0 {
		c.ReadTimeOut = time.Duration(defaultTimeOut) * time.Millisecond
	}

	if c.WriteTimeOut == 0 {
		c.WriteTimeOut = time.Duration(defaultTimeOut) * time.Millisecond
	}

	if c.Logger == nil {
		l := logger.NewLogger()
		c.Logger = l
	}
}

type RegisteredRoute struct {
	Method string
	Path   string
}

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
	Handle(pattern string, handler HandlerFunc)
	Route(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes
	Mount(pattern string, handler HandlerFunc)
	Group(prefix string) Routes
	Use(middlewares ...zmiddlewares.Middleware) Routes
	NoRoute(handler HandlerFunc)
	GetRegisteredRoutes() []RegisteredRoute
	SetDefaultMiddlewares(loggingOptions zmiddlewares.LoggingMiddlewareOptions)
	GetHandler() http.Handler
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	ServeFiles(routePattern string, httpHandler http.Handler)
}

type zrouter struct {
	router             *chi.Mux
	middlewares        []zmiddlewares.Middleware
	defaultMiddlewares []zmiddlewares.Middleware
	metricsServer      metrics.TaskMetrics
	routes             []RegisteredRoute
	mutex              sync.Mutex
	config             *Config
}

func New(metricsServer metrics.TaskMetrics, config *Config) ZRouter {
	if config == nil {
		config = &Config{}
	}

	if config.AppVersion == "" || config.AppRevision == "" {
		panic("appVersion and appRevision are mandatory.")
	}

	config.setDefaultValues()
	zr := &zrouter{
		router:        chi.NewRouter(),
		metricsServer: metricsServer,
		config:        config,
	}

	if config.SystemMetrics.Enable {
		if err := metrics.RegisterSystemMetrics(metricsServer); err != nil {
			logger.GetLoggerFromContext(context.Background()).Errorf("Error registering metrics %v", err)
		}

		updateInterval := config.SystemMetrics.UpdateInterval
		go metrics.UpdateSystemMetrics(metricsServer, updateInterval)
	}

	return zr
}

func (r *zrouter) SetDefaultMiddlewares(loggingOptions zmiddlewares.LoggingMiddlewareOptions) {
	if err := zmiddlewares.RegisterRequestMetrics(r.metricsServer); err != nil {
		logger.GetLoggerFromContext(context.Background()).Errorf("Error registering metrics %v", err)
	}

	r.useDefaultMiddleware(zmiddlewares.ErrorHandlerMiddleware())
	r.useDefaultMiddleware(zmiddlewares.RequestMetrics(r.metricsServer))
	if loggingOptions.Enable {
		r.useDefaultMiddleware(zmiddlewares.LoggingMiddleware(loggingOptions))
	}

	if r.config.JWTUsageMetricsConfig.Enable {
		if r.config.JWTUsageMetricsConfig.RemoteCache == nil {
			panic("If JWTUsageMetricsConfig is enable then remote cache is mandatory")
		}
		r.Use(zmiddlewares.JWTUsageMiddleware(r.config.JWTUsageMetricsConfig.RemoteCache, r.config.JWTUsageMetricsConfig.TokenDetailsTTL, r.config.JWTUsageMetricsConfig.UsageMetricTTL))
	}
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

	r.config.Logger.Infof("Start server at %v", address)

	if r.config.JWTUsageMetricsConfig.Enable {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go LogTopJWTPathMetrics(ctx, r.config.JWTUsageMetricsConfig.RemoteCache, r.config.JWTUsageMetricsConfig.UpdateInterval, r.config.JWTUsageMetricsConfig.TopNRequestMetric)
	}

	server := &http.Server{
		Addr:         address,
		Handler:      r.router,
		ReadTimeout:  r.config.ReadTimeOut,
		WriteTimeout: r.config.WriteTimeOut,
	}

	if err := r.metricsServer.RegisterMetric(uptimeMetricName, "Timestamp of when the application was started", []string{}, &collectors.Gauge{}); err != nil {
		panic(err)
	}

	if err := r.metricsServer.UpdateMetric(uptimeMetricName, float64(time.Now().Unix())); err != nil {
		panic(err)
	}

	if err := r.metricsServer.RegisterMetric(appVersionMetric, "Current version of the application", []string{appVersionMetric}, &collectors.Gauge{}); err != nil {
		panic(err)
	}

	if err := r.metricsServer.UpdateMetric(appVersionMetric, 1, r.config.AppVersion); err != nil {
		panic(err)
	}

	if err := r.metricsServer.RegisterMetric(appRevisionMetric, "Current revision of the application", []string{appRevisionMetric}, &collectors.Gauge{}); err != nil {
		panic(err)
	}

	if err := r.metricsServer.UpdateMetric(appRevisionMetric, 1, r.config.AppRevision); err != nil {
		panic(err)
	}

	return server.ListenAndServe()
}

func (r *zrouter) Method(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	chiHandler := getChiHandler(handler)
	finalHandler := r.applyMiddlewares(chiHandler, middlewares...)
	r.router.Method(method, path, finalHandler)

	r.mutex.Lock()
	r.routes = append(r.routes, RegisteredRoute{Method: method, Path: path})
	r.mutex.Unlock()
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

func (r *zrouter) Mount(pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("handler is required for Mount")
	}

	r.router.Mount(pattern, getChiHandler(handler))
}

func (r *zrouter) Handle(pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("handler is mandatory")
	}

	r.router.Handle(pattern, getChiHandler(handler))
}

func (r *zrouter) GetRegisteredRoutes() []RegisteredRoute {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	routesCopy := make([]RegisteredRoute, len(r.routes))
	copy(routesCopy, r.routes)
	return routesCopy
}

func (r *zrouter) GetHandler() http.Handler {
	return r.router
}

func (r *zrouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *zrouter) ServeFiles(routePattern string, httpHandler http.Handler) {
	r.router.Handle(routePattern, httpHandler)
}

func LogTopJWTPathMetrics(ctx context.Context, zCache zcache.RemoteCache, updateInterval time.Duration, topN int) {
	if updateInterval == 0 {
		updateInterval = defaultUpdateInterval
	}

	if topN == 0 {
		topN = topNRequestMetricDefault
	}

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.GetLoggerFromContext(ctx).Errorf("Recovered in LogTopJWTPathMetrics: %v", r)
					}
				}()

				var logEntries []string
				topTokens, err := zCache.ZRevRangeWithScores(ctx, zmiddlewares.PathUsageByJWTKey, 0, int64(topN-1))
				if err != nil {
					logger.GetLoggerFromContext(ctx).Errorf("Error fetching top tokens from cache: %v", err)
					return
				}

				for _, item := range topTokens {
					metricKey, ok := item.Member.(string)
					if !ok {
						logger.GetLoggerFromContext(ctx).Errorf("Failed to type assert member to string: %v", item.Member)
						continue
					}

					parts := strings.Split(metricKey, ":")
					if len(parts) != 2 {
						logger.GetLoggerFromContext(ctx).Errorf("Unexpected metric key format: %v", metricKey)
						continue
					}

					jti, path := parts[0], parts[1]
					count := item.Score
					logEntries = append(logEntries, fmt.Sprintf("JWT ID: %s, Path: %s, Count: %f", jti, path, count))
				}

				if len(logEntries) > 0 {
					logger.GetLoggerFromContext(ctx).Info(strings.Join(logEntries, "\n"))
				}
			}()
		case <-ctx.Done():
			logger.GetLoggerFromContext(ctx).Info("LogTopJWTPathMetrics stopped due to context cancellation")
			return
		}
	}
}

func (r *zrouter) useDefaultMiddleware(middlewares ...zmiddlewares.Middleware) {
	r.defaultMiddlewares = append(r.defaultMiddlewares, middlewares...)
}

func (r *zrouter) applyMiddlewares(handler http.HandlerFunc, middlewares ...zmiddlewares.Middleware) http.Handler {
	var wrappedHandler http.Handler = handler
	// The order of middleware application is crucial: Route-specific middlewares are applied first,
	// followed by router-level general middlewares. This ensures that general middlewares, which often
	// handle logging, security, etc... are executed first. This sequence is
	// important to maintain consistency in logging and to apply security measures before route-specific
	// logic is executed.

	for _, mw := range middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	for _, mw := range r.middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	for _, mw := range r.defaultMiddlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	if r.config.EnableRequestID {
		r.Use(zmiddlewares.RequestID()) // IMPORTANT: RequestID MUST always be the LAST middleware applied
	}
	return wrappedHandler
}
