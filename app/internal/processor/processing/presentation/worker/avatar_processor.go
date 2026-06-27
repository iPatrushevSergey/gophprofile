// Package worker runs avatar processing background tasks.
package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/port"
)

// AvatarProcessorWorker consumes broker events and runs processing use cases.
type AvatarProcessorWorker struct {
	useCases presport.ProcessingUseCases
	log      appport.Logger
}

// NewAvatarProcessorWorker creates avatar processor worker.
func NewAvatarProcessorWorker(
	useCases presport.ProcessingUseCases,
	log appport.Logger,
) *AvatarProcessorWorker {
	return &AvatarProcessorWorker{useCases: useCases, log: log}
}

// Run starts event consumption loop.
func (w *AvatarProcessorWorker) Run(ctx context.Context) error {
	messages, err := w.useCases.SubscribeAvatarEventsUseCase().Execute(ctx, struct{}{})
	if err != nil {
		return err
	}

	for msg := range messages {
		success := true
		requeue := false

		processCtx := context.WithoutCancel(ctx)

		switch {
		case msg.Uploaded != nil:
			event := msg.Uploaded
			_, err = w.useCases.ProcessUploadedUseCase().Execute(processCtx, dto.ProcessUploadedAvatarInput{
				AvatarID: event.AvatarID,
				UserID:   event.UserID,
				S3Key:    event.S3Key,
			})
			if err != nil {
				switch {
				case errors.Is(err, application.ErrAlreadyProcessed):
					// success stays true → Ack
				case errors.Is(err, application.ErrNotFound):
					w.log.Error("process uploaded avatar: avatar not found",
						"error", err,
						"avatar_id", event.AvatarID,
						"user_id", event.UserID,
					)
					success = false
				case errors.Is(err, application.ErrBadInput):
					w.log.Error("process uploaded avatar: bad input",
						"error", err,
						"avatar_id", event.AvatarID,
						"user_id", event.UserID,
					)
					success = false
				default:
					w.log.Error("process uploaded avatar failed",
						"error", err,
						"avatar_id", event.AvatarID,
						"user_id", event.UserID,
					)
					success = false
					requeue = true
				}
			}
		case msg.Deleted != nil:
			event := msg.Deleted
			_, err = w.useCases.PurgeDeletedUseCase().Execute(processCtx, dto.PurgeDeletedAvatarInput{
				AvatarID: event.AvatarID,
				S3Keys:   event.S3Keys,
			})
			if err != nil {
				success = false
				switch {
				case errors.Is(err, application.ErrBadInput):
					w.log.Error("purge deleted avatar: bad input",
						"error", err,
						"avatar_id", event.AvatarID,
					)
				default:
					w.log.Error("purge deleted avatar failed",
						"error", err,
						"avatar_id", event.AvatarID,
					)
					requeue = true
				}
			}
		default:
			w.log.Error("received broker event without payload")
			success = false
		}

		if _, err := w.useCases.ConfirmAvatarEventUseCase().Execute(processCtx, dto.ConfirmAvatarEventInput{
			Delivery: msg.Delivery,
			Success:  success,
			Requeue:  requeue,
		}); err != nil {
			w.log.Error("confirm avatar event failed", "error", err)
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return fmt.Errorf("broker: event channel closed")
}
