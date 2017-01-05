package consul

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/hashicorp/consul/api"
)

type serviceDiscovery struct {
	Address string
	Client  *api.Client
}

// NewServiceDiscovery creates service registration and localization based on Consul by Hashicorp
func NewServiceDiscovery(address string) *serviceDiscovery {
	c, err := api.NewClient(&api.Config{
		Address: address,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"address": address,
		}).Error("Cannot connect to Consul service")
	}

	return &serviceDiscovery{
		Address: address,
		Client:  c,
	}
}

func (c serviceDiscovery) RegisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	return c.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      opts.Info.Name,
		Name:    opts.Info.Name,
		Tags:    opts.Info.Tags,
		Port:    opts.Info.Port,
		Address: opts.Address.String(),
	})
}

func (c serviceDiscovery) DeregisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	return c.Client.Agent().ServiceDeregister(opts.Info.Name)
}

func (c serviceDiscovery) GetServiceAddress(options ...discovery.Option) (service.Address, error) {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	services, _, err := c.Client.Health().Service(opts.Info.Name, opts.Info.Tags[0], true, nil)

	if err != nil {
		return nil, err
	}

	srvs := make([]service.Address, 0, len(services))
	for _, s := range services {
		srvs = append(srvs, service.NewURIServiceAddress(s.Service.Address))
	}

	if len(srvs) == 0 {
		return nil, nil
	}

	if opts.Strategy == nil {
		return srvs[0], nil
	}

	return opts.Strategy.PickServiceAddress(srvs)
}

func (c serviceDiscovery) Dispose() {
	if c.Client != nil {
		c.Client = nil
	}
}
