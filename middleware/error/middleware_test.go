package error_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkarlik/quark-go/middleware/error"
	"github.com/stretchr/testify/assert"
)

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("Unexpected error")
}

func TestErrorMiddleware(t *testing.T) {
	em := error.NewRequestErrorMiddleware()
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h := em.Handle(&TestHttpHandler{})
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorMiddlewareWithNext(t *testing.T) {
	em := error.NewRequestErrorMiddleware()
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	th := &TestHttpHandler{}
	em.HandleWithNext(w, r, th.ServeHTTP)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
