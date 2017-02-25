package ratelimiter

import (
	"net/http"
	"time"

	"github.com/gkarlik/quark-go/logger"
	"golang.org/x/time/rate"
)

const componentName = "HttpRateLimiter"

// HTTPRateLimiter is default quark implementation of execution limitation mechanism.
type HTTPRateLimiter struct {
	Limiter *rate.Limiter // basic interval rate limiter
}

// NewHTTPRateLimiter creates instance of DefaultRateLimiter.
// Interval is a frequency of requests that are allowed to be handle.
func NewHTTPRateLimiter(interval time.Duration) *HTTPRateLimiter {
	return &HTTPRateLimiter{
		Limiter: rate.NewLimiter(rate.Every(interval), 1),
	}
}

// Handle handles rate limits in HTTP server.
// Responds with "429 - Too Many Requests" if requests frequency is to high.
func (rl HTTPRateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.Limiter.Allow() == false {
			logger.Log().InfoWithFields(logger.Fields{
				"component": componentName,
			}, "Too many request for the interval")
			w.WriteHeader(429) // too many requests
			w.Write([]byte{})

			return
		}
		next.ServeHTTP(w, r)
	})
}

// HandleWithNext handles rate limits in HTTP server.
// Responds with "429 - Too Many Requests" if requests frequency is to high.
// This is method to support Negroni library.
func (rl HTTPRateLimiter) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if rl.Limiter.Allow() == false {
		logger.Log().InfoWithFields(logger.Fields{
			"component": componentName,
		}, "Too many request for the interval")
		w.WriteHeader(429) // too many requests
		w.Write([]byte{})

		logger.Log().Info(w)

		return
	}
	if next != nil {
		next(w, r)
	}
}
