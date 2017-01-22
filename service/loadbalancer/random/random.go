package random

import (
	"errors"
	"math/rand"
	"net/url"
	"time"
)

// LoadBalancingStrategy represents Random Load Balancing mechanism
type LoadBalancingStrategy struct {
	Randomizer *rand.Rand
}

// NewRandomLBStrategy creates random load balancing mechanism
func NewRandomLBStrategy() *LoadBalancingStrategy {
	return &LoadBalancingStrategy{
		Randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// PickServiceAddress randomly picks service address from list of adresses
func (s LoadBalancingStrategy) PickServiceAddress(sa []*url.URL) (*url.URL, error) {
	l := len(sa)
	if l == 0 {
		return nil, errors.New("Registration list is empty")
	}

	i := s.Randomizer.Intn(l)

	return sa[i], nil
}
