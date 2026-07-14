package otel

import (
	"context"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// TableCarrier adapts RabbitMQ headers to the OpenTelemetry TextMapCarrier interface.
type TableCarrier amqp091.Table

// Get returns the value for a key.
func (c TableCarrier) Get(key string) string {
	if c == nil {
		return ""
	}

	v, ok := c[key]
	if !ok {
		return ""
	}

	s, _ := v.(string)
	return s
}

// Set stores a key-value pair.
func (c TableCarrier) Set(key, value string) {
	if c == nil {
		return
	}

	c[key] = value
}

// Keys lists stored keys.
func (c TableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}

	return keys
}

// InjectAMQP writes trace context from ctx into RabbitMQ headers.
func InjectAMQP(ctx context.Context, headers amqp091.Table) amqp091.Table {
	if headers == nil {
		headers = amqp091.Table{}
	}

	otel.GetTextMapPropagator().Inject(ctx, TableCarrier(headers))

	return headers
}

// ExtractAMQP restores trace context from RabbitMQ headers into ctx.
func ExtractAMQP(ctx context.Context, headers amqp091.Table) context.Context {
	if headers == nil {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, TableCarrier(headers))
}
