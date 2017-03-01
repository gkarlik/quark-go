package logging

import (
	"context"
	"net/http"

	"github.com/gkarlik/quark-go/logger"
	uuid "github.com/satori/go.uuid"
)

const componentName = "RequestLoggingMiddleware"

// Middleware is responsible for logging information about HTTP request.
type Middleware struct {
	requestIDKey string // key used to store request ID in request context.
}

// NewRequestLoggingMiddleware creates instance of Request Logging Middleware.
func NewRequestLoggingMiddleware(reqIDKey string) *Middleware {
	return &Middleware{
		requestIDKey: reqIDKey,
	}
}

// Handle logs request information.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := m.logRequest(r)

		if next != nil {
			next.ServeHTTP(w, req)
		}
	})
}

// HandleWithNext logs request information.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	req := m.logRequest(r)

	if next != nil {
		next(w, req)
	}
}

func (m Middleware) logRequest(r *http.Request) *http.Request {
	reqID := uuid.NewV4()
	ctx := context.WithValue(r.Context(), m.requestIDKey, reqID.String())

	logger.Log().DebugWithFields(logger.Fields{
		"requestID": reqID,
		"request":   r,
		"component": componentName,
	}, "Request information")

	return r.WithContext(ctx)
}
