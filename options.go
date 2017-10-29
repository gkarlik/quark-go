package quark

import (
	"net/url"

	"context"
	"github.com/gkarlik/quark-go/broker"
	log "github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/service"
	"github.com/gkarlik/quark-go/service/discovery"
	"github.com/gkarlik/quark-go/service/trace"
)

// Option represents function which is used to apply service options.
type Option func(*Options)

// Options represents service options.
type Options struct {
	Info      service.Info               // service metadata information
	Logger    log.Logger                 // service logger interface implementation
	Discovery discovery.ServiceDiscovery // service discovery interface implementation
	Broker    broker.MessageBroker       // service message broker interface implementation
	Metrics   metrics.Exposer            // service metrics collector interface implementation
	Tracer    trace.Tracer               // service request tracer interface implementation

	Context context.Context // service context
}

// Name allows to set service name.
func Name(name string) Option {
	return func(o *Options) {
		o.Info.Name = name
	}
}

// Version allows to set service version.
func Version(version string) Option {
	return func(o *Options) {
		o.Info.Version = version
	}
}

// Tags allows to set service tag(s).
func Tags(tags ...string) Option {
	return func(o *Options) {
		o.Info.Tags = tags
	}
}

// Address allows to set service address.
func Address(url *url.URL) Option {
	return func(o *Options) {
		o.Info.Address = url
	}
}

// Logger allows to set service logger implementation. If it is not set, internal logger will be taken.
func Logger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// Discovery allows to set service discovery implementation.
func Discovery(d discovery.ServiceDiscovery) Option {
	return func(o *Options) {
		o.Discovery = d
	}
}

// Broker allows to set message broker implementation.
func Broker(b broker.MessageBroker) Option {
	return func(o *Options) {
		o.Broker = b
	}
}

// Tracer allows to set tracer implementation.
func Tracer(t trace.Tracer) Option {
	return func(o *Options) {
		o.Tracer = t
	}
}

// Metrics allows to set service metrics collector implementation.
func Metrics(e metrics.Exposer) Option {
	return func(o *Options) {
		o.Metrics = e
	}
}
