package discovery

import (
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/system"
)

// ServiceDiscovery represents service registration and localization mechanism
type ServiceDiscovery interface {
	RegisterService(options ...Option) error
	DeregisterService(options ...Option) error
	GetServiceAddress(options ...Option) (service.Address, error)

	system.Disposer
}
