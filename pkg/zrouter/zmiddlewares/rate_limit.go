package zmiddlewares

import (
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

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
