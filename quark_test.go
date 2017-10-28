package quark_test

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"

	"context"

	"github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/broker"
	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/service/discovery"
	"github.com/gkarlik/quark-go/service/trace"
	"github.com/gkarlik/quark-go/service/trace/noop"
	"github.com/gkarlik/quark-go/service/trace/zipkin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type TestServiceDiscovery struct{}

func (sd *TestServiceDiscovery) RegisterService(options ...discovery.Option) error {
	return nil
}

func (sd *TestServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	return nil
}

func (sd *TestServiceDiscovery) GetServiceAddress(options ...discovery.Option) (*url.URL, error) {
	return nil, nil
}

func (sd *TestServiceDiscovery) Dispose() {}

type TestBroker struct{}

func (b *TestBroker) PublishMessage(ctx context.Context, message broker.Message) error {
	return nil
}

func (b *TestBroker) Subscribe(ctx context.Context, key string) (<-chan broker.Message, error) {
	return nil, nil
}

func (b *TestBroker) Dispose() {}

type TestMetrics struct{}

func (m *TestMetrics) Report(ms ...metrics.Metric) error {
	return nil
}

func (m *TestMetrics) Dispose() {}

type TestTracer struct{}

func (t *TestTracer) StartSpan(name string) trace.Span {
	return nil
}

func (t *TestTracer) StartSpanFromContext(name string, ctx context.Context) (trace.Span, context.Context) {
	return nil, nil
}

func (t *TestTracer) StartSpanWithParent(name string, parent trace.Span) trace.Span {
	return nil
}

func (t *TestTracer) SpanFromContext(ctx context.Context) trace.Span {
	return nil
}

func (t *TestTracer) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return nil
}

func (t *TestTracer) InjectSpan(s trace.Span, format interface{}, carrier interface{}) error {
	return nil
}

func (t *TestTracer) ExtractSpan(name string, format interface{}, carrier interface{}) (trace.Span, error) {
	return nil, nil
}

func (t *TestTracer) Dispose() {}

type TestService struct {
	*quark.ServiceBase
}

func TestServiceBase(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	discovery := &TestServiceDiscovery{}
	logger := logger.Log()
	broker := &TestBroker{}
	metrics := &TestMetrics{}
	tracer := &TestTracer{}

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A", "B"),
			quark.Address(a),
			quark.Logger(logger),
			quark.Discovery(discovery),
			quark.Broker(broker),
			quark.Metrics(metrics),
			quark.Tracer(tracer)),
	}

	defer ts.Dispose()

	assert.Equal(t, "TestService", ts.Info().Name)
	assert.Equal(t, ts.Info(), ts.Options().Info)
	assert.Equal(t, a, ts.Info().Address)
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
	assert.Equal(t, tracer, ts.Tracer())
	assert.Equal(t, tracer, ts.Options().Tracer)

	addr, _ := quark.GetHostAddress(5678)
	// address will change on CI server - validate only IP address format
	matched, err := regexp.MatchString("\\d+\\.\\d+\\.\\d+.\\d+:5678", addr.Host)
	assert.NoError(t, err, "Invalid IP address")
	assert.Equal(t, true, matched)

	addr, _ = quark.GetHostAddress(0)
	// address will change on CI server - validate only IP address format
	matched, err = regexp.MatchString("\\d+\\.\\d+\\.\\d+.\\d+", addr.String())
	assert.NoError(t, err, "Invalid IP address")
	assert.Equal(t, true, matched)
}

func TestLackOfName(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Version("1.0"),
				quark.Tags("A", "B"),
				quark.Address(a)),
		}
	})
}

func TestLackOfVersion(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Name("TestService"),
				quark.Tags("A", "B"),
				quark.Address(a)),
		}
	})
}

func TestLackOfAddress(t *testing.T) {
	assert.Panics(t, func() {
		var _ = &TestService{
			ServiceBase: quark.NewService(
				quark.Name("TestService"),
				quark.Version("1.0"),
				quark.Tags("A", "B")),
		}
	})
}

func TestGetEnvVar(t *testing.T) {
	key := "GET_ENV_TEST_VAR"

	os.Setenv(key, "Lorem ipsum")

	v := quark.GetEnvVar(key)

	assert.Equal(t, "Lorem ipsum", v)

	os.Setenv(key, "")

	assert.Panics(t, func() {
		quark.GetEnvVar(key)
	})

	os.Unsetenv(key)

	assert.Panics(t, func() {
		quark.GetEnvVar(key)
	})
}

func TestRPCMetadataCarrier(t *testing.T) {
	md := metadata.Pairs()
	mc := &quark.RPCMetadataCarrier{MD: &md}

	test_key := "test"
	test_val := "test_value"

	mc.Set(test_key, test_val)

	err := mc.ForeachKey(func(key string, value string) error {
		assert.Equal(t, test_key, key)
		assert.Equal(t, test_val, value)

		return nil
	})

	assert.NoError(t, err, "ForeachKey returned an error")
}

func TestRPCMetadataCarrierBinary(t *testing.T) {
	md := metadata.Pairs()
	mc := &quark.RPCMetadataCarrier{MD: &md}

	test_key := "test-bin"
	test_val := "test_value"
	test_val_bin := string(base64.StdEncoding.EncodeToString([]byte(test_val)))

	mc.Set(test_key, test_val)

	err := mc.ForeachKey(func(key string, value string) error {
		assert.Equal(t, test_key, key)
		assert.Equal(t, test_val_bin, value)

		return nil
	})

	assert.NoError(t, err, "ForeachKey returned an error")
}

func TestRPCMetadataCarrierError(t *testing.T) {
	md := metadata.Pairs()
	mc := &quark.RPCMetadataCarrier{MD: &md}

	mc.Set("test_key", "test_val")

	err := mc.ForeachKey(func(key string, value string) error {
		return errors.New("Key not found")
	})

	assert.Error(t, err, "ForeachKey should return an error")
}

type TestMetricRecorder struct {
	Metric metrics.Metric
}

func (sm *TestMetricRecorder) Report(m ...metrics.Metric) error {
	sm.Metric = m[0]

	return nil
}

func (sm *TestMetricRecorder) Dispose() {}

func TestReportServiceValue(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	metrics := &TestMetricRecorder{}

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Metrics(metrics)),
	}

	defer ts.Dispose()

	err := quark.ReportServiceValue(ts, "TestMetric", 1)

	assert.NoError(t, err, "ReportServiceValue returns an error")
	assert.Equal(t, "TestMetric", metrics.Metric.Name)
	assert.Equal(t, "TestService", metrics.Metric.Tags["service"])
	assert.Equal(t, 1, metrics.Metric.Values["value"])
}

func TestCallHTTPService(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	// trace http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))

	a, _ := quark.GetHostAddress(1234)
	tracer := zipkin.NewTracer(ts.URL, "TestService", a)

	s := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Tracer(tracer)),
	}

	// service http server
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		span, err := tracer.ExtractSpan("extracted_span", opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		assert.NoError(t, err, "ExtractSpan return an error")

		span.Finish()
	}))
	defer func() {
		s.Dispose()
		hs.Close()
		ts.Close()
	}()

	span := tracer.StartSpan("root_span")
	_, err := quark.CallHTTPService(s, "GET", hs.URL, nil, span)

	assert.NoError(t, err, "CallHTTPService returns an error")

	span.Finish()

	tracer.Dispose()

	assert.Contains(t, data.body, "root_span")
	assert.Contains(t, data.body, "extracted_span")
}

func TestMessageContextCarrierBinary(t *testing.T) {
	context := broker.MessageContext{}
	mc := &quark.MessageContextCarrier{Context: &context}

	test_key := "test-bin"
	test_val := "test_value"
	test_val_bin := string(base64.StdEncoding.EncodeToString([]byte(test_val)))

	mc.Set(test_key, test_val)

	err := mc.ForeachKey(func(key string, value interface{}) error {
		assert.Equal(t, test_key, key)
		assert.Equal(t, test_val_bin, value)

		return nil
	})

	assert.NoError(t, err, "ForeachKey returned an error")
}

func TestMessageContextCarrierError(t *testing.T) {
	context := broker.MessageContext{}
	mc := &quark.MessageContextCarrier{Context: &context}

	mc.Set("test_key", "test_val")

	err := mc.ForeachKey(func(key string, value interface{}) error {
		return errors.New("Key not found")
	})

	assert.Error(t, err, "ForeachKey should return an error")
}

func TestStartRPCSpan(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	s := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Tracer(noop.NewTracer())),
	}
	defer s.Dispose()

	var md metadata.MD
	ctx := metadata.NewIncomingContext(context.Background(), md)

	span := quark.StartRPCSpan(s, "Test", ctx)
	assert.NotNil(t, span, "Span is not nil")
}

func TestStartRPCSpanIncorrectContext(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	s := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Tracer(noop.NewTracer())),
	}
	defer s.Dispose()

	span := quark.StartRPCSpan(s, "Test", context.Background())
	assert.NotNil(t, span, "Span is not nil")
}

func TestStartMessageSpan(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	s := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Tracer(noop.NewTracer())),
	}
	defer s.Dispose()

	msg := broker.Message{
		Context: broker.MessageContext{},
	}

	span := quark.StartMessageSpan(s, "Test", msg)
	assert.NotNil(t, span, "Span is not nil")
}

func TestReportError(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	metrics := &TestMetricRecorder{}

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Metrics(metrics),
			quark.Tracer(noop.NewTracer())),
	}

	defer ts.Dispose()

	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	quark.ReportError(ts, r, "Error", "errors", "Test error")

	assert.Equal(t, "errors", metrics.Metric.Name)
	assert.Equal(t, "TestService", metrics.Metric.Tags["service"])
	assert.Equal(t, 1, metrics.Metric.Values["value"])
}
