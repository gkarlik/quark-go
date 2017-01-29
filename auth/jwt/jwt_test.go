package jwt_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gkarlik/quark-go/auth/jwt"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationMiddleware(t *testing.T) {
	assert.Panics(t, func() {
		var _ = jwt.NewAuthenticationMiddleware()
	})

	assert.Panics(t, func() {
		var _ = jwt.NewAuthenticationMiddleware(jwt.WithAuthenticationFunc(func(c jwt.Credentials) (jwt.Claims, error) {
			return jwt.Claims{}, nil
		}))
	})

	assert.Panics(t, func() {
		var _ = jwt.NewAuthenticationMiddleware(jwt.WithSecret("secret"))
	})

	am := jwt.NewAuthenticationMiddleware(jwt.WithAuthenticationFunc(func(c jwt.Credentials) (jwt.Claims, error) {
		return jwt.Claims{}, nil
	}), jwt.WithContextKey("NewKey"), jwt.WithSecret("NewSecret"))

	assert.Equal(t, "NewKey", am.Options.ContextKey)
	assert.Equal(t, "NewSecret", am.Options.Secret)
}

var am *jwt.AuthenticationMiddleware = jwt.NewAuthenticationMiddleware(jwt.WithAuthenticationFunc(func(c jwt.Credentials) (jwt.Claims, error) {
	if c.Username == "test" && c.Password == "test" {
		return jwt.Claims{
			Username: "test",
			Properties: map[string]interface{}{
				"A":    1,
				"Role": "test",
			},
		}, nil
	}
	return jwt.Claims{}, errors.New("Invalid username or password")
}), jwt.WithSecret("0123456789"))

func TestGenerateTokenMethod(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/generateToken", nil)

	am.GenerateToken(w, r)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestGenerateTokenInvalidCredentials(t *testing.T) {
	w := httptest.NewRecorder()

	c := "wrong json payload"
	r, _ := http.NewRequest(http.MethodPost, "/generateToken", strings.NewReader(c))

	am.GenerateToken(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGenerateTokenWrongCredentials(t *testing.T) {
	w := httptest.NewRecorder()

	c := jwt.Credentials{
		Username: "test",
		Password: "wrong",
	}

	payload, _ := json.Marshal(c)

	r, _ := http.NewRequest(http.MethodPost, "/generateToken", strings.NewReader(string(payload)))

	am.GenerateToken(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGenerateToken(t *testing.T) {
	w := httptest.NewRecorder()

	c := jwt.Credentials{
		Username: "test",
		Password: "test",
	}

	payload, _ := json.Marshal(c)

	r, _ := http.NewRequest(http.MethodPost, "/generateToken", strings.NewReader(string(payload)))

	am.GenerateToken(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

type AuthenticateHandler struct{}

func (h *AuthenticateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("In protected page"))
}

type Data struct {
	Token string `json:"token"`
}

var ah *AuthenticateHandler = &AuthenticateHandler{}

func TestAuthentication(t *testing.T) {
	w := httptest.NewRecorder()

	c := jwt.Credentials{
		Username: "test",
		Password: "test",
	}

	payload, _ := json.Marshal(c)
	r, _ := http.NewRequest(http.MethodPost, "/generateToken", strings.NewReader(string(payload)))

	am.GenerateToken(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	b, _ := ioutil.ReadAll(w.Body)
	var data Data
	json.Unmarshal(b, &data)

	r, _ = http.NewRequest(http.MethodGet, "/authenticate", nil)
	r.Header.Add("Authorization", "bearer "+data.Token)

	am.Authenticate(ah).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, string(body), "In protected page")
}

func TestIncorrectAuthorizationHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)
	r.Header.Add("Authorization", "wrong 0123456789")

	am.Authenticate(ah).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncorrectToken(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)
	r.Header.Add("Authorization", "bearer 0123456789")

	am.Authenticate(ah).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEmptyToken(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)
	r.Header.Add("Authorization", "bearer ")

	am.Authenticate(ah).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLackOfAuthorizationHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/authenticate", nil)

	am.Authenticate(ah).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
