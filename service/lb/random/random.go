package random

import (
	"errors"
	"github.com/gkarlik/quark/service"
	"math/rand"
	"time"
)

type loadBalancingStrategy struct {
}

// NewRandomLBStrategy creates random load balancing mechanism
func NewRandomLBStrategy() *loadBalancingStrategy {
	return &loadBalancingStrategy{}
}

func (s loadBalancingStrategy) PickServiceAddress(sa []service.Address) (service.Address, error) {
	if len(sa) == 0 {
		return nil, errors.New("Registration list is empty")
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(sa))

	return sa[i], nil
}
