package zmiddlewares

import (
	"bytes"
	"github.com/zondax/golem/pkg/logger"
	"net/http"
	"regexp"
	"time"
)

type LoggingMiddlewareOptions struct {
	ExcludePaths []string
	Enable       bool
}

func LoggingMiddleware(options LoggingMiddlewareOptions) func(http.Handler) http.Handler {
	excludeRegexps := make([]*regexp.Regexp, len(options.ExcludePaths))
	for i, path := range options.ExcludePaths {
		excludeRegexps[i] = PathToRegexp(path)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath := r.URL.Path
			for _, re := range excludeRegexps {
				if re.MatchString(requestPath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			log := logger.GetLoggerFromContext(r.Context())
			rw := &responseWriter{ResponseWriter: w}
			if log.IsDebugEnabled() {
				rw.body = &bytes.Buffer{}
			}

			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)
			if log.IsDebugEnabled() {
				log.Debugf("Method: %s - URL: %s | Status: %d - Duration: %s - Response Body: %s",
					r.Method, r.URL.String(), rw.status, duration, string(rw.Body()))
				return
			}

			log.Infof("Method: %s - URL: %s | Status: %d - Duration: %s",
				r.Method, r.URL.String(), rw.status, duration)
		})
	}
}
