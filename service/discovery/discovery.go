package discovery

import (
	"github.com/gkarlik/quark/service"
)

// ServiceDiscovery represents service registration and localization mechanism
type ServiceDiscovery interface {
	RegisterService(options ...Option) error
	DeregisterService(options ...Option) error
	GetServiceAddress(options ...Option) (service.Address, error)

	service.Disposer
}
