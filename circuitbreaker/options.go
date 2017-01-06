package circuitbreaker

import (
	"time"
)

// Option represents function which is used to set service discovery options
type Option func(*Options)

// Options represents circuit breaker options
type Options struct {
	Attempts int
	Timeout  time.Duration
}

// Retry allows to set number of retries
func Retry(retries int) Option {
	return func(o *Options) {
		o.Attempts = retries
	}
}

// Timeout allows to set sleep period between retries
func Timeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}
