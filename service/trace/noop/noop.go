package noop

import (
	"context"
	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/service/trace"
)

const componentName = "NoopTracer"

// Span represents tracing span based on NOOP request tracing.
type Span struct{}

// SetTag sets tag on tracing span.
func (s *Span) SetTag(key string, value interface{}) {
	logger.Log().InfoWithFields(logger.Fields{
		"key":       key,
		"value":     value,
		"component": componentName,
	}, "Setting tag on span")
}

// Log logs tracing span event.
func (s *Span) Log(event string) {
	logger.Log().InfoWithFields(logger.Fields{
		"event":     event,
		"component": componentName,
	}, "Logging event on span")
}

// LogWithFields logs tracing span with event name and fields.
func (s *Span) LogWithFields(event string, fields map[string]interface{}) {
	logger.Log().InfoWithFields(logger.Fields{
		"event":     event,
		"fields":    fields,
		"component": componentName,
	}, "Logging event with fields on span")
}

// Finish stops tracing span.
func (s *Span) Finish() {
	logger.Log().InfoWithFields(logger.Fields{
		"component": componentName,
	}, "Finishing span")
}

// Tracer represents tracing mechanism based on NOOP request tracing.
type Tracer struct{}

// NewTracer creates an instance of tracer based on NOOP request tracing.
func NewTracer() *Tracer {
	return &Tracer{}
}

// StartSpan starts tracing span with name.
func (t *Tracer) StartSpan(name string) trace.Span {
	logger.Log().InfoWithFields(logger.Fields{
		"name":      name,
		"component": componentName,
	}, "Starting span")

	return &Span{}
}

// StartSpanFromContext starts tracing span with name from context.
func (t *Tracer) StartSpanFromContext(ctx context.Context, name string) (trace.Span, context.Context) {
	logger.Log().InfoWithFields(logger.Fields{
		"name":      name,
		"context":   ctx,
		"component": componentName,
	}, "Starting span from context")

	return &Span{}, ctx
}

// StartSpanWithParent starts tracing span with parent span and name.
func (t *Tracer) StartSpanWithParent(name string, parent trace.Span) trace.Span {
	logger.Log().InfoWithFields(logger.Fields{
		"name":      name,
		"parent":    parent,
		"component": componentName,
	}, "Starting span with parent span")

	return &Span{}
}

// SpanFromContext creates tracing span from context.
func (t *Tracer) SpanFromContext(ctx context.Context) trace.Span {
	logger.Log().InfoWithFields(logger.Fields{
		"context":   ctx,
		"component": componentName,
	}, "Creating span from context")

	return &Span{}
}

// ContextWithSpan creates context with tracing span.
func (t *Tracer) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	logger.Log().InfoWithFields(logger.Fields{
		"context":   ctx,
		"span":      span,
		"component": componentName,
	}, "Creating context with span")

	return ctx
}

// InjectSpan injects tracing span in particular format into carrier.
func (t *Tracer) InjectSpan(s trace.Span, format interface{}, carrier interface{}) error {
	logger.Log().InfoWithFields(logger.Fields{
		"span":      s,
		"format":    format,
		"carrier":   carrier,
		"component": componentName,
	}, "Injecting span in particular format into carrier")

	return nil
}

// ExtractSpan extracts tracing span in particular format from carrier.
func (t *Tracer) ExtractSpan(name string, format interface{}, carrier interface{}) (trace.Span, error) {
	logger.Log().InfoWithFields(logger.Fields{
		"name":      name,
		"format":    format,
		"carrier":   carrier,
		"component": componentName,
	}, "Extracting span in particular format from carrier")

	return &Span{}, nil
}

// Dispose cleans up NOOP tracer instance.
func (t *Tracer) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing tracer component")
}
