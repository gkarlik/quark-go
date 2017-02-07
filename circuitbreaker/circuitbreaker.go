package circuitbreaker

import (
	"time"

	"github.com/gkarlik/quark-go/logger"
)

// CircuitBreaker represents Circuit Breaker pattern mechanism.
type CircuitBreaker interface {
	Execute(f func() (interface{}, error), opts ...Option) (interface{}, error)
}

const componentName = "CircuitBreaker"

// DefaultCircuitBreaker is default quark-go implementation of Circuit Breaker pattern.
type DefaultCircuitBreaker struct{}

// Execute implements Circuit Breaker pattern for function f.
// Default settings are: 3 attempts (1 failure + 3 retries), 5 second sleep time between retries.
func (cb DefaultCircuitBreaker) Execute(f func() (interface{}, error), opts ...Option) (interface{}, error) {
	options := &Options{
		Attempts: 3,
		Timeout:  5 * time.Second,
	}
	for _, o := range opts {
		o(options)
	}

	r, err := f()
	if err != nil {
		logger.Log().WarningWithFields(logger.Fields{
			"error":     err,
			"attempts":  options.Attempts,
			"component": componentName,
		}, "Detected failure. Retrying...")

		for i := 1; i <= options.Attempts; i++ {
			logger.Log().InfoWithFields(logger.Fields{
				"timeout":   options.Timeout,
				"component": componentName,
			}, "Sleeping for configured timeout...")
			time.Sleep(options.Timeout)

			logger.Log().InfoWithFields(logger.Fields{
				"attempt":   i,
				"from":      options.Attempts,
				"component": componentName,
			}, "Retrying execution...")

			r, err = f()
			if err != nil {
				logger.Log().WarningWithFields(logger.Fields{
					"error":     err,
					"attempt":   i,
					"component": componentName,
				}, "Last execution failed")
				continue
			} else {
				logger.Log().InfoWithFields(logger.Fields{
					"attempt":   i,
					"component": componentName,
				}, "Last retry succeed")
				return r, nil
			}
		}
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "All retries failed.")

		return nil, err
	}
	return r, nil
}
