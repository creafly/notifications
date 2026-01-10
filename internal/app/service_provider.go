package app

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/creafly/logger"
	"github.com/creafly/notifications/internal/config"
	"github.com/creafly/notifications/internal/domain/repository"
	"github.com/creafly/notifications/internal/domain/service"
	"github.com/creafly/notifications/internal/handler"
	"github.com/creafly/notifications/internal/infra/client"
	"github.com/creafly/notifications/internal/infra/database"
	"github.com/creafly/notifications/internal/infra/kafka"
	"github.com/creafly/notifications/internal/infra/websocket"
	"github.com/creafly/outbox"
	"github.com/creafly/outbox/strategies"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/xlab/closer"
)

type serviceProvider struct {
	cfg *config.Config

	db       *sqlx.DB
	migrator *database.Migrator

	kafkaProducer sarama.SyncProducer

	hub *websocket.Hub

	outboxEventHandler outbox.EventHandler
	outboxWorker       *outbox.Worker

	notificationRepo     repository.NotificationRepository
	invitationRepo       repository.InvitationRepository
	outboxRepo           outbox.Repository
	pushNotificationRepo repository.PushNotificationRepository

	invitationsConsumer *kafka.InvitationsConsumer

	notificationSvc     service.NotificationService
	invitationSvc       service.InvitationService
	pushNotificationSvc service.PushNotificationService

	notificationHnd     *handler.NotificationHandler
	invitationHnd       *handler.InvitationHandler
	healthHnd           *handler.HealthHandler
	pushNotificationHnd *handler.PushNotificationHandler

	identityClient *client.IdentityClient

	httpEngine *gin.Engine
}

func NewServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (sp *serviceProvider) GetConfig() *config.Config {
	if sp.cfg == nil {
		sp.cfg = config.Load()
	}
	return sp.cfg
}

func (sp *serviceProvider) GetDB() *sqlx.DB {
	if sp.db == nil {
		db, err := sqlx.Connect("postgres", sp.GetConfig().Database.URL)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to database")
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		sp.db = db

		closer.Bind(func() {
			_ = sp.db.Close()
		})
	}
	return sp.db
}

func (sp *serviceProvider) GetMigrator() *database.Migrator {
	if sp.migrator == nil {
		sp.migrator = database.NewMigrator(sp.GetDB(), "migrations")
	}
	return sp.migrator
}

func (sp *serviceProvider) GetHub() *websocket.Hub {
	if sp.hub == nil {
		sp.hub = websocket.NewHub()
	}
	return sp.hub
}

func (sp *serviceProvider) GetKafkaProducer() sarama.SyncProducer {
	if sp.kafkaProducer == nil && sp.GetConfig().Kafka.Enabled {
		kafkaConfig := sarama.NewConfig()
		kafkaConfig.Producer.Return.Successes = true
		kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
		kafkaConfig.Producer.Retry.Max = 3

		producer, err := sarama.NewSyncProducer(sp.GetConfig().Kafka.Brokers, kafkaConfig)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create Kafka producer, using log handler")
			return nil
		}

		sp.kafkaProducer = producer

		closer.Bind(func() {
			if err := sp.kafkaProducer.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing Kafka producer")
			}
		})
	}
	return sp.kafkaProducer
}

func (sp *serviceProvider) GetOutboxEventHandler() outbox.EventHandler {
	if sp.outboxEventHandler == nil {
		if sp.GetConfig().Kafka.Enabled && sp.GetKafkaProducer() != nil {
			kafkaStrategy := strategies.NewKafkaStrategyBuilder(sp.GetKafkaProducer()).
				WithDefaultTopic("events").
				WithTopicMappings(map[string]string{
					"invitations.created":     "invitations",
					"invitations.accepted":    "invitations",
					"invitations.rejected":    "invitations",
					"notifications.created":   "notifications",
					"subscriptions.created":   "subscriptions",
					"subscriptions.updated":   "subscriptions",
					"subscriptions.cancelled": "subscriptions",
					"usage.logged":            "usage",
				}).
				Build()
			resolver := outbox.NewStrategyResolver(kafkaStrategy)
			sp.outboxEventHandler = outbox.NewCompositeHandler(resolver)
		} else {
			sp.outboxEventHandler = outbox.NewFuncHandler(func(ctx context.Context, event *outbox.Event) error {
				logger.Info().
					Str("event_id", event.ID.String()).
					Str("event_type", event.EventType).
					Str("payload", event.Payload).
					Msg("outbox event (log handler)")
				return nil
			})
		}
	}
	return sp.outboxEventHandler
}

func (sp *serviceProvider) GetOutboxWorker() *outbox.Worker {
	if sp.outboxWorker == nil && sp.GetConfig().Outbox.Enabled {
		cfg := sp.GetConfig().Outbox
		workerConfig := outbox.WorkerConfig{
			PollInterval:          cfg.PollInterval,
			BatchSize:             cfg.BatchSize,
			CleanupInterval:       cfg.CleanupInterval,
			RetentionPeriod:       cfg.RetentionPeriod,
			FailedRetentionPeriod: cfg.FailedRetentionPeriod,
		}

		sp.outboxWorker = outbox.NewWorker(sp.GetOutboxRepo(), sp.GetOutboxEventHandler(), workerConfig)
		closer.Bind(sp.outboxWorker.Stop)
	}
	return sp.outboxWorker
}

func (sp *serviceProvider) GetInvitationsConsumer() *kafka.InvitationsConsumer {
	if sp.invitationsConsumer == nil && sp.GetConfig().Kafka.Enabled {
		consumer, err := kafka.NewInvitationsConsumer(
			sp.GetConfig().Kafka.Brokers,
			sp.GetConfig().Kafka.GroupID,
			sp.GetInvitationSvc(),
		)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create Kafka consumer")
			return nil
		}

		sp.invitationsConsumer = consumer
		closer.Bind(func() {
			_ = sp.invitationsConsumer.Stop()
		})
	}
	return sp.invitationsConsumer
}

func (sp *serviceProvider) GetNotificationRepo() repository.NotificationRepository {
	if sp.notificationRepo == nil {
		sp.notificationRepo = repository.NewNotificationRepository(sp.GetDB())
	}
	return sp.notificationRepo
}

func (sp *serviceProvider) GetInvitationRepo() repository.InvitationRepository {
	if sp.invitationRepo == nil {
		sp.invitationRepo = repository.NewInvitationRepository(sp.GetDB())
	}
	return sp.invitationRepo
}

func (sp *serviceProvider) GetOutboxRepo() outbox.Repository {
	if sp.outboxRepo == nil {
		sp.outboxRepo = outbox.NewPostgresRepository(sp.GetDB())
	}
	return sp.outboxRepo
}

func (sp *serviceProvider) GetPushNotificationRepo() repository.PushNotificationRepository {
	if sp.pushNotificationRepo == nil {
		sp.pushNotificationRepo = repository.NewPushNotificationRepository(sp.GetDB())
	}
	return sp.pushNotificationRepo
}

func (sp *serviceProvider) GetNotificationSvc() service.NotificationService {
	if sp.notificationSvc == nil {
		sp.notificationSvc = service.NewNotificationService(
			sp.GetNotificationRepo(),
			sp.GetHub(),
		)
	}
	return sp.notificationSvc
}

func (sp *serviceProvider) GetInvitationSvc() service.InvitationService {
	if sp.invitationSvc == nil {
		sp.invitationSvc = service.NewInvitationService(
			sp.GetInvitationRepo(),
			sp.GetOutboxRepo(),
			sp.GetNotificationSvc(),
		)
	}
	return sp.invitationSvc
}

func (sp *serviceProvider) GetPushNotificationSvc() service.PushNotificationService {
	if sp.pushNotificationSvc == nil {
		sp.pushNotificationSvc = service.NewPushNotificationService(
			sp.GetPushNotificationRepo(),
			sp.GetNotificationSvc(),
			sp.GetHub(),
		)
	}
	return sp.pushNotificationSvc
}

func (sp *serviceProvider) GetNotificationHnd() *handler.NotificationHandler {
	if sp.notificationHnd == nil {
		sp.notificationHnd = handler.NewNotificationHandler(sp.GetNotificationSvc())
	}
	return sp.notificationHnd
}

func (sp *serviceProvider) GetInvitationHnd() *handler.InvitationHandler {
	if sp.invitationHnd == nil {
		sp.invitationHnd = handler.NewInvitationHandler(sp.GetInvitationSvc())
	}
	return sp.invitationHnd
}

func (sp *serviceProvider) GetHealthHnd() *handler.HealthHandler {
	if sp.healthHnd == nil {
		sp.healthHnd = handler.NewHealthHandler()
	}
	return sp.healthHnd
}

func (sp *serviceProvider) GetPushNotificationHnd() *handler.PushNotificationHandler {
	if sp.pushNotificationHnd == nil {
		sp.pushNotificationHnd = handler.NewPushNotificationHandler(sp.GetPushNotificationSvc(), sp.GetIdentityClient())
	}
	return sp.pushNotificationHnd
}

func (sp *serviceProvider) GetIdentityClient() *client.IdentityClient {
	if sp.identityClient == nil {
		sp.identityClient = client.NewIdentityClient(sp.GetConfig().Identity.ServiceURL)
	}
	return sp.identityClient
}

func (sp *serviceProvider) GetHttpEngine() *gin.Engine {
	if sp.httpEngine == nil {
		sp.httpEngine = gin.New()
	}
	return sp.httpEngine
}
