package metrics

import (
	"github.com/gkarlik/quark-go/metrics"
	"net/http"
	"time"

	quark "github.com/gkarlik/quark-go"
)

const (
	componentName = "RequestMetricsMiddleware"
	metricName    = "response_time"
	metricDesc    = "Request response time"
)

// Middleware is responsible for reporting metrics in HTTP pipeline.
type Middleware struct {
	s quark.Service // service
	g metrics.Gauge // gauge
}

// NewRequestMetricsMiddleware creates instance of Request Metrics Middleware.
func NewRequestMetricsMiddleware(s quark.Service) *Middleware {
	gauge := s.Metrics().CreateGauge(metricName, metricDesc)

	return &Middleware{
		s: s,
		g: gauge,
	}
}

// Handle reports metrics about request.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			m.g.Set(float64(time.Since(start).Nanoseconds()))
		}()

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// HandleWithNext reports metrics about request.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	defer func() {
		m.g.Set(float64(time.Since(start).Nanoseconds()))
	}()

	if next != nil {
		next(w, r)
	}
}
