package consul

import (
	"net"
	"net/url"
	"strconv"

	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/service/discovery"
	"github.com/hashicorp/consul/api"
)

const componentName = "ConsulServiceDiscovery"

// ServiceDiscovery represents service discovery mechanism based on Consul by Hashicorp.
type ServiceDiscovery struct {
	Client *api.Client // consul client
}

// NewServiceDiscovery creates service registration and localization based on Consul by Hashicorp.
// Panics if cannot create an instance.
func NewServiceDiscovery(address string) *ServiceDiscovery {
	c, err := api.NewClient(&api.Config{
		Address: address,
	})

	if err != nil {
		logger.Log().PanicWithFields(logger.LogFields{
			"error":     err,
			"address":   address,
			"component": componentName,
		}, "Cannot connect to Consul server")
	}

	return &ServiceDiscovery{
		Client: c,
	}
}

// RegisterService registers service in service discovery catalog.
func (c ServiceDiscovery) RegisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	// parsing port from service address
	// trust that url.URL contains host in proper format, so there should not be any errors
	_, port, _ := net.SplitHostPort(opts.Info.Address.Host)
	p, _ := strconv.Atoi(port)

	logger.Log().InfoWithFields(logger.LogFields{
		"ID":        opts.Info.Name,
		"Name":      opts.Info.Name,
		"Tags":      opts.Info.Tags,
		"Port":      p,
		"Address":   opts.Info.Address.String(),
		"component": componentName,
	}, "Registering service in Consul server")

	return c.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      opts.Info.Name,
		Name:    opts.Info.Name,
		Tags:    opts.Info.Tags,
		Port:    p,
		Address: opts.Info.Address.String(),
	})
}

// DeregisterService unregisters service in service discovery catalog.
func (c ServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	logger.Log().InfoWithFields(logger.LogFields{
		"ID":        opts.Info.Name,
		"component": componentName,
	}, "Deregistering service in Consul server")

	return c.Client.Agent().ServiceDeregister(opts.Info.Name)
}

// GetServiceAddress gets service address from service discovery catalog.
func (c ServiceDiscovery) GetServiceAddress(options ...discovery.Option) (*url.URL, error) {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	tag := ""
	if len(opts.Info.Tags) > 0 {
		tag = opts.Info.Tags[0]
	}

	logger.Log().InfoWithFields(logger.LogFields{
		"Name":      opts.Info.Name,
		"Tags":      opts.Info.Tags,
		"component": componentName,
	}, "Getting services list from Consul server")

	services, _, err := c.Client.Health().Service(opts.Info.Name, tag, true, nil)

	if err != nil {
		return nil, err
	}

	srvs := make([]*url.URL, 0, len(services))
	for _, s := range services {
		addr, _ := url.Parse(s.Service.Address)
		srvs = append(srvs, addr)
	}

	if len(srvs) == 0 {
		return nil, nil
	}

	if opts.Strategy == nil {
		logger.Log().DebugWithFields(logger.LogFields{"component": componentName}, "Load balancing strategy is not set. Picking first item from the list.")

		return srvs[0], nil
	}

	logger.Log().InfoWithFields(logger.LogFields{"component": componentName}, "Picking service using load balancing strategy")

	sa, err := opts.Strategy.PickServiceAddress(srvs)
	if sa != nil {
		logger.Log().InfoWithFields(logger.LogFields{"component": componentName, "address": sa.String()})
	}

	return sa, err
}

// Dispose closes consul client and cleans up ServiceDiscovery instance.
func (c ServiceDiscovery) Dispose() {
	logger.Log().InfoWithFields(logger.LogFields{"component": componentName}, "Disposing service discovery component")

	if c.Client != nil {
		c.Client = nil
	}
}
