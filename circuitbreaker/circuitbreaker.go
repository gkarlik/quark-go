package circuitbreaker

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

// CircuitBreaker represents Circuit Breaker pattern mechanism
type CircuitBreaker interface {
	Execute(f func() (interface{}, error), opts ...Option) (interface{}, error)
}

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
		log.WithFields(log.Fields{
			"error":    err,
			"attempts": options.Attempts,
		}).Warn("Detected failure. Retrying...")

		for i := 1; i <= options.Attempts; i++ {
			log.WithField("timeout", options.Timeout).Info("Sleeping for configured timeout...")
			time.Sleep(options.Timeout)

			log.WithFields(log.Fields{
				"attempt": i,
				"from":    options.Attempts,
			}).Info("Retrying execution...")

			r, err = f()
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"attempt": i,
				}).Warn("Last execution failed")
				continue
			} else {
				log.WithField("attempt", i).Info("Last retry succeed")
				return r, nil
			}
		}
		log.Error("All retries failed.")

		return nil, err
	}
	return r, nil
}
