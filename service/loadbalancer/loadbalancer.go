package loadbalancer

import (
	"net/url"
)

// LoadBalancingStrategy represents Load Balancing mechanism
type LoadBalancingStrategy interface {
	PickServiceAddress(sa []*url.URL) (*url.URL, error)
}
