package jwt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"errors"

	"context"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gkarlik/quark-go/logger"
)

// Credentials represents user credentials (username and password).
type Credentials struct {
	Username   string            `json:"Username"`             // user name
	Password   string            `json:"Password"`             // password
	Properties map[string]string `json:"Properties,omitempty"` // additional properties
}

// Claims represents jwt claims.
type Claims struct {
	Username   string                 `json:"username"`   // user name
	Properties map[string]interface{} `json:"properties"` // additional jwt claims properties

	jwt.StandardClaims // standard jwt claims properties
}

const componentName = "JwtAuthenticationMiddleware"

// AuthenticationMiddleware represents HTTP middleware responsible for authentication (JWT based).
type AuthenticationMiddleware struct {
	Options Options // authentication middleware options
}

// AuthenticationFunc is a function used to authenticate user. Function receives user credentials and should return claims or an error.
type AuthenticationFunc func(credentials Credentials) (Claims, error)

// NewAuthenticationMiddleware creates instance of authentication middleware with options passed as argument.
// AuthenticationFunc and Secret options are required.
// Default context key value used to store jwt claims in request context is "Claims".
func NewAuthenticationMiddleware(opts ...Option) *AuthenticationMiddleware {
	am := &AuthenticationMiddleware{
		Options: Options{
			ContextKey: "Claims",
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

func (am AuthenticationMiddleware) authenticate(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	logger.Log().DebugWithFields(logger.Fields{
		"component": componentName,
	}, "Authenticating user request")

	// parse token from Authorization header
	ah := r.Header.Get("Authorization")
	if ah == "" {
		return nil, errors.New("No authorization header")
	}

	s := strings.Split(ah, " ")

	if len(s) != 2 || strings.ToUpper(s[0]) != "BEARER" {
		return nil, errors.New("Incorrect authorization header")
	}

	tokenString := s[1]
	if tokenString == "" {
		return nil, errors.New("TokenString is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return []byte(am.Options.Secret), nil
	})

	if err != nil {
		logger.Log().DebugWithFields(logger.Fields{
			"error":       err,
			"tokenString": tokenString,
			"component":   componentName,
		}, "Error parsing token string")
		return nil, errors.New("Error parsing token string")
	}

	// get tokens claims and pass it into the original request
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		ctx := context.WithValue(r.Context(), am.Options.ContextKey, *claims)
		return ctx, nil
	}
	return nil, errors.New("Token is invalid")

}

// Authenticate validates token using jwt specification. It parses token from 'Authorization' header which must be in form "bearer token".
func (am AuthenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := am.authenticate(w, r)
		if err != nil {
			logger.Log().ErrorWithFields(logger.Fields{"component": componentName, "error": err}, "Could not authentiate user")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if next != nil {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// AuthenticateWithNext validates token using jwt specification. It parses token from 'Authorization' header which must be in form "bearer token".
// This is method to support Negroni library.
func (am AuthenticationMiddleware) AuthenticateWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx, err := am.authenticate(w, r)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName, "error": err}, "Could not authentiate user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if next != nil {
		next(w, r.WithContext(ctx))
	}
}

// GenerateToken generates token using jwt specification, only HTTP POST method is allowed.
func (am AuthenticationMiddleware) GenerateToken(w http.ResponseWriter, r *http.Request) {
	logger.Log().DebugWithFields(logger.Fields{
		"component": componentName,
	}, "Generating token for the user")

	// check if method is POST
	if r.Method != http.MethodPost {
		logger.Log().ErrorWithFields(logger.Fields{
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
		logger.Log().ErrorWithFields(logger.Fields{
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
		logger.Log().ErrorWithFields(logger.Fields{
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
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot sign token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Log().InfoWithFields(logger.Fields{
		"token":     tokenString,
		"component": componentName,
	}, "Token generated - sending to client")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{ \"token\": %q }", tokenString)
}
