package zmiddlewares

import (
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getLimiter(key string, maxRPM int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(maxRPM)), maxRPM)
		limiters[key] = limiter
	}

	return limiter
}

func RateLimitByFullPath(maxRPM int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Path
			limiter := getLimiter(key, maxRPM)

			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(maxRPM int) Middleware {
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(maxRPM)), maxRPM)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
