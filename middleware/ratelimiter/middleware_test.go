package ratelimiter_test

import (
	"github.com/gkarlik/quark-go/middleware/ratelimiter"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestHTTPRateLimiter(t *testing.T) {
	interval := time.Second

	hl := ratelimiter.NewRateLimiterMiddleware(interval)
	h := hl.Handle(&TestHttpHandler{})

	srv := httptest.NewServer(h)
	defer srv.Close()

	r, err := http.Get(srv.URL)
	assert.NoError(t, err, "Error while calling GET on HTTP server")
	assert.Equal(t, http.StatusOK, r.StatusCode)

	time.Sleep(interval)

	r, err = http.Get(srv.URL)
	assert.NoError(t, err, "Error while calling GET on HTTP server")
	assert.Equal(t, http.StatusOK, r.StatusCode)

	for i := 0; i < 10; i++ {
		r, err := http.Get(srv.URL)
		assert.NoError(t, err, "Error while calling GET on HTTP server")
		assert.Equal(t, http.StatusTooManyRequests, r.StatusCode)
	}
}

func TestAuthenticationWithNext(t *testing.T) {
	interval := time.Second

	hl := ratelimiter.NewRateLimiterMiddleware(interval)
	th := &TestHttpHandler{}
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)

	w := httptest.NewRecorder()
	hl.HandleWithNext(w, r, th.ServeHTTP)
	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(interval)

	w = httptest.NewRecorder()
	hl.HandleWithNext(w, r, th.ServeHTTP)
	assert.Equal(t, http.StatusOK, w.Code)

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		hl.HandleWithNext(w, r, th.ServeHTTP)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	}
}
