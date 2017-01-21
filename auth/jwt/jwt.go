package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

// Credentials represents user credentials
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims represents user claims
type Claims struct {
	Username   string                 `json:"username"`
	Properties map[string]interface{} `json:"properties"`

	jwt.StandardClaims
}

// AuthenticationMiddleware represents HTTP middleware responsible for authentication (jwt based)
type AuthenticationMiddleware struct {
	Options Options
}

// AuthenticationFunc is a function used to authenticate user
type AuthenticationFunc func(credentials Credentials) (Claims, error)

// NewAuthenticationMiddleware creates instance of authentication middleware
func NewAuthenticationMiddleware(opts ...Option) *AuthenticationMiddleware {
	am := &AuthenticationMiddleware{
		Options: Options{
			ContextKey: "Token_Claims",
		},
	}

	for _, opt := range opts {
		opt(&am.Options)
	}

	if am.Options.Authenticate == nil {
		panic("Authentication function must be set")
	}

	if am.Options.Secret == "" {
		panic("Secret must be set")
	}

	return am
}

// Authenticate validates token using JWT mechanism
func (am AuthenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse token from Authorization header
		ah := r.Header.Get("Authorization")
		if ah == "" {
			log.Error("No authorization header")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		s := strings.Split(ah, " ")

		if len(s) != 2 || strings.ToUpper(s[0]) != "BEARER" {
			log.Error("Incorrect Authorization header")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tokenString := s[1]
		if tokenString == "" {
			log.Error("TokenString is empty")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}

			return []byte(am.Options.Secret), nil
		})

		if err != nil {
			log.Error(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// get tokens claims and pass it into the original request
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), am.Options.ContextKey, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Error("Token is invalid")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	})
}

// GenerateToken generates token using JWT mechanism, only HTTP POST method is allowed
func (am AuthenticationMiddleware) GenerateToken(w http.ResponseWriter, r *http.Request) {
	// check if method is POST
	if r.Method != http.MethodPost {
		log.WithFields(log.Fields{
			"method": r.Method,
		}).Info("Request method is")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// load credentials from request body
	var credentials Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentials); err != nil {
		log.WithFields(log.Fields{
			"error":       err,
			"credentials": credentials,
		}).Error("Cannot decode credentials")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// authenticate user using external function
	claims, err := am.Options.Authenticate(credentials)
	if err != nil {
		log.WithError(err).Error("Cannot authenticate user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(am.Options.Secret))
	if err != nil {
		// this code probably won't be execute - don't know how to test it
		log.WithError(err).Error("Cannot sign token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{ \"token\": %q }", tokenString)
}
