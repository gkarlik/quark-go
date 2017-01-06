package quark

import (
	"github.com/gkarlik/quark/broker"
	log "github.com/gkarlik/quark/logger"
	"github.com/gkarlik/quark/metrics"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
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

// Port allows to set service port
func Port(port int) Option {
	return func(o *Options) {
		o.Info.Port = port
	}
}

// Logger allows to set service logger
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

// Metrics allows to set service metrics reporter mechanism
func Metrics(r metrics.Reporter) Option {
	return func(o *Options) {
		o.Metrics = r
	}
}
