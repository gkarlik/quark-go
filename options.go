package quark

import (
	"net/url"

	"github.com/gkarlik/quark/broker"
	log "github.com/gkarlik/quark/logger"
	"github.com/gkarlik/quark/metrics"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/gkarlik/quark/service/trace"
	"golang.org/x/net/context"
)

// Option represents function which is used to set service options
type Option func(*Options)

// Options represents service options
type Options struct {
	Info      service.Info
	Logger    log.Logger
	Discovery discovery.ServiceDiscovery
	Broker    broker.MessageBroker
	Metrics   metrics.Reporter
	Tracer    trace.Tracer

	Context context.Context
}

// Name allows to set service name
func Name(name string) Option {
	return func(o *Options) {
		o.Info.Name = name
	}
}

// Version allows to set service version
func Version(version string) Option {
	return func(o *Options) {
		o.Info.Version = version
	}
}

// Tags allows to set service tag(s)
func Tags(tags ...string) Option {
	return func(o *Options) {
		o.Info.Tags = tags
	}
}

// Address allows to set service address
func Address(url *url.URL) Option {
	return func(o *Options) {
		o.Info.Address = url
	}
}

// Logger allows to set service logger. If it is not set, internal logger will be taken.
func Logger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// Discovery allows to set service discovery mechanism
func Discovery(d discovery.ServiceDiscovery) Option {
	return func(o *Options) {
		o.Discovery = d
	}
}

// Broker allows to set message broker
func Broker(b broker.MessageBroker) Option {
	return func(o *Options) {
		o.Broker = b
	}
}

// Tracer allows to set tracer
func Tracer(t trace.Tracer) Option {
	return func(o *Options) {
		o.Tracer = t
	}
}

// Metrics allows to set service metrics reporter mechanism
func Metrics(r metrics.Reporter) Option {
	return func(o *Options) {
		o.Metrics = r
	}
}
