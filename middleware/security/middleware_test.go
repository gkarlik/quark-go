package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkarlik/quark-go/middleware/security"
	"github.com/stretchr/testify/assert"
)

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestSecurityMiddleware(t *testing.T) {
	sm := security.NewRequestSecurityMiddleware()
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h := sm.Handle(&TestHttpHandler{})
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	v := w.Header().Get("X-Content-Type-Options")
	assert.Equal(t, v, "nosniff")
}

func TestSecurityMiddlewareWithNext(t *testing.T) {
	sm := security.NewRequestSecurityMiddleware()
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	th := &TestHttpHandler{}
	sm.HandleWithNext(w, r, th.ServeHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	v := w.Header().Get("X-Content-Type-Options")
	assert.Equal(t, v, "nosniff")
}
