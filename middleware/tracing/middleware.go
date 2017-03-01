package tracing

import (
	"net/http"

	quark "github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/logger"
)

const (
	componentName = "RequestTracingMiddleware"
	spanName      = "http_root_request"
)

// Middleware is responsible for handling errors in HTTP pipeline.
type Middleware struct {
	s quark.Service // service
}

// NewRequestTracingMiddleware creates instance of Request Tracing Middleware.
func NewRequestTracingMiddleware(s quark.Service) *Middleware {
	return &Middleware{
		s: s,
	}
}

// Handle creates span associated with request.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.trace(w, r, next)
	})
}

// HandleWithNext creates span associated with request.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	m.trace(w, r, next)
}

func (m Middleware) trace(w http.ResponseWriter, r *http.Request, handler interface{}) {
	logger.Log().DebugWithFields(logger.Fields{
		"component": componentName,
	}, "Creating root span")

	req := r
	if m.s.Tracer() != nil {
		span := m.s.Tracer().StartSpan(spanName)
		defer span.Finish()

		ctx := m.s.Tracer().ContextWithSpan(r.Context(), span)
		req = r.WithContext(ctx)
	}

	if handler != nil {
		switch n := handler.(type) {
		case http.Handler:
			n.ServeHTTP(w, req)
		case http.HandlerFunc:
			n(w, req)
		}
	}
}
