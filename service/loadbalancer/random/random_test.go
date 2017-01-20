package random_test

import (
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/loadbalancer/random"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestNewRandomLoadBalancer(t *testing.T) {
	lb := random.NewRandomLBStrategy()

	if lb == nil {
		t.Errorf("NewRandomLBStrategy return nil, want instance")
	}
}

func TestRandomLoadBalancer(t *testing.T) {
	addr1 := service.NewURIServiceAddress("uri 1")
	addr2 := service.NewURIServiceAddress("uri 2")
	addr3 := service.NewURIServiceAddress("uri 3")
	addr4 := service.NewURIServiceAddress("uri 4")

	var cases = []struct {
		in   []service.Address
		want service.Address
	}{
		{nil, nil},
		{[]service.Address{}, nil},
		{[]service.Address{addr1}, addr1},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr4},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr3},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr3},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr3},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr4},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr2},
		{[]service.Address{addr1, addr2, addr3, addr4}, addr1},
	}

	lb := &random.LoadBalancingStrategy{
		// fixed seed value (99) to get the same "random" results - 4, 3, 3, 3, 4, 2, 1, ...
		Randomizer: rand.New(rand.NewSource(99)),
	}

	for _, c := range cases {
		got, _ := lb.PickServiceAddress(c.in)

		assert.Equal(t, c.want, got)
	}
}
