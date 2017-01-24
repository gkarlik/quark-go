package quark

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/gkarlik/quark/broker"
	"github.com/gkarlik/quark/logger"
	"github.com/gkarlik/quark/metrics"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/gkarlik/quark/service/trace"
	"github.com/gkarlik/quark/system"
)

// Service represents service instance
type Service interface {
	Info() service.Info
	Options() Options
	Log() logger.Logger
	Discovery() discovery.ServiceDiscovery
	Broker() broker.MessageBroker
	Metrics() metrics.Reporter

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

// GetEnvVar gets environment variable by key. Panics is variable is not set.
func GetEnvVar(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("Environment variable %q is not set!", key))
	}
	return v
}

// GetHostAddress return host and port address on which service is hosted
func GetHostAddress(port int) (*url.URL, error) {
	ip, err := getLocalIPAddress()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s:%d", ip, port)
	if port == 0 {
		u = fmt.Sprintf(ip)
	}

	return url.Parse(u)
}

func getLocalIPAddress() (string, error) {
	ifaces, error := net.Interfaces()
	if error != nil {
		return "", error
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, error := iface.Addrs()
		if error != nil {
			return "", error
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Network not available")
}
