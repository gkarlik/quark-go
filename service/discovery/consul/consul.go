package consul

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/hashicorp/consul/api"
	"net"
	"net/url"
	"strconv"
)

// ServiceDiscovery represents service discovery mechanism based on Consul by Hashicorp
type ServiceDiscovery struct {
	Address string
	Client  *api.Client
}

// NewServiceDiscovery creates service registration and localization based on Consul by Hashicorp. Panics if cannot create an instance
func NewServiceDiscovery(address string) *ServiceDiscovery {
	c, err := api.NewClient(&api.Config{
		Address: address,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"address": address,
		}).Panic("Cannot connect to Consul service")
	}

	return &ServiceDiscovery{
		Address: address,
		Client:  c,
	}
}

// RegisterService registers service in service discovery catalog
func (c ServiceDiscovery) RegisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	_, port, _ := net.SplitHostPort(opts.Info.Address.Host)

	p, _ := strconv.Atoi(port)

	return c.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      opts.Info.Name,
		Name:    opts.Info.Name,
		Tags:    opts.Info.Tags,
		Port:    p,
		Address: opts.Address.String(),
	})
}

// DeregisterService unregisters service in service discovery catalog
func (c ServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	return c.Client.Agent().ServiceDeregister(opts.Info.Name)
}

// GetServiceAddress gets service address from service discovery catalog
func (c ServiceDiscovery) GetServiceAddress(options ...discovery.Option) (*url.URL, error) {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	services, _, err := c.Client.Health().Service(opts.Info.Name, opts.Info.Tags[0], true, nil)

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
		return srvs[0], nil
	}

	return opts.Strategy.PickServiceAddress(srvs)
}

// Dispose cleans up ServiceDiscovery instance
func (c ServiceDiscovery) Dispose() {
	if c.Client != nil {
		c.Client = nil
	}
}
