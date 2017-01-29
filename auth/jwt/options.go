package jwt

// Option represents function which is used to apply authentication options.
type Option func(*Options)

// Options represents possible authentication options.
type Options struct {
	Authenticate AuthenticationFunc // function used to authenticate user
	Secret       string             // string used to encrypt/decrypt token
	ContextKey   string             // key in request context where claims are stored
}

// WithAuthenticationFunc allows to set function used to authenticate user.
func WithAuthenticationFunc(f AuthenticationFunc) Option {
	return func(opts *Options) {
		opts.Authenticate = f
	}
}

// WithSecret allows to set secret string to ecrypt/decrypt token.
func WithSecret(secret string) Option {
	return func(opts *Options) {
		opts.Secret = secret
	}
}

// WithContextKey allows to set key in request context which is used to store claims.
func WithContextKey(key string) Option {
	return func(opts *Options) {
		opts.ContextKey = key
	}
}
