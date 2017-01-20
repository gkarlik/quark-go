package jwt_test

import (
	"github.com/gkarlik/quark/auth/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestHTTPRateLimiter(t *testing.T) {
	auth := jwt.NewAuthenticationMiddleware("secret", func(jwt.Credentials) (interface{}, error) {
		return nil, nil
	})
	h := auth.GenerateToken(&TestHttpHandler{})

	srv := httptest.NewServer(h)
	defer srv.Close()

	r, err := http.Get(srv.URL)
	assert.NoError(t, err, "Error while calling GET on HTTP server")
	assert.Equal(t, 200, r.StatusCode)
}
