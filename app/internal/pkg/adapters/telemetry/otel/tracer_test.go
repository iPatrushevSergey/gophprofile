package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	otelprop "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTracer_ContextToMapMapToContext_roundTrip(t *testing.T) {
	otel.SetTextMapPropagator(otelprop.NewCompositeTextMapPropagator(
		otelprop.TraceContext{},
		otelprop.Baggage{},
	))

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "upload")
	defer span.End()

	propagator := NewTracer()
	carrier := propagator.ContextToMap(ctx)
	require.NotEmpty(t, carrier)

	restored := propagator.MapToContext(context.Background(), carrier)
	assert.Equal(t, span.SpanContext().TraceID(), oteltrace.SpanContextFromContext(restored).TraceID())
}

func TestVersion_defaultsWhenBuildVersionUnset(t *testing.T) {
	assert.Equal(t, "0.0.0", Version())
}
