package otel

import (
	"context"
	"testing"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	otelprop "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestInjectAMQPExtractAMQP_roundTrip(t *testing.T) {
	otel.SetTextMapPropagator(otelprop.NewCompositeTextMapPropagator(
		otelprop.TraceContext{},
		otelprop.Baggage{},
	))

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent")
	defer span.End()

	headers := InjectAMQP(ctx, amqp091.Table{})
	require.NotEmpty(t, headers)

	childCtx := ExtractAMQP(context.Background(), headers)
	childSpanCtx := oteltrace.SpanContextFromContext(childCtx)
	parentSpanCtx := span.SpanContext()

	assert.True(t, childSpanCtx.IsValid())
	assert.Equal(t, parentSpanCtx.TraceID(), childSpanCtx.TraceID())
	assert.Equal(t, parentSpanCtx.SpanID(), childSpanCtx.SpanID())
}
