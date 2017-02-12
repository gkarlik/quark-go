package noop_test

import (
	"testing"

	"github.com/gkarlik/quark-go/service/trace/noop"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNoopTracer(t *testing.T) {
	tracer := noop.NewTracer()
	defer tracer.Dispose()

	s := tracer.StartSpan("Test")
	s.SetTag("tag", 1)
	s.Log("TestEvent")
	s.LogWithFields("TestEventWithFields", map[string]interface{}{
		"key": 2,
	})
	s.Finish()

	ctx := tracer.ContextWithSpan(context.Background(), s)
	assert.NotNil(t, ctx, "ContextWithSpan returned nil")

	span, err := tracer.ExtractSpan("TestExtractOperation", "TestExtractFormat", "TestExtractCarrier")
	assert.NotNil(t, span, "Extracted span is nil")
	assert.NoError(t, err, "Unexpected error")

	err = tracer.InjectSpan(s, "TestInjectFormat", "TestInjectCarrier")
	assert.NoError(t, err, "Unexpected error")

	span = tracer.SpanFromContext(context.Background())
	assert.NotNil(t, span, "Span from context is nil")

	span, ctx = tracer.StartSpanFromContext("TestSpanFromContext", context.Background())
	assert.NotNil(t, span, "Span from context is nil")
	assert.NotNil(t, ctx, "Context is nil")

	span = tracer.StartSpanWithParent("TestSpanWithParent", s)
	assert.NotNil(t, span, "Span with parent is nil")
}
