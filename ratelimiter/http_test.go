package ratelimiter_test

import (
	"github.com/gkarlik/quark/ratelimiter"
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

	hl := ratelimiter.NewHTTPRateLimiter(interval)
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
