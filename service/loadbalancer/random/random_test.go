package random_test

import (
	"github.com/gkarlik/quark-go/service/loadbalancer/random"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/url"
	"testing"
)

func TestNewRandomLoadBalancer(t *testing.T) {
	lb := random.NewRandomLBStrategy()

	if lb == nil {
		t.Errorf("NewRandomLBStrategy return nil, want instance")
	}
}

func TestRandomLoadBalancer(t *testing.T) {
	addr1, _ := url.Parse("http://server/url1")
	addr2, _ := url.Parse("http://server/url2")
	addr3, _ := url.Parse("http://server/url3")
	addr4, _ := url.Parse("http://server/url4")

	var cases = []struct {
		in   []*url.URL
		want *url.URL
	}{
		{nil, nil},
		{[]*url.URL{}, nil},
		{[]*url.URL{addr1}, addr1},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr4},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr3},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr3},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr3},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr4},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr2},
		{[]*url.URL{addr1, addr2, addr3, addr4}, addr1},
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
