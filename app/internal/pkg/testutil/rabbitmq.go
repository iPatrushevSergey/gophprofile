//go:build integration || e2e || component || contract

package testutil

import (
	"context"
	"testing"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	tcrabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	avatarrabbitmq "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq"
)

const (
	testRabbitMQExchange = "gophprofile-test"
	testRabbitMQQueue    = "gophprofile-test-queue"
)

// SetupRabbitMQ starts RabbitMQ in a container and returns a configured publisher.
func SetupRabbitMQ(tb testing.TB) *avatarrabbitmq.Publisher {
	tb.Helper()
	ctx := context.Background()

	container, err := tcrabbitmq.Run(ctx, "rabbitmq:3.13-management-alpine")
	require.NoError(tb, err, "start rabbitmq container")
	tb.Cleanup(func() {
		require.NoError(tb, container.Terminate(ctx))
	})

	amqpURL, err := container.AmqpURL(ctx)
	require.NoError(tb, err)

	conn, err := amqp091.Dial(amqpURL)
	require.NoError(tb, err)

	ch, err := conn.Channel()
	require.NoError(tb, err)

	require.NoError(tb, ch.ExchangeDeclare(
		testRabbitMQExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	))

	queue, err := ch.QueueDeclare(
		testRabbitMQQueue,
		false,
		true,
		false,
		false,
		nil,
	)
	require.NoError(tb, err)
	require.NoError(tb, ch.QueueBind(queue.Name, "avatar.uploaded", testRabbitMQExchange, false, nil))
	require.NoError(tb, ch.QueueBind(queue.Name, "avatar.deleted", testRabbitMQExchange, false, nil))

	require.NoError(tb, ch.Close())
	require.NoError(tb, conn.Close())

	publisher, err := avatarrabbitmq.NewPublisher(avatarrabbitmq.Config{
		URL:                     amqpURL,
		Exchange:                testRabbitMQExchange,
		PublishInterval:         time.Second,
		OutboxBatchSize:         10,
		OutboxPublishingTimeout: time.Minute,
	})
	require.NoError(tb, err)

	return publisher
}
