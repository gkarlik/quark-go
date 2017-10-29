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

// Service represents service instance.
type Service interface {
	Info() service.Info                    // service metadata
	Options() Options                      // service options
	Log() logger.Logger                    // service logger
	Discovery() discovery.ServiceDiscovery // service discovery
	Broker() broker.MessageBroker          // service message broker
	Metrics() metrics.Exposer              // service metrics collector
	Tracer() trace.Tracer                  // service request tracer

	system.Disposer
}

// RPCService represents service which exposes procedures that could be called remotely.
type RPCService interface {
	Service

	RegisterServiceInstance(server interface{}, serviceInstance interface{}) error
}

// ServiceBase is base structure for custom service.
type ServiceBase struct {
	options Options // options
}

// NewService creates instance of service with options passed as arguments.
// Service name, version and address are required options.
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

// Info gets service metadata information.
func (sb ServiceBase) Info() service.Info {
	return sb.options.Info
}

// Metrics gets service metrics collector instance.
func (sb ServiceBase) Metrics() metrics.Exposer {
	return sb.options.Metrics
}

// Tracer gets service request tracer instance.
func (sb ServiceBase) Tracer() trace.Tracer {
	return sb.options.Tracer
}

// Options gets service options.
func (sb ServiceBase) Options() Options {
	return sb.options
}

// Log gets service logger instance.
func (sb ServiceBase) Log() logger.Logger {
	return sb.options.Logger
}

// Discovery gets service discovery instance.
func (sb ServiceBase) Discovery() discovery.ServiceDiscovery {
	return sb.options.Discovery
}

// Broker gets message broker instance.
func (sb ServiceBase) Broker() broker.MessageBroker {
	return sb.options.Broker
}

// Dispose cleans up service instance.
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

	if sb.Tracer() != nil {
		sb.Tracer().Dispose()
	}
}
