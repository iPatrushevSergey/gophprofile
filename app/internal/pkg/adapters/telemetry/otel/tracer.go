package otel

import (
	"context"
	"strings"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelprop "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"

// Version returns the instrumentation library version.
func Version() string {
	if v := strings.TrimSpace(apputil.Version); v != "" {
		return v
	}

	return "0.0.0"
}

// Tracer is an OpenTelemetry implementation of the pkgport.Tracer port.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates an OpenTelemetry implementation of the pkgport.Tracer port.
func NewTracer() *Tracer {
	return &Tracer{
		tracer: otel.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(Version()),
		),
	}
}

var _ port.Tracer = (*Tracer)(nil)

// Start starts a span with the given configuration.
func (t *Tracer) Start(ctx context.Context, config port.SpanConfig) (context.Context, port.Span) {
	if t == nil {
		return ctx, noopSpan{}
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(toOTelSpanKind(config.Kind)),
	}
	if attrs := toOTelAttributes(config.Attributes); len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}

	ctx, span := t.tracer.Start(ctx, config.Name, opts...)
	if config.Key != "" {
		span.SetAttributes(attribute.String("gophprofile.telemetry.key", config.Key))
	}

	return ctx, tracedSpan{span: span}
}

// ContextToMap serializes active trace context into a string map.
func (*Tracer) ContextToMap(ctx context.Context) map[string]string {
	carrier := otelprop.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}

	out := make(map[string]string, len(carrier))
	for k, v := range carrier {
		out[k] = v
	}

	return out
}

// MapToContext reconstructs trace context from a previously captured carrier map.
func (*Tracer) MapToContext(ctx context.Context, carrier map[string]string) context.Context {
	if len(carrier) == 0 {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, otelprop.MapCarrier(carrier))
}

// tracedSpan is an OpenTelemetry implementation of the pkgport.Span port.
type tracedSpan struct {
	span trace.Span
}

// End ends the span.
func (s tracedSpan) End() {
	s.span.End()
}

// Fail fails the span.
func (s tracedSpan) Fail(err error) {
	if err == nil {
		return
	}

	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// AddAttributes adds attributes to the span.
func (s tracedSpan) AddAttributes(attrs ...port.Attribute) {
	s.span.SetAttributes(toOTelAttributes(attrs)...)
}

// noopSpan is a no-op implementation of the pkgport.Span port.
type noopSpan struct{}

// End ends the span.
func (noopSpan) End() {}

// Fail fails the span.
func (noopSpan) Fail(error) {}

// AddAttributes adds attributes to the span.
func (noopSpan) AddAttributes(...port.Attribute) {}

// toOTelSpanKind converts port.SpanKind to OpenTelemetry trace.SpanKind.
func toOTelSpanKind(kind port.SpanKind) trace.SpanKind {
	switch kind {
	case port.SpanKindServer:
		return trace.SpanKindServer
	case port.SpanKindClient:
		return trace.SpanKindClient
	case port.SpanKindProducer:
		return trace.SpanKindProducer
	case port.SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

// toOTelAttributes converts port.Attributes to OpenTelemetry attribute.KeyValue.
func toOTelAttributes(attrs []port.Attribute) []attribute.KeyValue {
	if len(attrs) == 0 {
		return nil
	}

	out := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		switch v := attr.Value.(type) {
		case string:
			out = append(out, attribute.String(attr.Key, v))
		case bool:
			out = append(out, attribute.Bool(attr.Key, v))
		case int:
			out = append(out, attribute.Int(attr.Key, v))
		case int64:
			out = append(out, attribute.Int64(attr.Key, v))
		case float64:
			out = append(out, attribute.Float64(attr.Key, v))
		default:
			if v != nil {
				out = append(out, attribute.String(attr.Key, "unsupported"))
			}
		}
	}

	return out
}
