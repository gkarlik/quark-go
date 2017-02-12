package metrics

import (
	"time"

	"github.com/gkarlik/quark-go/system"
)

const (
	// Counter is a cumulative metric that represents a single numerical value that only ever goes up.
	Counter int64 = iota
	// Gauge is a metric that represents a single numerical value that can arbitrarily go up and down.
	Gauge
	// Histogram samples observations (usually things like request durations or response sizes) and counts them in configurable buckets. It also provides a sum of all observed values.
	Histogram
	// Set supports counting unique occurrences of events between flushes.
	Set
	// Summary samples observations (usually things like request durations and response sizes).
	Summary
	// Other type of metric.
	Other
)

// Metric represents metric collected by the service
type Metric struct {
	Date   time.Time              // metric date - default: time.Now()
	Type   int64                  // metric type
	Name   string                 // metric name
	Values map[string]interface{} // metric values
	Tags   map[string]string      // metric tags
}

// Reporter represents metrics reporter mechanism.
type Reporter interface {
	Report(ms ...Metric) error

	system.Disposer
}
