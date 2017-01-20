package quark_test

import (
	"github.com/gkarlik/quark"
	"github.com/gkarlik/quark/broker"
	"github.com/gkarlik/quark/logger/logrus"
	"github.com/gkarlik/quark/metrics"
	"github.com/gkarlik/quark/service"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestServiceDiscovery struct{}

func (sd *TestServiceDiscovery) RegisterService(options ...discovery.Option) error {
	return nil
}

func (sd *TestServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	return nil
}

func (sd *TestServiceDiscovery) GetServiceAddress(options ...discovery.Option) (service.Address, error) {
	return nil, nil
}

func (sd *TestServiceDiscovery) Dispose() {}

type TestBroker struct{}

func (b *TestBroker) PublishMessage(message broker.Message) error {
	return nil
}

func (b *TestBroker) Subscribe(key string) (<-chan broker.Message, error) {
	return nil, nil
}

func (b *TestBroker) Dispose() {}

type TestService struct {
	*quark.ServiceBase
}

type TestMetrics struct{}

func (m *TestMetrics) Report(ms []metrics.Metric) error {
	return nil
}

func (m *TestMetrics) Dispose() {}

func TestServiceBase(t *testing.T) {
	discovery := &TestServiceDiscovery{}
	logger := logrus.NewLogger()
	broker := &TestBroker{}
	metrics := &TestMetrics{}

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Port(5678),
			quark.Logger(logger),
			quark.Discovery(discovery),
			quark.Broker(broker),
			quark.Metrics(metrics)),
	}

	defer ts.Dispose()

	assert.Equal(t, "TestService", ts.Info().Name)
	assert.Equal(t, ts.Info(), ts.Options().Info)
	assert.Equal(t, 5678, ts.Info().Port)
	assert.Equal(t, "1.0", ts.Info().Version)
	assert.Equal(t, 2, len(ts.Info().Tags))
	assert.Equal(t, "A", ts.Info().Tags[0])
	assert.Equal(t, "B", ts.Info().Tags[1])
	assert.Equal(t, logger, ts.Log())
	assert.Equal(t, logger, ts.Options().Logger)
	assert.Equal(t, discovery, ts.Discovery())
	assert.Equal(t, discovery, ts.Options().Discovery)
	assert.Equal(t, broker, ts.Broker())
	assert.Equal(t, broker, ts.Options().Broker)
	assert.Equal(t, metrics, ts.Metrics())
	assert.Equal(t, metrics, ts.Options().Metrics)

	addr, _ := ts.GetHostAddress()
	// address will change on CI server
	assert.Equal(t, "192.168.1.107:5678", addr.String())
}

func TestLackOfLogger(t *testing.T) {
	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Name("TestService"),
				quark.Version("1.0"),
				quark.Tags("A", "B"),
				quark.Port(5678)),
		}
	})
}

func TestLackOfName(t *testing.T) {
	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Version("1.0"),
				quark.Tags("A", "B"),
				quark.Port(5678),
				quark.Logger(logrus.NewLogger())),
		}
	})
}

func TestLackOfVersion(t *testing.T) {
	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Name("TestService"),
				quark.Tags("A", "B"),
				quark.Port(5678),
				quark.Logger(logrus.NewLogger())),
		}
	})
}
