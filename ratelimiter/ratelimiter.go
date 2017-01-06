package ratelimiter

import (
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"time"
)

// RateLimiter represents execution limitation mechanism
type RateLimiter interface {
	Limit(f func() (interface{}, error)) (interface{}, error)
}

// DefaultRateLimiter is default quark implementation of execution limitation mechanism
type DefaultRateLimiter struct {
	l *rate.Limiter
}

// NewDefaultRateLimiter creates instance of DefaultRateLimiter
func NewDefaultRateLimiter(interval time.Duration) *DefaultRateLimiter {
	return &DefaultRateLimiter{
		l: rate.NewLimiter(rate.Every(interval), 1),
	}
}

// Limit limits function f execution to particular interval
func (d DefaultRateLimiter) Limit(f func() (interface{}, error)) (interface{}, error) {
	if err := d.l.Wait(context.TODO()); err != nil {
		return nil, err
	}
	return f()
}
