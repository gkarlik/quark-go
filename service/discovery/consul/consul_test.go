package consul_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/gkarlik/quark/service/discovery/consul"
	"github.com/gkarlik/quark/service/loadbalancer/random"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type HttpTransportMock struct {
	Request  *http.Request
	Response *http.Response
	Error    error
}

func (m *HttpTransportMock) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// remember request
	m.Request = req

	// return specified response or error
	return m.Response, m.Error
}

func NewConsulClient(t *HttpTransportMock) *consul.ServiceDiscovery {
	addr := "consul"

	c, _ := api.NewClient(&api.Config{
		Address: addr,
		HttpClient: &http.Client{
			Transport: t,
		},
	})

	return &consul.ServiceDiscovery{
		Address: addr,
		Client:  c,
	}
}

func TestNewServiceDiscovery(t *testing.T) {
	c := consul.NewServiceDiscovery("http://consul:8080/")
	defer c.Dispose()

	assert.Equal(t, "http://consul:8080/", c.Address)
}

func TestRegisterService(t *testing.T) {
	addr := "http://server/service"
	url, _ := url.Parse(addr)

	info := service.Info{
		Name:    "ServiceName",
		Address: url,
		Tags:    []string{"A", "B"},
		Version: "1.0",
	}

	m := &HttpTransportMock{}
	m.Response = prepareResponse(http.StatusOK, "OK")

	c := NewConsulClient(m)
	err := c.RegisterService(discovery.WithInfo(info))

	assert.NoError(t, err, "RegisterService returns an error")

	sr := &api.AgentServiceRegistration{}
	b, _ := ioutil.ReadAll(m.Request.Body)
	json.Unmarshal(b, sr)

	assert.Equal(t, addr, sr.Address)
	assert.Equal(t, info.Name, sr.ID)
	assert.Equal(t, info.Address.String(), sr.Address)
	assert.Equal(t, info.Name, sr.Name)
	assert.Equal(t, info.Tags, sr.Tags)
}

func TestDeregisterService(t *testing.T) {
	name := "ServiceID"

	m := &HttpTransportMock{}
	m.Response = prepareResponse(http.StatusOK, "OK")

	c := NewConsulClient(m)
	err := c.DeregisterService(discovery.ByName(name))

	assert.NoError(t, err, "RegisterService returns an error")

	// last segment in request url is id of the service
	s := strings.Split(m.Request.URL.Path, "/")
	id := s[len(s)-1]

	assert.Equal(t, name, id)
}

func TestGetServiceAddress(t *testing.T) {
	name := "ServiceID"
	tag := "A"
	addr := "http://server/service"

	m := &HttpTransportMock{}

	services := make([]*api.ServiceEntry, 0)

	s := &api.ServiceEntry{
		Checks: nil,
		Node:   nil,
		Service: &api.AgentService{
			Address: addr,
			ID:      name,
			Service: name,
			Tags:    []string{tag},
		},
	}

	services = append(services, s)
	m.Response = prepareResponse(http.StatusOK, services)

	c := NewConsulClient(m)
	a, err := c.GetServiceAddress(
		discovery.ByName(name),
		discovery.ByTag(tag),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()))

	assert.NoError(t, err, "RegisterService returns an error")
	assert.Equal(t, addr, a.String())

	// last segment in request url is id of the service
	u := strings.Split(m.Request.URL.Path, "/")
	id := u[len(u)-1]

	// skip query parameters
	u = strings.Split(id, "?")
	assert.Equal(t, name, u[0])

	p := m.Request.URL.Query()
	assert.Equal(t, tag, p["tag"][0])
}

func TestGetServiceAddressWithoutTag(t *testing.T) {
	name := "ServiceID"
	addr := "http://server/service"

	m := &HttpTransportMock{}

	services := make([]*api.ServiceEntry, 0)

	s := &api.ServiceEntry{
		Checks: nil,
		Node:   nil,
		Service: &api.AgentService{
			Address: addr,
			ID:      name,
			Service: name,
			Tags:    nil,
		},
	}

	services = append(services, s)
	m.Response = prepareResponse(http.StatusOK, services)

	c := NewConsulClient(m)
	a, err := c.GetServiceAddress(
		discovery.ByName(name),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()))

	assert.NoError(t, err, "RegisterService returns an error")
	assert.Equal(t, addr, a.String())

	// last segment in request url is id of the service
	u := strings.Split(m.Request.URL.Path, "/")
	id := u[len(u)-1]

	// skip query parameters
	u = strings.Split(id, "?")
	assert.Equal(t, name, u[0])
}

func TestGetServiceAddressMissingLBStrategy(t *testing.T) {
	name := "ServiceID"
	tag := "A"
	addr := "http://server/service"

	m := &HttpTransportMock{}

	services := make([]*api.ServiceEntry, 0)

	s := &api.ServiceEntry{
		Checks: nil,
		Node:   nil,
		Service: &api.AgentService{
			Address: addr,
			ID:      name,
			Service: name,
			Tags:    []string{tag},
		},
	}

	services = append(services, s)
	m.Response = prepareResponse(http.StatusOK, services)

	c := NewConsulClient(m)
	a, err := c.GetServiceAddress(
		discovery.ByName(name),
		discovery.ByTag(tag))

	assert.NoError(t, err, "RegisterService returns an error")
	assert.Equal(t, addr, a.String())

	// last segment in request url is id of the service
	u := strings.Split(m.Request.URL.Path, "/")
	id := u[len(u)-1]

	// skip query parameters
	u = strings.Split(id, "?")
	assert.Equal(t, name, u[0])

	p := m.Request.URL.Query()
	assert.Equal(t, tag, p["tag"][0])
}

func TestGetServiceAddressEmptyServicesList(t *testing.T) {
	name := "ServiceID"
	tag := "A"

	m := &HttpTransportMock{}

	services := make([]*api.ServiceEntry, 0)
	m.Response = prepareResponse(http.StatusOK, services)

	c := NewConsulClient(m)
	a, err := c.GetServiceAddress(
		discovery.ByName(name),
		discovery.ByTag(tag),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()))

	assert.NoError(t, err, "RegisterService returns an error")
	assert.Nil(t, a, "Result should be nil")

	// last segment in request url is id of the service
	u := strings.Split(m.Request.URL.Path, "/")
	id := u[len(u)-1]

	// skip query parameters
	u = strings.Split(id, "?")
	assert.Equal(t, name, u[0])

	p := m.Request.URL.Query()
	assert.Equal(t, tag, p["tag"][0])
}

func TestGetServiceAddressError(t *testing.T) {
	name := "ServiceID"
	tag := "A"

	m := &HttpTransportMock{}

	m.Response = nil
	m.Error = errors.New("Network problems")

	c := NewConsulClient(m)
	a, err := c.GetServiceAddress(
		discovery.ByName(name),
		discovery.ByTag(tag),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()))

	assert.Error(t, err, "GetServiceAddress should return an error")
	assert.Nil(t, a, "Result should be nil")

	// last segment in request url is id of the service
	u := strings.Split(m.Request.URL.Path, "/")
	id := u[len(u)-1]

	// skip query parameters
	u = strings.Split(id, "?")
	assert.Equal(t, name, u[0])

	p := m.Request.URL.Query()
	assert.Equal(t, tag, p["tag"][0])
}

func prepareResponse(code int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)

	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewBufferString(string(b))),
	}
}
