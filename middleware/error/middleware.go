package error

import (
	"net/http"

	"github.com/gkarlik/quark-go/logger"
)

const componentName = "RequestErrorMiddleware"

// Middleware is responsible for recovery from panic errors in HTTP pipeline.
type Middleware struct {
}

// NewRequestErrorMiddleware creates instance of Request Error Middleware.
func NewRequestErrorMiddleware() *Middleware {
	return &Middleware{}
}

// Handle recovers from panic error in handler.
func (m Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer m.recover(w)

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// HandleWithNext recovers from panic error in handler.
// This is method to support Negroni library.
func (m Middleware) HandleWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer m.recover(w)

	if next != nil {
		next(w, r)
	}
}

func (m Middleware) recover(w http.ResponseWriter) {
	if err := recover(); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)

		logger.Log().ErrorWithFields(logger.Fields{
			"component": componentName,
			"err":       err,
		}, "Recovered from panic error in handler")
	}
}
