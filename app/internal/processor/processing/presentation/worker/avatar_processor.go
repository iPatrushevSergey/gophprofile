// Package worker runs avatar processing background tasks.
package worker

import (
	"context"
	"errors"
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/port"
)

// AvatarProcessorWorker consumes broker events and runs processing use cases.
type AvatarProcessorWorker struct {
	useCases presport.ProcessingUseCases
	log      pkgport.Logger
}

// NewAvatarProcessorWorker creates avatar processor worker.
func NewAvatarProcessorWorker(
	useCases presport.ProcessingUseCases,
	log pkgport.Logger,
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

		// Detach in-flight work from worker shutdown so processing and ack can finish.
		workCtx := context.WithoutCancel(msg.Ctx)

		switch {
		case msg.Uploaded != nil:
			event := msg.Uploaded
			_, err = w.useCases.ProcessUploadedUseCase().Execute(workCtx, dto.ProcessUploadedAvatarInput{
				AvatarID: event.AvatarID,
				UserID:   event.UserID,
				S3Key:    event.S3Key,
			})
			if err != nil {
				switch {
				case errors.Is(err, application.ErrAlreadyProcessed):
					// success stays true → Ack
				case errors.Is(err, application.ErrNotFound):
					w.log.Error(workCtx, "process uploaded avatar: avatar not found",
						"error", err,
						"avatar_id", event.AvatarID,
						"user_id", event.UserID,
					)
					success = false
				case errors.Is(err, application.ErrBadInput):
					w.log.Error(workCtx, "process uploaded avatar: bad input",
						"error", err,
						"avatar_id", event.AvatarID,
						"user_id", event.UserID,
					)
					success = false
				default:
					w.log.Error(workCtx, "process uploaded avatar failed",
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
			_, err = w.useCases.PurgeDeletedUseCase().Execute(workCtx, dto.PurgeDeletedAvatarInput{
				AvatarID: event.AvatarID,
				S3Keys:   event.S3Keys,
			})
			if err != nil {
				success = false
				switch {
				case errors.Is(err, application.ErrBadInput):
					w.log.Error(workCtx, "purge deleted avatar: bad input",
						"error", err,
						"avatar_id", event.AvatarID,
					)
				default:
					w.log.Error(workCtx, "purge deleted avatar failed",
						"error", err,
						"avatar_id", event.AvatarID,
					)
					requeue = true
				}
			}
		default:
			w.log.Error(workCtx, "received broker event without payload")
			success = false
		}

		if _, err := w.useCases.ConfirmAvatarEventUseCase().Execute(workCtx, dto.ConfirmAvatarEventInput{
			Delivery: msg.Delivery,
			Success:  success,
			Requeue:  requeue,
		}); err != nil {
			w.log.Error(workCtx, "confirm avatar event failed", "error", err)
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return fmt.Errorf("broker: event channel closed")
}
