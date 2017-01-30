package discovery

import (
	"github.com/gkarlik/quark-go/service"
	lb "github.com/gkarlik/quark-go/service/loadbalancer"
)

// Option represents function which is used to apply service discovery options.
type Option func(*Options)

// Options represents service discovery options.
type Options struct {
	Info     service.Info             // service info
	Strategy lb.LoadBalancingStrategy // load balancing strategy
}

// ByInfo allows to discover service by its info metadata.
func ByInfo(i service.Info) Option {
	return func(opts *Options) {
		opts.Info = i
	}
}

// WithInfo allows to registed service by its info metadata.
func WithInfo(i service.Info) Option {
	return ByInfo(i)
}

// ByName allows to discover service by its name.
func ByName(name string) Option {
	return func(opts *Options) {
		opts.Info.Name = name
	}
}

// ByTag allows to discover service by its tag(s).
func ByTag(tag string) Option {
	return func(opts *Options) {
		opts.Info.Tags = []string{tag}
	}
}

// UsingLBStrategy allows to discover service using specified load balancing startegy.
func UsingLBStrategy(s lb.LoadBalancingStrategy) Option {
	return func(opts *Options) {
		opts.Strategy = s
	}
}
