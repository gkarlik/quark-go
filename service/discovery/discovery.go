package discovery

import (
	"net/url"

	"github.com/gkarlik/quark-go/system"
)

// ServiceDiscovery represents service registration and localization mechanism.
type ServiceDiscovery interface {
	RegisterService(options ...Option) error
	DeregisterService(options ...Option) error
	GetServiceAddress(options ...Option) (*url.URL, error)

	system.Disposer
}
