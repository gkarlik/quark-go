package consul_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gkarlik/quark-go/service"
	"github.com/gkarlik/quark-go/service/discovery"
	"github.com/gkarlik/quark-go/service/discovery/consul"
	"github.com/gkarlik/quark-go/service/loadbalancer/random"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
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
		Client: c,
	}
}

func TestNewServiceDiscovery(t *testing.T) {
	c := consul.NewServiceDiscovery("http://consul:8080/")
	defer c.Dispose()

	config := reflect.Indirect(reflect.ValueOf(c.Client)).FieldByName("config")
	addr := config.FieldByName("Address")

	assert.Equal(t, "http://consul:8080/", addr.String())
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
	assert.Equal(t, "/v1/agent/service/deregister/ServiceID", m.Request.URL.Path)
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
	assert.Equal(t, "/v1/health/service/ServiceID", m.Request.URL.Path)
	assert.Equal(t, tag, m.Request.URL.Query()["tag"][0])
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
	assert.Equal(t, "/v1/health/service/ServiceID", m.Request.URL.Path)
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
	assert.Equal(t, "/v1/health/service/ServiceID", m.Request.URL.Path)
	assert.Equal(t, tag, m.Request.URL.Query()["tag"][0])
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
	assert.Equal(t, "/v1/health/service/ServiceID", m.Request.URL.Path)
	assert.Equal(t, tag, m.Request.URL.Query()["tag"][0])
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
	assert.Equal(t, "/v1/health/service/ServiceID", m.Request.URL.Path)
	assert.Equal(t, tag, m.Request.URL.Query()["tag"][0])
}

func prepareResponse(code int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)

	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewBufferString(string(b))),
	}
}
