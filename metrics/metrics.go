package metrics

import (
	"time"

	"github.com/gkarlik/quark/system"
)

// Metric represents metric created by the service
type Metric struct {
	Date   time.Time
	Name   string
	Values map[string]interface{}
	Tags   map[string]string
}

// Reporter represents metrics reporter mechanism
type Reporter interface {
	Report(ms ...Metric) error
	system.Disposer
}
