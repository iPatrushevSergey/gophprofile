package bootstrap

import (
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	processingappusecase "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/usecase"
	processingpresport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/port"
)

// GlobalUseCases provides all module use cases needed by composition root.
type GlobalUseCases interface {
	processingpresport.ProcessingUseCases
}

// globalUseCases implements GlobalUseCases.
type globalUseCases struct {
	subscribeAvatarEvents  appport.UseCase[struct{}, <-chan appdto.BrokerMessage]
	confirmAvatarEvent     appport.UseCase[appdto.ConfirmAvatarEventInput, struct{}]
	processUploaded        appport.UseCase[appdto.ProcessUploadedAvatarInput, struct{}]
	purgeDeleted           appport.UseCase[appdto.PurgeDeletedAvatarInput, struct{}]
	collectPeriodicMetrics appport.UseCase[struct{}, struct{}]
}

var _ processingpresport.ProcessingUseCases = (*globalUseCases)(nil)

// NewGlobalUseCases builds global use cases using functional options.
func NewGlobalUseCases(opts ...apputil.Option[globalUseCasesParams]) GlobalUseCases {
	p := globalUseCasesParams{}
	apputil.Apply(&p, opts...)
	p.validate()

	processingUseCases := processingappusecase.NewProcessingUseCases(processingappusecase.ProcessingUseCasesParams{
		AvatarRepo:     p.avatarRepo,
		AvatarStorage:  p.avatarStorage,
		ImageProcessor: p.imageProcessor,
		EventConsumer:  p.eventConsumer,
		Clock:          p.clock,
		Tracer:         p.tracer,
		Metrics:        p.metrics,
		PoolStats:      p.poolStats,
	})

	return &globalUseCases{
		subscribeAvatarEvents:  processingUseCases.SubscribeAvatarEvents,
		confirmAvatarEvent:     processingUseCases.ConfirmAvatarEvent,
		processUploaded:        processingUseCases.ProcessUploaded,
		purgeDeleted:           processingUseCases.PurgeDeleted,
		collectPeriodicMetrics: processingUseCases.CollectPeriodicMetrics,
	}
}

// SubscribeAvatarEventsUseCase returns the subscribe avatar events use case.
func (f *globalUseCases) SubscribeAvatarEventsUseCase() appport.UseCase[struct{}, <-chan appdto.BrokerMessage] {
	return f.subscribeAvatarEvents
}

// ConfirmAvatarEventUseCase returns the confirm avatar event use case.
func (f *globalUseCases) ConfirmAvatarEventUseCase() appport.UseCase[appdto.ConfirmAvatarEventInput, struct{}] {
	return f.confirmAvatarEvent
}

// ProcessUploadedUseCase returns the process uploaded avatar use case.
func (f *globalUseCases) ProcessUploadedUseCase() appport.UseCase[appdto.ProcessUploadedAvatarInput, struct{}] {
	return f.processUploaded
}

// PurgeDeletedUseCase returns the purge deleted avatar use case.
func (f *globalUseCases) PurgeDeletedUseCase() appport.UseCase[appdto.PurgeDeletedAvatarInput, struct{}] {
	return f.purgeDeleted
}

// CollectPeriodicMetricsUseCase returns the collect periodic metrics use case.
func (f *globalUseCases) CollectPeriodicMetricsUseCase() appport.UseCase[struct{}, struct{}] {
	return f.collectPeriodicMetrics
}

// globalUseCasesParams holds dependencies required to build global use cases.
type globalUseCasesParams struct {
	avatarRepo     appport.AvatarRepo
	avatarStorage  appport.AvatarStorage
	imageProcessor appport.ImageProcessor
	eventConsumer  appport.EventConsumer
	clock          appport.Clock
	tracer         pkgport.Tracer
	metrics        pkgport.Metrics
	poolStats      pkgport.PoolStats
}

// validate validates the global use cases parameters.
func (p globalUseCasesParams) validate() {
	if p.avatarRepo == nil {
		panic("NewGlobalUseCases: WithAvatarRepo is required")
	}
	if p.avatarStorage == nil {
		panic("NewGlobalUseCases: WithAvatarStorage is required")
	}
	if p.imageProcessor == nil {
		panic("NewGlobalUseCases: WithImageProcessor is required")
	}
	if p.eventConsumer == nil {
		panic("NewGlobalUseCases: WithEventConsumer is required")
	}
	if p.clock == nil {
		panic("NewGlobalUseCases: WithClock is required")
	}
	if p.tracer == nil {
		panic("NewGlobalUseCases: WithTracer is required")
	}
}

// WithAvatarRepo sets the avatar repository.
func WithAvatarRepo(r appport.AvatarRepo) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.avatarRepo = r }
}

// WithAvatarStorage sets the avatar object storage adapter.
func WithAvatarStorage(s appport.AvatarStorage) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.avatarStorage = s }
}

// WithImageProcessor sets the image processor adapter.
func WithImageProcessor(p appport.ImageProcessor) apputil.Option[globalUseCasesParams] {
	return func(params *globalUseCasesParams) { params.imageProcessor = p }
}

// WithEventConsumer sets the broker event consumer adapter.
func WithEventConsumer(c appport.EventConsumer) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.eventConsumer = c }
}

// WithClock sets the clock.
func WithClock(c appport.Clock) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.clock = c }
}

// WithTracer sets the tracer.
func WithTracer(t pkgport.Tracer) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.tracer = t }
}

// WithMetrics sets the metrics recorder.
func WithMetrics(m pkgport.Metrics) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.metrics = m }
}

// WithPoolStats sets the database pool statistics provider.
func WithPoolStats(s pkgport.PoolStats) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.poolStats = s }
}
