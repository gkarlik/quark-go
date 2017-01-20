package jwt

import (
	"net/http"

	"errors"

	"encoding/json"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

type Credentials struct {
	Username string                 `json:username`
	Password string                 `json:password`
	Claims   map[string]interface{} `json:claims`
}

// AuthenticationMiddleware represents HTTP middeware responsilbe for authentication (jwt based)
type AuthenticationMiddleware struct {
	secret       string
	authenticate AuthenticationFunc
}

type AuthenticationFunc func(credentials Credentials) (interface{}, error)

func NewAuthenticationMiddleware(secret string, authenticate AuthenticationFunc) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		secret:       secret,
		authenticate: authenticate,
	}
}

// Authenticate validates token using JWT mechanism
func (am AuthenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("bearer")
		if tokenString != "" {
			w.WriteHeader(403)
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Wrong signing method")
			}

			return am.secret, nil
		})

		if err != nil {
			w.WriteHeader(403)
			return
		}

		ctx := context.WithValue(r.Context(), "user", token.Header["user"])
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

// GenerateToken generates token using JWT mechanism, request must be done using POST method
func (am AuthenticationMiddleware) GenerateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var credentials Credentials

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&credentials)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		_, err = am.authenticate(credentials)
		if err != nil {
			w.WriteHeader(403)
			return
		}

		claims := jwt.MapClaims{}
		for k, v := range credentials.Claims {
			claims[k] = v
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString(am.secret)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		fmt.Fprintf(w, "{ 'token': '%q' }", tokenString)

		next.ServeHTTP(w, r)
	})
}
