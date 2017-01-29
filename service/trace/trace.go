package trace

import (
	"github.com/gkarlik/quark-go/system"
	"golang.org/x/net/context"
)

// Tracer represents tracing mechanism
type Tracer interface {
	StartSpan(name string) Span
	StartSpanFromContext(name string, ctx context.Context) (Span, context.Context)
	StartSpanWithParent(name string, parent Span) Span
	SpanFromContext(ctx context.Context) Span
	ContextWithSpan(ctx context.Context, span Span) context.Context
	InjectSpan(s Span, format interface{}, carrier interface{}) error
	ExtractSpan(name string, format interface{}, carrier interface{}) (Span, error)

	system.Disposer
}

// Span represents tracing span
type Span interface {
	SetTag(key string, value interface{})
	Log(event string)
	LogWithFields(event string, fields map[string]interface{})
	Finish()
}
