package lb

import (
	"carlos/quark/service"
)

// LoadBalancingStrategy represents Load Balancing mechanism
type LoadBalancingStrategy interface {
	PickServiceAddress(sa []service.Address) (service.Address, error)
}
