package zipkin

import (
	"net/url"

	cb "github.com/gkarlik/quark/circuitbreaker"
	"github.com/gkarlik/quark/service/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"golang.org/x/net/context"
)

// Span represents tracing span based on opentracing zipkin framework
type Span struct {
	RawSpan opentracing.Span
}

// LogWithFields logs span with event name and fields
func (s Span) LogWithFields(event string, params map[string]interface{}) {
	s.RawSpan.LogEvent(event)

	fields := []log.Field{}
	for k, v := range params {
		f := log.Object(k, v)
		fields = append(fields, f)
	}
	s.RawSpan.LogFields(fields...)
}

// Log logs span event
func (s Span) Log(event string) {
	s.RawSpan.LogEvent(event)
}

// SetTag sets tag on span
func (s Span) SetTag(key string, value interface{}) {
	s.RawSpan.SetTag(key, value)
}

// Finish stops tracing span
func (s Span) Finish() {
	s.RawSpan.Finish()
}

// Tracer represents tracing mechanism based on opentracing zipkin framework
type Tracer struct {
	collector zipkin.Collector
}

// NewTracer creates an instance of tracer based on opentracing zipkin framework. Panics if cannot connect to collector or cannot create zipkin instance.
func NewTracer(address string, serviceName string, serviceAddress *url.URL, opts ...cb.Option) trace.Tracer {
	collector, err := new(cb.DefaultCircuitBreaker).Execute(func() (interface{}, error) {
		return zipkin.NewHTTPCollector(address)
	}, opts...)

	if err != nil {
		panic("Cannot connect to zipkin collector")
	}

	c := collector.(zipkin.Collector)

	recorder := zipkin.NewRecorder(c, false, serviceAddress.String(), serviceName)

	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
	)

	if err != nil {
		panic("Cannot create zipkin tracer")
	}

	opentracing.SetGlobalTracer(tracer)

	return &Tracer{
		collector: c,
	}
}

// StartSpan starts span with name
func (t Tracer) StartSpan(name string) trace.Span {
	s := opentracing.StartSpan(name)

	return &Span{
		RawSpan: s,
	}
}

// StartSpanFromContext starts span from context with name
func (t Tracer) StartSpanFromContext(name string, ctx context.Context) (trace.Span, context.Context) {
	s, c := opentracing.StartSpanFromContext(ctx, name)

	return &Span{
		RawSpan: s,
	}, c
}

// StartSpanWithParent starts span with parent span and name
func (t Tracer) StartSpanWithParent(name string, parent trace.Span) trace.Span {
	ps := assertSpanType(parent)
	s := opentracing.StartSpan(name, opentracing.ChildOf(ps.RawSpan.Context()))

	return &Span{
		RawSpan: s,
	}
}

// SpanFromContext creates span from context
func (t Tracer) SpanFromContext(ctx context.Context) trace.Span {
	s := opentracing.SpanFromContext(ctx)

	return &Span{
		RawSpan: s,
	}
}

// ContextWithSpan creates context with span
func (t Tracer) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	s := assertSpanType(span)

	return opentracing.ContextWithSpan(ctx, s.RawSpan)
}

// InjectSpan injects span in particular format to carrier
func (t Tracer) InjectSpan(s trace.Span, format interface{}, carrier interface{}) error {
	span := assertSpanType(s)
	tracer := opentracing.GlobalTracer()

	return tracer.Inject(span.RawSpan.Context(), format, carrier)
}

// ExtractSpan extracts span in particular format from carrier and starts it with name
func (t Tracer) ExtractSpan(name string, format interface{}, carrier interface{}) (trace.Span, error) {
	tracer := opentracing.GlobalTracer()
	ctx, err := tracer.Extract(format, carrier)
	s := opentracing.StartSpan(name, ext.RPCServerOption(ctx))

	return &Span{
		RawSpan: s,
	}, err
}

// Dispose cleans up tracer instance
func (t Tracer) Dispose() {
	if t.collector != nil {
		t.collector.Close()
	}
}

func assertSpanType(s trace.Span) *Span {
	span, ok := s.(*Span)
	if !ok {
		panic("Incorrect type of span")
	}
	return span
}
