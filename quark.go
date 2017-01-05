package quark

import (
	"carlos/quark/service"
	"carlos/quark/service/bus"
	"carlos/quark/service/discovery"
	"carlos/quark/service/log"
	"carlos/quark/service/metrics"
	"errors"
	"fmt"
	"net"
)

// Service represents service instance
type Service interface {
	Info() service.Info
	Options() Options
	Log() log.Logger
	Discovery() discovery.ServiceDiscovery
	Bus() bus.ServiceBus
	Metrics() metrics.Reporter

	GetHostAddress() (service.Address, error)

	service.Disposer
}

// RPCService represents service which exposes procedures that could be called remotelly
type RPCService interface {
	Service

	RegisterServiceInstance(server interface{}, serviceInstance interface{}) error
}

// ServiceBase is base structure for custom service
type ServiceBase struct {
	options Options
	info    service.Info
}

// NewService creates instance of service
func NewService(opts ...Option) *ServiceBase {
	s := &ServiceBase{
		options: Options{
			Info: service.Info{},
		},
	}

	for _, opt := range opts {
		opt(&s.options)
	}

	return s
}

// Info gets service information metadata
func (sb ServiceBase) Info() service.Info {
	return sb.info
}

// Metrics gets service metrics reporter
func (sb ServiceBase) Metrics() metrics.Reporter {
	return sb.options.Metrics
}

// Options gets service options
func (sb ServiceBase) Options() Options {
	return sb.options
}

// Log gets service logger instance
func (sb ServiceBase) Log() log.Logger {
	return sb.options.Logger
}

// Discovery gets service discovery instance for service
func (sb ServiceBase) Discovery() discovery.ServiceDiscovery {
	return sb.options.Discovery
}

// Bus gets service message bus mechanism
func (sb ServiceBase) Bus() bus.ServiceBus {
	return sb.options.Bus
}

// GetHostAddress gets address on which service is hosted
func (sb ServiceBase) GetHostAddress() (service.Address, error) {
	ip, err := sb.getLocalIPAddress()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s:%d", ip, sb.Options().Info.Port)

	return service.NewURIServiceAddress(url), nil
}

// Dispose disposes service instance
func (sb ServiceBase) Dispose() {
	if sb.Bus() != nil {
		sb.Bus().Dispose()
	}

	if sb.Metrics() != nil {
		sb.Metrics().Dispose()
	}

	if sb.Discovery() != nil {
		sb.Discovery().Dispose()
	}
}

func (sb ServiceBase) getLocalIPAddress() (string, error) {
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
