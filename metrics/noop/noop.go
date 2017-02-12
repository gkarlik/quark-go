package noop

import (
	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/metrics"
)

const componentName = "NoopMetricsReporter"

// MetricsReporter represents NOOP (No Operation) metrics collecting mechanism.
type MetricsReporter struct{}

// NewMetricsReporter creates instance of NOOP metrics reporter.
func NewMetricsReporter() *MetricsReporter {
	return &MetricsReporter{}
}

// Report only logs metrics using quark-go logger.
func (mr *MetricsReporter) Report(ms ...metrics.Metric) error {
	logger.Log().InfoWithFields(logger.Fields{"metrics": ms}, "Reporting metrics")

	return nil
}

// Dispose cleans up NOOP MetricsReporter instance.
func (mr *MetricsReporter) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing metrics reporter component")
}
