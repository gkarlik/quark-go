package metrics

import (
	"time"

	"github.com/gkarlik/quark-go/system"
)

// Metric represents metric collected by the service
type Metric struct {
	Date   time.Time              // metric date - default: time.Now()
	Name   string                 // metric name
	Values map[string]interface{} // metric values
	Tags   map[string]string      // metric tags
}

// Reporter represents metrics reporter mechanism.
type Reporter interface {
	Report(ms ...Metric) error

	system.Disposer
}
