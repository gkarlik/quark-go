package jwt

// Option represents function which is used to set authentication options
type Option func(*Options)

// Options represents authentication options
type Options struct {
	Authenticate AuthenticationFunc
	Secret       string
	ContextKey   string
}

// WithAuthenticationFunc allows to function used to authenticate user
func WithAuthenticationFunc(f AuthenticationFunc) Option {
	return func(opts *Options) {
		opts.Authenticate = f
	}
}

// WithSecret allows to set secret
func WithSecret(secret string) Option {
	return func(opts *Options) {
		opts.Secret = secret
	}
}

// WithContextKey allows to set key in request context which will be used to store token claims
func WithContextKey(key string) Option {
	return func(opts *Options) {
		opts.ContextKey = key
	}
}
