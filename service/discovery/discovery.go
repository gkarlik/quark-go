package discovery

import (
	"github.com/gkarlik/quark-go/system"
	"net/url"
)

// ServiceDiscovery represents service registration and localization mechanism
type ServiceDiscovery interface {
	RegisterService(options ...Option) error
	DeregisterService(options ...Option) error
	GetServiceAddress(options ...Option) (*url.URL, error)

	system.Disposer
}
