package zipkin_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gkarlik/quark/service/trace/zipkin"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestTracer(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))
	defer ts.Close()

	url, _ := url.Parse("http://server/service")
	tracer := zipkin.NewTracer(ts.URL, "Test", url)

	span := tracer.StartSpan("root_span")
	span.Log("test_event")
	span.Finish()

	tracer.Dispose()

	assert.Contains(t, data.body, "root_span")
	assert.Contains(t, data.body, "test_event")
	assert.Contains(t, data.body, "Test")
}

func TestTracerSpan(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))
	defer ts.Close()

	url, _ := url.Parse("http://server/service")
	tracer := zipkin.NewTracer(ts.URL, "Test", url)

	span := tracer.StartSpan("root_span")
	span.Log("test_event")
	span.SetTag("error", "this_is_an_error_message")
	span.LogWithFields("new_test_event", map[string]interface{}{
		"a_field": "a_field_value",
		"b_field": "b_field_value",
		"c_field": 1234567,
	})

	child := tracer.StartSpanWithParent("child_span", span)
	child.Log("child_span_log")
	child.Finish()

	span.Finish()

	tracer.Dispose()

	assert.Contains(t, data.body, "root_span")
	assert.Contains(t, data.body, "test_event")
	assert.Contains(t, data.body, "this_is_an_error_message")
	assert.Contains(t, data.body, "new_test_event")
	assert.Contains(t, data.body, "a_field=a_field_value")
	assert.Contains(t, data.body, "b_field=b_field_value")
	assert.Contains(t, data.body, "c_field=1234567")
	assert.Contains(t, data.body, "Test")
	assert.Contains(t, data.body, "child_span")
	assert.Contains(t, data.body, "child_span_log")
}

func TestTracerSpanFromContext(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))
	defer ts.Close()

	url, _ := url.Parse("http://server/service")
	tracer := zipkin.NewTracer(ts.URL, "Test", url)

	span := tracer.StartSpan("root_span")

	ctx := tracer.ContextWithSpan(context.Background(), span)
	s := tracer.SpanFromContext(ctx)

	cs, _ := tracer.StartSpanFromContext("span_from_context", ctx)
	cs.Finish()

	s.Finish()

	tracer.Dispose()

	assert.Contains(t, data.body, "root_span")
	assert.Contains(t, data.body, "Test")
	assert.Contains(t, data.body, "span_from_context")
}

func TestTracerSpanIncorrectType(t *testing.T) {
	url, _ := url.Parse("http://server/service")
	tracer := zipkin.NewTracer(url.String(), "Test", url)

	span := tracer.StartSpan("root_span")

	s, _ := span.(*zipkin.Span)

	assert.Panics(t, func() {
		child := tracer.StartSpanWithParent("child_span", *s)
		child.Finish()
	})

	span.Finish()

	tracer.Dispose()
}

func TestTracerInjectExtract(t *testing.T) {
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

	url, _ := url.Parse("http://server/service")
	tracer := zipkin.NewTracer(ts.URL, "Test", url)

	// service http server
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		span, err := tracer.ExtractSpan("extracted_span", opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		assert.NoError(t, err, "ExtractSpan return an error")

		span.Finish()
	}))
	defer func() {
		hs.Close()
		ts.Close()
	}()

	span := tracer.StartSpan("root_span")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", hs.URL, nil)

	err := tracer.InjectSpan(span, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	assert.NoError(t, err, "InjectSpan returns an error")

	client.Do(req)

	span.Finish()

	tracer.Dispose()

	assert.Contains(t, data.body, "root_span")
	assert.Contains(t, data.body, "extracted_span")
}
