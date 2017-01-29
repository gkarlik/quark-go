package jwt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gkarlik/quark-go/logger"
	"golang.org/x/net/context"
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

const componentName = "JwtAuthenticationMiddleware"

// AuthenticationMiddleware represents HTTP middleware responsible for authentication (JWT based)
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
		panic(fmt.Sprintf("[%s]: Cannot create instance - authentication function must be set!", componentName))
	}

	if am.Options.Secret == "" {
		panic(fmt.Sprintf("[%s]: Cannot create instance - secret must be set!", componentName))
	}

	return am
}

// Authenticate validates token using JWT mechanism
func (am AuthenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse token from Authorization header
		ah := r.Header.Get("Authorization")
		if ah == "" {
			logger.Log().ErrorWithFields(logger.LogFields{"component": componentName}, "No authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		s := strings.Split(ah, " ")

		if len(s) != 2 || strings.ToUpper(s[0]) != "BEARER" {
			logger.Log().ErrorWithFields(logger.LogFields{"component": componentName}, "Incorrect authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := s[1]
		if tokenString == "" {
			logger.Log().ErrorWithFields(logger.LogFields{"component": componentName}, "TokenString is empty")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("[%s]: Unexpected signing method", componentName)
			}

			return []byte(am.Options.Secret), nil
		})

		if err != nil {
			logger.Log().ErrorWithFields(logger.LogFields{
				"error":       err,
				"tokenString": tokenString,
				"component":   componentName,
			}, "Error parsing token string")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// get tokens claims and pass it into the original request
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), am.Options.ContextKey, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			logger.Log().ErrorWithFields(logger.LogFields{"component": componentName}, "Token is invalid")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	})
}

// GenerateToken generates token using JWT mechanism, only HTTP POST method is allowed
func (am AuthenticationMiddleware) GenerateToken(w http.ResponseWriter, r *http.Request) {
	// check if method is POST
	if r.Method != http.MethodPost {
		logger.Log().ErrorWithFields(logger.LogFields{
			"method":    r.Method,
			"component": componentName,
		}, "Incorrect http method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// load credentials from request body
	var credentials Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentials); err != nil {
		logger.Log().ErrorWithFields(logger.LogFields{
			"error":       err,
			"credentials": credentials,
			"component":   componentName,
		}, "Cannot decode credentials")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// authenticate user using external function
	claims, err := am.Options.Authenticate(credentials)
	if err != nil {
		logger.Log().ErrorWithFields(logger.LogFields{
			"error":     err,
			"component": componentName,
		}, "Cannot authenticate user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(am.Options.Secret))
	if err != nil {
		// this code probably won't be execute - don't know how to test it
		logger.Log().ErrorWithFields(logger.LogFields{
			"error":     err,
			"component": componentName,
		}, "Cannot sign token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Log().InfoWithFields(logger.LogFields{
		"token":     tokenString,
		"component": componentName,
	}, "Token generated - sending to client")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{ \"token\": %q }", tokenString)
}
