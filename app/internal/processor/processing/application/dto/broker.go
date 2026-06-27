package dto

import "context"

// MessageDelivery confirms broker message delivery.
type MessageDelivery interface {
	Ack(ctx context.Context) error
	Nack(ctx context.Context, requeue bool) error
}

// BrokerMessage is a decoded broker event paired with its delivery handle.
type BrokerMessage struct {
	Uploaded *AvatarUploadedEvent
	Deleted  *AvatarDeletedEvent
	Delivery MessageDelivery
}

// ConfirmAvatarEventInput is application input for broker delivery confirmation.
type ConfirmAvatarEventInput struct {
	Delivery MessageDelivery
	Success  bool
	Requeue  bool
}
