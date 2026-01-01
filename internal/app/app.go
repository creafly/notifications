package app

import (
	"context"
	"net/http"

	"github.com/creafly/notifications/internal/i18n"
	"github.com/creafly/notifications/internal/infra/websocket"
	"github.com/creafly/notifications/internal/logger"
	"github.com/creafly/notifications/internal/middleware"
	"github.com/creafly/notifications/internal/tracing"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xlab/closer"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type App struct {
	ServiceProvider *serviceProvider
	HttpServer      *http.Server
}

func NewApp() *App {
	return (&App{}).initApp()
}

func (a *App) StartApp(ctx context.Context) {
	cfg := a.ServiceProvider.GetConfig()

	tracingShutdown, err := tracing.Init(tracing.Config{
		ServiceName:    cfg.Tracing.ServiceName,
		ServiceVersion: cfg.Tracing.ServiceVersion,
		Environment:    cfg.Tracing.Environment,
		OTLPEndpoint:   cfg.Tracing.OTLPEndpoint,
		Enabled:        cfg.Tracing.Enabled,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize tracing")
	} else {
		closer.Bind(func() {
			if err := tracingShutdown(context.Background()); err != nil {
				logger.Error().Err(err).Msg("Error shutting down tracer provider")
			}
		})
	}

	i18n.PreloadLocales()

	hub := a.ServiceProvider.GetHub()
	go hub.Run()

	if outboxWorker := a.ServiceProvider.GetOutboxWorker(); outboxWorker != nil {
		outboxWorker.Start(ctx)
	}

	if consumer := a.ServiceProvider.GetInvitationsConsumer(); consumer != nil {
		consumer.Start(ctx)
	}

	go func() {
		if err := a.getHttpServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
}

func (a *App) StartMigrator(migrateUp, migrateDown bool) {
	migrator := a.ServiceProvider.GetMigrator()

	if migrateUp {
		if err := migrator.Up(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations up")
		}
		logger.Info().Msg("Migrations completed successfully")
		return
	}

	if migrateDown {
		if err := migrator.Down(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations down")
		}
		logger.Info().Msg("Migrations rolled back successfully")
		return
	}

	if a.ServiceProvider.GetConfig().Database.AutoMigrate {
		logger.Info().Msg("Running auto-migrations...")
		if err := migrator.Up(); err != nil {
			logger.Warn().Err(err).Msg("Auto-migration failed")
		}
	}
}

func (a *App) initApp() *App {
	logger.Init("notifications")
	closer.Init(closer.Config{
		ExitSignals: closer.DefaultSignalSet,
	})

	a.initServiceProvider()
	a.initHttpServer()

	return a
}

func (a *App) getHttpServer() *http.Server {
	if a.HttpServer == nil {
		cfg := a.ServiceProvider.GetConfig()
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		logger.Info().Str("addr", addr).Msg("Starting Notifications Service")

		a.HttpServer = &http.Server{
			Addr:    addr,
			Handler: a.ServiceProvider.GetHttpEngine(),
		}

		closer.Bind(func() {
			if err := a.HttpServer.Shutdown(context.Background()); err != nil {
				logger.Error().Err(err).Msg("Server forced to shutdown")
			}
		})
	}

	return a.HttpServer
}

func (a *App) initHttpServer() {
	gin.SetMode(a.ServiceProvider.GetConfig().Server.GinMode)

	a.initHttpMiddleware()
	a.initHttpRouting()
}

func (a *App) initServiceProvider() {
	if a.ServiceProvider == nil {
		a.ServiceProvider = NewServiceProvider()
	}
}

func (a *App) initHttpMiddleware() {
	r := a.ServiceProvider.GetHttpEngine()
	cfg := a.ServiceProvider.GetConfig()

	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(otelgin.Middleware("notifications"))
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.PrometheusMiddleware())
	r.Use(middleware.LocaleMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.CORS))
	r.Use(middleware.Compression())
}

func (a *App) initHttpRouting() {
	r := a.ServiceProvider.GetHttpEngine()

	hub := a.ServiceProvider.GetHub()
	identityClient := a.ServiceProvider.GetIdentityClient()

	healthHandler := a.ServiceProvider.GetHealthHnd()
	notificationHandler := a.ServiceProvider.GetNotificationHnd()
	invitationHandler := a.ServiceProvider.GetInvitationHnd()
	pushHandler := a.ServiceProvider.GetPushNotificationHnd()

	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	{
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(identityClient))
		{
			protected.GET("/ws", func(c *gin.Context) {
				websocket.ServeWs(hub, c)
			})

			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetMyNotifications)
				notifications.GET("/unread", notificationHandler.GetUnreadNotifications)
				notifications.GET("/unread/count", notificationHandler.GetUnreadCount)
				notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", notificationHandler.Delete)
			}

			invitations := protected.Group("/invitations")
			{
				invitations.GET("", invitationHandler.GetMyInvitations)
				invitations.POST("", invitationHandler.Create)
				invitations.PUT("/:id/accept", invitationHandler.Accept)
				invitations.PUT("/:id/reject", invitationHandler.Reject)
			}

			protected.GET("/tenants/:tenantId/invitations", invitationHandler.GetByTenant)

			userPush := protected.Group("/push")
			{
				userPush.GET("", pushHandler.GetMyPushNotifications)
				userPush.PUT("/:id/read", pushHandler.MarkAsRead)
			}

			adminPush := protected.Group("/admin/push")
			adminPush.Use(middleware.RequireAnyClaim(identityClient, "push:manage", "push:view"))
			{
				adminPush.GET("", pushHandler.GetAll)
				adminPush.GET("/:id", pushHandler.GetByID)
			}

			adminPushCreate := protected.Group("/admin/push")
			adminPushCreate.Use(middleware.RequireAnyClaim(identityClient, "push:manage", "push:create"))
			{
				adminPushCreate.POST("", pushHandler.Create)
			}

			adminPushUpdate := protected.Group("/admin/push")
			adminPushUpdate.Use(middleware.RequireAnyClaim(identityClient, "push:manage", "push:update"))
			{
				adminPushUpdate.PUT("/:id", pushHandler.Update)
				adminPushUpdate.POST("/:id/send", pushHandler.Send)
				adminPushUpdate.POST("/:id/cancel", pushHandler.Cancel)
			}

			adminPushDelete := protected.Group("/admin/push")
			adminPushDelete.Use(middleware.RequireAnyClaim(identityClient, "push:manage", "push:delete"))
			{
				adminPushDelete.DELETE("/:id", pushHandler.Delete)
			}
		}
	}
}
