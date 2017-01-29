package circuitbreaker

import (
	"time"

	"github.com/gkarlik/quark-go/logger"
)

// CircuitBreaker represents Circuit Breaker pattern mechanism
type CircuitBreaker interface {
	Execute(f func() (interface{}, error), opts ...Option) (interface{}, error)
}

const componentName = "CircuitBreaker"

// DefaultCircuitBreaker is default quark implementation of Circuit Breaker pattern
type DefaultCircuitBreaker struct{}

// Execute implements Circuit Breaker pattern for function f
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
		logger.Log().WarningWithFields(logger.LogFields{
			"error":     err,
			"attempts":  options.Attempts,
			"component": componentName,
		}, "Detected failure. Retrying...")

		for i := 1; i <= options.Attempts; i++ {
			logger.Log().InfoWithFields(logger.LogFields{
				"timeout":   options.Timeout,
				"component": componentName,
			}, "Sleeping for configured timeout...")
			time.Sleep(options.Timeout)

			logger.Log().InfoWithFields(logger.LogFields{
				"attempt":   i,
				"from":      options.Attempts,
				"component": componentName,
			}, "Retrying execution...")

			r, err = f()
			if err != nil {
				logger.Log().WarningWithFields(logger.LogFields{
					"error":     err,
					"attempt":   i,
					"component": componentName,
				}, "Last execution failed")
				continue
			} else {
				logger.Log().InfoWithFields(logger.LogFields{
					"attempt":   i,
					"component": componentName,
				}, "Last retry succeed")
				return r, nil
			}
		}
		logger.Log().ErrorWithFields(logger.LogFields{"component": componentName}, "All retries failed.")

		return nil, err
	}
	return r, nil
}
