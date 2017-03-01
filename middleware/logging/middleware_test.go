package logging_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkarlik/quark-go/middleware/logging"
	"github.com/stretchr/testify/assert"
)

const reqIDKey = "req-ID"

var reqID string

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID = r.Context().Value(reqIDKey).(string)
}

func TestLoggingMiddleware(t *testing.T) {
	reqID = ""

	lm := logging.NewRequestLoggingMiddleware(reqIDKey)
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h := lm.Handle(&TestHttpHandler{})
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, reqID, "Request ID is empty")
}

func TestLoggingMiddlewareWithNext(t *testing.T) {
	reqID = ""

	lm := logging.NewRequestLoggingMiddleware(reqIDKey)
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	th := &TestHttpHandler{}
	lm.HandleWithNext(w, r, th.ServeHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, reqID, "Request ID is empty")
}
