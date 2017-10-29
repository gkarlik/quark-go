package metrics

import (
	"net/http"

	"github.com/gkarlik/quark-go/system"
)

type metric interface {
	Name() string
	Description() string
}

// Gauge is a metric that represents a single numerical value that can arbitrarily go up and down.
type Gauge interface {
	metric

	Set(value float64)
}

// Counter is a cumulative metric that represents a single numerical value that only ever goes up.
type Counter interface {
	metric

	Inc()
}

// Histogram samples observations (usually things like request durations or response sizes) and counts them in configurable buckets. It also provides a sum of all observed values.
type Histogram interface {
	metric

	Observe(value float64)
}

// Summary samples observations (usually things like request durations and response sizes).
type Summary interface {
	metric

	Observe(value float64)
}

// Exposer represents metrics exposer mechanism.
type Exposer interface {
	CreateGauge(name, description string) Gauge
	CreateCounter(name, description string) Counter
	CreateHistogram(name, description string, buckets []float64) Histogram
	CreateSummary(name, description string, objectives map[float64]float64) Summary

	Expose()
	ExposeHandler() http.Handler

	system.Disposer
}
