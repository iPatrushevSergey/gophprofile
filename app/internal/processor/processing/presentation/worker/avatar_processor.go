// Package worker runs avatar processing background tasks.
package worker

import (
	"context"
	"errors"
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/port"
)

// AvatarProcessorWorker consumes broker events and runs processing use cases.
type AvatarProcessorWorker struct {
	useCases presport.ProcessingUseCases
	log      pkgport.Logger
	tracer   pkgport.Tracer
}

// NewAvatarProcessorWorker creates avatar processor worker.
func NewAvatarProcessorWorker(
	useCases presport.ProcessingUseCases,
	log pkgport.Logger,
	tracer pkgport.Tracer,
) *AvatarProcessorWorker {
	return &AvatarProcessorWorker{useCases: useCases, log: log, tracer: tracer}
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

		processCtx := msg.Ctx
		if processCtx == nil {
			processCtx = context.WithoutCancel(ctx)
		}

		var processSpan pkgport.Span
		switch {
		case msg.Uploaded != nil:
			processCtx, processSpan = w.tracer.Start(processCtx, pkgport.SpanConfig{
				Key:  "processing.presentation.avatar_processor_worker.process_uploaded",
				Name: "process avatar.uploaded",
				Kind: pkgport.SpanKindConsumer,
				Attributes: []pkgport.Attribute{
					{Key: "messaging.system", Value: "rabbitmq"},
					{Key: "messaging.rabbitmq.destination.routing_key", Value: string(vo.EventAvatarUploaded)},
					{Key: "messaging.operation.type", Value: "process"},
				},
			})
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
			processCtx, processSpan = w.tracer.Start(processCtx, pkgport.SpanConfig{
				Key:  "processing.presentation.avatar_processor_worker.process_deleted",
				Name: "process avatar.deleted",
				Kind: pkgport.SpanKindConsumer,
				Attributes: []pkgport.Attribute{
					{Key: "messaging.system", Value: "rabbitmq"},
					{Key: "messaging.rabbitmq.destination.routing_key", Value: string(vo.EventAvatarDeleted)},
					{Key: "messaging.operation.type", Value: "process"},
				},
			})
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

		if processSpan != nil {
			if err != nil && !errors.Is(err, application.ErrAlreadyProcessed) {
				processSpan.Fail(err)
			}
			processSpan.End()
		}

		if _, err := w.useCases.ConfirmAvatarEventUseCase().Execute(ctx, dto.ConfirmAvatarEventInput{
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
