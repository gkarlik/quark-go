package circuitbreaker_test

import (
	"errors"
	"github.com/gkarlik/quark-go/circuitbreaker"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var counter = 1

func threeFailuresFunc() (interface{}, error) {
	switch counter {
	case 1:
		counter++
		return nil, errors.New("First error")
	case 2:
		counter++
		return nil, errors.New("Second error")
	case 3:
		counter++
		return nil, errors.New("Third error")
	}

	return counter, nil
}

func workingFunc() (interface{}, error) {
	return 1, nil
}

func TestCircuitBreaker(t *testing.T) {
	cb := &circuitbreaker.DefaultCircuitBreaker{}

	r, err := cb.Execute(workingFunc, circuitbreaker.Retry(3), circuitbreaker.Timeout(10*time.Millisecond))
	assert.NoError(t, err, "Error executing workingFunc()")
	assert.Equal(t, 1, r)

	r, err = cb.Execute(threeFailuresFunc, circuitbreaker.Retry(3), circuitbreaker.Timeout(100*time.Millisecond))
	assert.NoError(t, err, "Error executing threeFailuresFunc()")
	assert.Equal(t, 4, r)

	counter = 1

	r, err = cb.Execute(threeFailuresFunc, circuitbreaker.Retry(2), circuitbreaker.Timeout(200*time.Millisecond))
	assert.Error(t, err, "Error executing threeFailuresFunc()")
	assert.Equal(t, nil, r)
}
