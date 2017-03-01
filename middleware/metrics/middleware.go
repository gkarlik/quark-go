package metrics

import (
	"net/http"
	"time"

	quark "github.com/gkarlik/quark-go"
)

const (
	componentName = "RequestMetricsMiddleware"
	metricName    = "response_time"
)

// Middleware is responsible for reporting metrics in HTTP pipeline.
type Middleware struct {
	s quark.Service // service
}

// NewRequestMetricsMiddleware creates instance of Request Metrics Middleware.
func NewRequestMetricsMiddleware(s quark.Service) *Middleware {
	return &Middleware{
		s: s,
	}
}

// Handle reports metrics about request.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			quark.ReportServiceValue(m.s, metricName, time.Since(start).Nanoseconds())
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
		quark.ReportServiceValue(m.s, metricName, time.Since(start).Nanoseconds())
	}()

	if next != nil {
		next(w, r)
	}
}
