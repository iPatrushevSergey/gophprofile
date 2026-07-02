package port

import "context"

// SpanKind describes the role of a span in a distributed trace.
type SpanKind string

const (
	SpanKindInternal SpanKind = "internal"
	SpanKindServer   SpanKind = "server"
	SpanKindClient   SpanKind = "client"
	SpanKindProducer SpanKind = "producer"
	SpanKindConsumer SpanKind = "consumer"
)

// Tracer starts manual spans and propagates trace context without leaking the telemetry library into callers.
type Tracer interface {
	Start(ctx context.Context, config SpanConfig) (context.Context, Span)
	ContextToMap(ctx context.Context) map[string]string
	MapToContext(ctx context.Context, carrier map[string]string) context.Context
}

// Span represents an in-flight trace span.
type Span interface {
	End()
	Fail(err error)
	AddAttributes(attrs ...Attribute)
}

// Attribute is a telemetry attribute attached to a span.
type Attribute struct {
	Key   string
	Value any
}

// SpanConfig describes a manual span started through the tracer port.
type SpanConfig struct {
	Key        string
	Name       string
	Kind       SpanKind
	Attributes []Attribute
}
