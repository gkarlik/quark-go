package ratelimiter

import (
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

// HTTPRateLimiter is default quark implementation of execution limitation mechanism
type HTTPRateLimiter struct {
	l *rate.Limiter
}

// NewHTTPRateLimiter creates instance of DefaultRateLimiter
func NewHTTPRateLimiter(interval time.Duration) *HTTPRateLimiter {
	return &HTTPRateLimiter{
		l: rate.NewLimiter(rate.Every(interval), 1),
	}
}

// Handle handles rate limits in http server
func (rl HTTPRateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.l.Allow() == false {
			w.WriteHeader(429) // to many requests
			w.Write([]byte{})

			return
		}
		next.ServeHTTP(w, r)
	})
}
