package bootstrap

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	avatarappusecase "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/usecase"
	avatarpresport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

// GlobalUseCases provides all module use cases needed by composition root.
type GlobalUseCases interface {
	avatarpresport.AvatarUseCases
}

// globalUseCases implements GlobalUseCases.
type globalUseCases struct {
	upload                     appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput]
	get                        appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput]
	getMetadata                appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]
	list                       appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput]
	delete                     appport.UseCase[appdto.DeleteAvatarInput, struct{}]
	health                     appport.UseCase[struct{}, appdto.HealthCheckOutput]
	expireUploadingAvatars     appport.UseCase[struct{}, struct{}]
	publishPendingOutboxEvents appport.UseCase[struct{}, struct{}]
	collectPeriodicMetrics     appport.UseCase[struct{}, struct{}]
}

var _ avatarpresport.AvatarUseCases = (*globalUseCases)(nil)

// NewGlobalUseCases builds global use cases using functional options.
func NewGlobalUseCases(opts ...apputil.Option[globalUseCasesParams]) GlobalUseCases {
	p := globalUseCasesParams{}
	apputil.Apply(&p, opts...)
	p.validate()

	avatarUseCases := avatarappusecase.NewAvatarUseCases(avatarappusecase.AvatarUseCasesParams{
		AvatarRepo:              p.avatarRepo,
		AvatarStorage:           p.avatarStorage,
		EventPublisher:          p.eventPublisher,
		OutboxRepo:              p.outboxRepo,
		IDGenerator:             p.idGenerator,
		Transactor:              p.transactor,
		Clock:                   p.clock,
		Logger:                  p.logger,
		Tracer:                  p.tracer,
		Metrics:                 p.metrics,
		PoolStats:               p.poolStats,
		OutboxBatchSize:         p.outboxBatchSize,
		OutboxPublishingTimeout: p.outboxPublishingTimeout,
		UploadReservationTTL:    p.uploadReservationTTL,
	})

	return &globalUseCases{
		upload:                     avatarUseCases.Upload,
		get:                        avatarUseCases.Get,
		getMetadata:                avatarUseCases.GetMetadata,
		list:                       avatarUseCases.ListByUser,
		delete:                     avatarUseCases.Delete,
		health:                     avatarUseCases.Health,
		expireUploadingAvatars:     avatarUseCases.ExpireUploadingAvatars,
		publishPendingOutboxEvents: avatarUseCases.PublishPendingOutboxEvents,
		collectPeriodicMetrics:     avatarUseCases.CollectPeriodicMetrics,
	}
}

// UploadUseCase returns the upload avatar use case.
func (f *globalUseCases) UploadUseCase() appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput] {
	return f.upload
}

// GetUseCase returns the get avatar use case.
func (f *globalUseCases) GetUseCase() appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput] {
	return f.get
}

// GetMetadataUseCase returns the get avatar metadata use case.
func (f *globalUseCases) GetMetadataUseCase() appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput] {
	return f.getMetadata
}

// ListByUserUseCase returns the list user avatars use case.
func (f *globalUseCases) ListByUserUseCase() appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput] {
	return f.list
}

// DeleteUseCase returns the delete avatar use case.
func (f *globalUseCases) DeleteUseCase() appport.UseCase[appdto.DeleteAvatarInput, struct{}] {
	return f.delete
}

// HealthUseCase returns the health check use case.
func (f *globalUseCases) HealthUseCase() appport.UseCase[struct{}, appdto.HealthCheckOutput] {
	return f.health
}

// ExpireUploadingAvatarsUseCase returns the expire uploading avatars use case.
func (f *globalUseCases) ExpireUploadingAvatarsUseCase() appport.UseCase[struct{}, struct{}] {
	return f.expireUploadingAvatars
}

// PublishPendingOutboxEventsUseCase returns the publish pending outbox events use case.
func (f *globalUseCases) PublishPendingOutboxEventsUseCase() appport.UseCase[struct{}, struct{}] {
	return f.publishPendingOutboxEvents
}

// CollectPeriodicMetricsUseCase returns the collect periodic metrics use case.
func (f *globalUseCases) CollectPeriodicMetricsUseCase() appport.UseCase[struct{}, struct{}] {
	return f.collectPeriodicMetrics
}

// globalUseCasesParams holds dependencies required to build global use cases.
type globalUseCasesParams struct {
	avatarRepo              appport.AvatarRepo
	outboxRepo              appport.OutboxRepo
	avatarStorage           appport.AvatarStorage
	eventPublisher          appport.EventPublisher
	idGenerator             appport.IDGenerator
	transactor              appport.Transactor
	clock                   appport.Clock
	logger                  pkgport.Logger
	tracer                  pkgport.Tracer
	metrics                 pkgport.Metrics
	poolStats               pkgport.PoolStats
	outboxBatchSize         int
	outboxPublishingTimeout time.Duration
	uploadReservationTTL    time.Duration
}

// validate validates the global use cases parameters.
func (p globalUseCasesParams) validate() {
	if p.avatarRepo == nil {
		panic("NewGlobalUseCases: WithAvatarRepo is required")
	}
	if p.outboxRepo == nil {
		panic("NewGlobalUseCases: WithOutboxRepo is required")
	}
	if p.avatarStorage == nil {
		panic("NewGlobalUseCases: WithAvatarStorage is required")
	}
	if p.eventPublisher == nil {
		panic("NewGlobalUseCases: WithEventPublisher is required")
	}
	if p.idGenerator == nil {
		panic("NewGlobalUseCases: WithIDGenerator is required")
	}
	if p.transactor == nil {
		panic("NewGlobalUseCases: WithTransactor is required")
	}
	if p.clock == nil {
		panic("NewGlobalUseCases: WithClock is required")
	}
	if p.logger == nil {
		panic("NewGlobalUseCases: WithLogger is required")
	}
}

// WithAvatarRepo sets the avatar repository.
func WithAvatarRepo(r appport.AvatarRepo) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.avatarRepo = r }
}

// WithOutboxRepo sets the outbox repository.
func WithOutboxRepo(r appport.OutboxRepo) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.outboxRepo = r }
}

// WithAvatarStorage sets the avatar object storage adapter.
func WithAvatarStorage(s appport.AvatarStorage) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.avatarStorage = s }
}

// WithEventPublisher sets the avatar event publisher.
func WithEventPublisher(publisher appport.EventPublisher) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.eventPublisher = publisher }
}

// WithIDGenerator sets the avatar id generator.
func WithIDGenerator(g appport.IDGenerator) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.idGenerator = g }
}

// WithTransactor sets the transactor.
func WithTransactor(t appport.Transactor) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.transactor = t }
}

// WithClock sets the clock.
func WithClock(c appport.Clock) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.clock = c }
}

// WithLogger sets the logger.
func WithLogger(l pkgport.Logger) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.logger = l }
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

// WithOutboxBatchSize sets the outbox publish batch size.
func WithOutboxBatchSize(size int) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.outboxBatchSize = size }
}

// WithOutboxPublishingTimeout sets how long a publishing row may live before recovery.
func WithOutboxPublishingTimeout(timeout time.Duration) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.outboxPublishingTimeout = timeout }
}

// WithUploadReservationTTL sets the upload reservation TTL.
func WithUploadReservationTTL(ttl time.Duration) apputil.Option[globalUseCasesParams] {
	return func(p *globalUseCasesParams) { p.uploadReservationTTL = ttl }
}
