package prometheus

import (
	"net/http"
	"sync"

	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	addr          = "localhost:9999"
	endPointName  = "/metrics"
	componentName = "PrometheusMetricsExposer"
)

// MetricsExposer represents metrics collecting mechanism based on Prometheus.
type MetricsExposer struct {
	m       sync.Mutex             // mutex for synchronizing registered metrics
	metrics []prometheus.Collector // registered metrics
	Options Options                // options
}

// Option represents function which is used to apply metrics exposer options.
type Option func(*Options)

// Options represents cofiguration options for metrics exposer.
type Options struct {
	Address      string // metrics endpoint address
	EndPointName string // metrics endpoint name
}

// Address allows to set endpoint address in format <server>:<port>.
func Address(address string) Option {
	return func(o *Options) {
		o.Address = address
	}
}

// EndPointName allows to set endpoint address name.
func EndPointName(endPointName string) Option {
	return func(o *Options) {
		o.EndPointName = endPointName
	}
}

// NewMetricsExposer creates instance of metrics exposer based on Prometheus.
// It configures metrics exposer based on options passed as arguments.
func NewMetricsExposer(opts ...Option) *MetricsExposer {
	options := new(Options)
	options.Address = addr
	options.EndPointName = endPointName

	for _, o := range opts {
		o(options)
	}

	return &MetricsExposer{
		m:       sync.Mutex{},
		metrics: make([]prometheus.Collector, 0),
		Options: *options,
	}
}

// Dispose cleans up MetricsExposer instance.
func (mex *MetricsExposer) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing metrics exposer component")

	mex.m.Lock()
	for _, c := range mex.metrics {
		prometheus.Unregister(c)
	}
	mex.m.Unlock()

	mex.metrics = nil
}

// Expose creates HTTP server with only one handler defined by Address and EndPointName.
// All metrics registered with Create* method will be exposed via this endpoint.
// Panics if cannot start HTTP server.
func (mex *MetricsExposer) Expose() {
	logger.Log().InfoWithFields(logger.Fields{
		"address":  mex.Options.Address,
		"endpoint": mex.Options.EndPointName,
	}, "Exposing metrics via HTTP endpoint")

	http.Handle(mex.Options.EndPointName, promhttp.Handler())
	logger.Log().Fatal(http.ListenAndServe(mex.Options.Address, nil))
}

// ExposeHandler returns HTTP handler with all metrics registered with Create* method.
func (mex *MetricsExposer) ExposeHandler() http.Handler {
	logger.Log().Info("Exposing HTTP handler")

	return promhttp.Handler()
}

type metric struct {
	name        string
	description string
}

func (m *metric) Name() string {
	return m.name
}

func (m *metric) Description() string {
	return m.description
}

// Gauge is a metric that represents a single numerical value that can arbitrarily go up and down.
type Gauge struct {
	metric

	g prometheus.Gauge
}

// Set sets value on gauge.
func (g *Gauge) Set(value float64) {
	g.g.Set(value)
}

// CreateGauge creates and registers metric of type Gauge.
func (mex *MetricsExposer) CreateGauge(name, description string) metrics.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: description,
	})

	mex.register(gauge)

	return &Gauge{
		metric: metric{
			name:        name,
			description: description,
		},
		g: gauge,
	}
}

// Counter is a cumulative metric that represents a single numerical value that only ever goes up.
type Counter struct {
	metric

	c prometheus.Counter
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.c.Inc()
}

// CreateCounter creates and registers metric of type Counter.
func (mex *MetricsExposer) CreateCounter(name, description string) metrics.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: description,
	})

	mex.register(counter)

	return &Counter{
		metric: metric{
			name:        name,
			description: description,
		},
		c: counter,
	}
}

// Histogram samples observations (usually things like request durations or response sizes) and counts them in configurable buckets. It also provides a sum of all observed values.
type Histogram struct {
	metric

	h prometheus.Histogram
}

// Observe adds a single observation to the histogram.
func (h *Histogram) Observe(value float64) {
	h.h.Observe(value)
}

// CreateHistogram creates and registers metric of type Histogram.
func (mex *MetricsExposer) CreateHistogram(name, description string, buckets []float64) metrics.Histogram {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    description,
		Buckets: buckets,
	})

	mex.register(histogram)

	return &Histogram{
		metric: metric{
			name:        name,
			description: description,
		},
		h: histogram,
	}
}

// Summary samples observations (usually things like request durations and response sizes).
type Summary struct {
	metric

	s prometheus.Summary
}

// Observe adds a single observation to the summary.
func (s *Summary) Observe(value float64) {
	s.s.Observe(value)
}

// CreateSummary creates and registers metric of type Summary.
func (mex *MetricsExposer) CreateSummary(name, description string, objectives map[float64]float64) metrics.Summary {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       name,
		Help:       description,
		Objectives: objectives,
	})

	mex.register(summary)

	return &Summary{
		metric: metric{
			name:        name,
			description: description,
		},
		s: summary,
	}
}

func (mex *MetricsExposer) register(c prometheus.Collector) {
	mex.m.Lock()
	mex.metrics = append(mex.metrics, c)
	mex.m.Unlock()

	prometheus.Register(c)
}
