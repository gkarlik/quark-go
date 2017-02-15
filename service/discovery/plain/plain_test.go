package plain_test

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gkarlik/quark-go"
	sd "github.com/gkarlik/quark-go/service/discovery"
	"github.com/gkarlik/quark-go/service/discovery/plain"
	"github.com/gkarlik/quark-go/service/loadbalancer/random"
	"github.com/stretchr/testify/assert"
)

type TestService struct {
	*quark.ServiceBase
}

func TestNewServiceDiscovery(t *testing.T) {
	sd := plain.NewServiceDiscovery(":8080")
	defer sd.Dispose()
}

func TestPlainDiscoveryService(t *testing.T) {
	var wg sync.WaitGroup

	sa, _ := quark.GetHostAddress(1234)
	ha, _ := quark.GetHostAddress(7777)
	discovery := plain.NewServiceDiscovery("http://" + ha.String())

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(sa),
			quark.Discovery(discovery)),
	}
	defer ts.Dispose()

	wg.Add(1)
	discovery.Serve(ha.String())

	go func() {
		err := ts.Discovery().RegisterService(sd.ByInfo(ts.Info()))
		assert.NoError(t, err, "Unexpected error during service registration")

		url, err := ts.Discovery().GetServiceAddress(
			sd.ByName("TestService"),
			sd.ByTag("A"),
			sd.ByVersion("1.0"),
			sd.UsingLBStrategy(random.NewRandomLBStrategy()))

		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Equal(t, sa.String(), url.String())

		url, err = ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByTag("A"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Equal(t, sa.String(), url.String())

		err = ts.Discovery().DeregisterService(sd.ByName("TestService"), sd.ByTag("A"), sd.ByTag("B"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while service deregistration")

		url, err = ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByTag("A"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Nil(t, url, "Url should be nil")

		wg.Done()
	}()

	wg.Wait()
}

func TestPlainDiscoveryServiceTags(t *testing.T) {
	var wg sync.WaitGroup

	sa, _ := quark.GetHostAddress(1234)
	ha, _ := quark.GetHostAddress(7777)
	discovery := plain.NewServiceDiscovery("http://" + ha.String())

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(sa),
			quark.Discovery(discovery)),
	}
	defer ts.Dispose()

	wg.Add(1)
	discovery.Serve(ha.String())

	go func() {
		err := ts.Discovery().RegisterService(sd.ByInfo(ts.Info()))
		assert.NoError(t, err, "Unexpected error during service registration")

		url, err := ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByTag("A"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Equal(t, sa.String(), url.String())

		url, err = ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByTag("C"), sd.ByTag("D"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Nil(t, url, "Url should be nil")

		err = ts.Discovery().DeregisterService(sd.ByName("TestService"), sd.ByTag("C"), sd.ByTag("D"), sd.ByTag("E"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")

		err = ts.Discovery().DeregisterService(sd.ByName("TestService"), sd.ByTag("C"), sd.ByTag("D"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")

		err = ts.Discovery().RegisterService(sd.ByName("TestService"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error during service registration")

		url, err = ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByVersion("1.0"))
		assert.NoError(t, err, "Unexpected error while getting services list")
		assert.Equal(t, "", url.String())

		wg.Done()
	}()

	wg.Wait()
}

func TestPlainDiscoveryServiceIncorrectAddress(t *testing.T) {
	sa, _ := quark.GetHostAddress(1234)
	ha, _ := quark.GetHostAddress(7777)
	discovery := plain.NewServiceDiscovery(ha.String())

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(sa),
			quark.Discovery(discovery)),
	}
	defer ts.Dispose()

	err := ts.Discovery().RegisterService(sd.ByInfo(ts.Info()))
	assert.Error(t, err, "RegisterService should return an error")

	err = ts.Discovery().DeregisterService(sd.ByInfo(ts.Info()))
	assert.Error(t, err, "DeregisterService should return an error")

	_, err = ts.Discovery().GetServiceAddress(sd.ByName("TestService"), sd.ByTag("A"), sd.ByVersion("1.0"))
	assert.Error(t, err, "GetServiceAddress should return an error")
}

func TestPlainDiscoveryServiceDuplicatedEntry(t *testing.T) {
	var wg sync.WaitGroup

	sa, _ := quark.GetHostAddress(1234)
	ha, _ := quark.GetHostAddress(7777)
	discovery := plain.NewServiceDiscovery("http://" + ha.String())

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(sa),
			quark.Discovery(discovery)),
	}
	defer ts.Dispose()

	wg.Add(1)
	discovery.Serve(ha.String())

	go func() {
		err := ts.Discovery().RegisterService(sd.ByInfo(ts.Info()))
		assert.NoError(t, err, "Unexpected error during service registration")

		err = ts.Discovery().RegisterService(sd.ByInfo(ts.Info()))
		assert.NoError(t, err, "Unexpected error during service registration")

		wg.Done()
	}()

	wg.Wait()
}

func TestPlainDiscoveryHandlers(t *testing.T) {
	var wg sync.WaitGroup

	sa, _ := quark.GetHostAddress(1234)
	ha, _ := quark.GetHostAddress(7777)

	addr := "http://" + ha.String()
	discovery := plain.NewServiceDiscovery(addr)

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(sa),
			quark.Discovery(discovery)),
	}
	defer ts.Dispose()

	wg.Add(1)
	discovery.Serve(ha.String())

	go func() {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		r, err := http.NewRequest(http.MethodPost, addr+plain.RegisterServiceURL, bytes.NewBufferString("incorrect payload"))
		assert.NoError(t, err, "Unexpected error during request preparation")

		resp, err := client.Do(r)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err, "Unexpected error during HTTP call")

		r, err = http.NewRequest(http.MethodPost, addr+plain.UnregisterServiceURL, bytes.NewBufferString("incorrect payload"))
		assert.NoError(t, err, "Unexpected error during request preparation")

		resp, err = client.Do(r)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err, "Unexpected error during HTTP call")

		r, err = http.NewRequest(http.MethodPost, addr+plain.ListServicesURL, bytes.NewBufferString("incorrect payload"))
		assert.NoError(t, err, "Unexpected error during request preparation")

		resp, err = client.Do(r)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err, "Unexpected error during HTTP call")

		wg.Done()
	}()

	wg.Wait()
}
