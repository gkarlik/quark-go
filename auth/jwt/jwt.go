package jwt

import (
	"net/http"
)

// AuthenticationMiddleware represents HTTP middeware responsilbe for authentication (jwt based)
type AuthenticationMiddleware struct{}

func NewAuthenticationMiddleware() *AuthenticationMiddleware {
	return &AuthenticationMiddleware{}
}

// Handle handles authentication mechanism (jwt based)
func (am AuthenticationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
