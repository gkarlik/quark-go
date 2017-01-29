package quark

import (
	"github.com/gkarlik/quark-go/broker"
	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/service"
	"github.com/gkarlik/quark-go/service/discovery"
	"github.com/gkarlik/quark-go/service/trace"
	"github.com/gkarlik/quark-go/system"
)

// Service represents service instance
type Service interface {
	Info() service.Info
	Options() Options
	Log() logger.Logger
	Discovery() discovery.ServiceDiscovery
	Broker() broker.MessageBroker
	Metrics() metrics.Reporter
	Tracer() trace.Tracer

	system.Disposer
}

// RPCService represents service which exposes procedures that could be called remotelly
type RPCService interface {
	Service

	RegisterServiceInstance(server interface{}, serviceInstance interface{}) error
}

// ServiceBase is base structure for custom service
type ServiceBase struct {
	options Options
}

// NewService creates instance of service
func NewService(opts ...Option) *ServiceBase {
	s := &ServiceBase{
		options: Options{
			Info:   service.Info{},
			Logger: logger.Log(),
		},
	}

	for _, opt := range opts {
		opt(&s.options)
	}

	if s.Info().Name == "" {
		panic("Service name option must be specified")
	}

	if s.Info().Version == "" {
		panic("Service version option must be specified")
	}

	if s.Info().Address == nil {
		panic("Service address option must be specified")
	}

	return s
}

// Info gets service information metadata
func (sb ServiceBase) Info() service.Info {
	return sb.options.Info
}

// Metrics gets service metrics reporter
func (sb ServiceBase) Metrics() metrics.Reporter {
	return sb.options.Metrics
}

// Tracer gets service tracer
func (sb ServiceBase) Tracer() trace.Tracer {
	return sb.options.Tracer
}

// Options gets service options
func (sb ServiceBase) Options() Options {
	return sb.options
}

// Log gets service logger instance
func (sb ServiceBase) Log() logger.Logger {
	return sb.options.Logger
}

// Discovery gets service discovery instance for service
func (sb ServiceBase) Discovery() discovery.ServiceDiscovery {
	return sb.options.Discovery
}

// Broker gets message broker mechanism
func (sb ServiceBase) Broker() broker.MessageBroker {
	return sb.options.Broker
}

// Dispose disposes service instance
func (sb ServiceBase) Dispose() {
	sb.Log().Info("Disposing service")

	if sb.Broker() != nil {
		sb.Broker().Dispose()
	}

	if sb.Metrics() != nil {
		sb.Metrics().Dispose()
	}

	if sb.Discovery() != nil {
		sb.Discovery().Dispose()
	}
}
