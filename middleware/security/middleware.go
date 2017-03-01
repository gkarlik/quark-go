package security

import (
	"github.com/gkarlik/quark-go/logger"
	"net/http"
)

const componentName = "RequestSecurityMiddleware"

// Middleware is responsible for securing request.
type Middleware struct {
}

// NewRequestSecurityMiddleware creates instance of Request Logging Middleware.
func NewRequestSecurityMiddleware() *Middleware {
	return &Middleware{}
}

// Handle secures request.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.secureRequest(w)

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// HandleWithNext secures request.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	m.secureRequest(w)

	if next != nil {
		next(w, r)
	}
}

func (m Middleware) secureRequest(w http.ResponseWriter) {
	logger.Log().DebugWithFields(logger.Fields{
		"component": componentName,
	}, "Securing request")

	w.Header().Set("X-Content-Type-Options", "nosniff")
}
