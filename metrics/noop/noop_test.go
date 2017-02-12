package noop_test

import (
	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/metrics/noop"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNoopMetricsReported(t *testing.T) {
	r := noop.NewMetricsReporter()
	defer r.Dispose()

	ms := []metrics.Metric{
		{
			Date:   time.Now(),
			Name:   "Test 1",
			Type:   metrics.Other,
			Tags:   map[string]string{"key": "test1"},
			Values: map[string]interface{}{"key": 1},
		},
		{
			Date:   time.Now(),
			Name:   "Test 2",
			Type:   metrics.Other,
			Tags:   map[string]string{"key": "test2"},
			Values: map[string]interface{}{"key": 2},
		},
	}

	err := r.Report(ms...)
	assert.NoError(t, err, "Unexpected error reporting metrics")
}
