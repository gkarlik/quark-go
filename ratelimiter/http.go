package ratelimiter

import (
	"net/http"
	"time"

	"github.com/gkarlik/quark-go/logger"
	"golang.org/x/time/rate"
)

const componentName = "HttpRateLimiter"

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
			logger.Log().InfoWithFields(logger.LogFields{
				"component": componentName,
			}, "Too many request for the interval")
			w.WriteHeader(429) // to many requests
			w.Write([]byte{})

			return
		}
		next.ServeHTTP(w, r)
	})
}
