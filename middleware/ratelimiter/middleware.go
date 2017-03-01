package ratelimiter

import (
	"net/http"
	"time"

	"github.com/gkarlik/quark-go/logger"
	"golang.org/x/time/rate"
)

const componentName = "RateLimiterMiddleware"

// Middleware is default quark implementation of request execution limitation mechanism.
type Middleware struct {
	Limiter *rate.Limiter // basic interval rate limiter
}

// NewRateLimiterMiddleware creates instance of Rate Limiter Middleware.
// Interval is a frequency of requests that are allowed to be handle.
func NewRateLimiterMiddleware(interval time.Duration) *Middleware {
	return &Middleware{
		Limiter: rate.NewLimiter(rate.Every(interval), 1),
	}
}

// Handle handles rate limits in HTTP server.
// Responds with "429 - Too Many Requests" if requests frequency is to high.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := m.limit(w); !ok {
			return
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// HandleWithNext handles rate limits in HTTP server.
// Responds with "429 - Too Many Requests" if requests frequency is to high.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if ok := m.limit(w); !ok {
		return
	}

	if next != nil {
		next(w, r)
	}
}

func (m Middleware) limit(w http.ResponseWriter) bool {
	if m.Limiter.Allow() == false {
		logger.Log().InfoWithFields(logger.Fields{
			"component": componentName,
		}, "Too many request for the interval")
		w.WriteHeader(429) // too many requests
		w.Write([]byte{})

		return false
	}
	return true
}
