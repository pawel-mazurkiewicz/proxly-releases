package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/config"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/core"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/external"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/http/handlers"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/http/middleware"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/notify"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func New(cfg *config.Config, logger *zap.Logger, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// Initialize repositories
	licensesRepo := store.NewLicensesRepository(db)
	activationsRepo := store.NewActivationsRepository(db)
	trialsRepo := store.NewTrialsRepository(db)
	webhooksRepo := store.NewWebhooksRepository(db)
	auditRepo := store.NewAuditLogsRepository(db)

	// Initialize services
	authService := core.NewAuthService(cfg.JWTAdminSecret, cfg.HMACSecret)
	webhookNotifier := notify.NewWebhookNotifier(webhooksRepo, cfg.WebhookTimeoutMS)
	emailNotifier := notify.NewEmailNotifier(false) // TODO: Make configurable
	notifier := core.NewCompositeNotifier(webhookNotifier, emailNotifier)

	// Initialize Gumroad client
	gumroadClient := external.NewGumroadClient(cfg.GumroadProductID)

	trialService := core.NewTrialService(trialsRepo, activationsRepo, cfg)
	trialService.StartCleanupRoutine()

	licenseService := core.NewLicenseService(
		licensesRepo,
		activationsRepo,
		notifier,
		auditRepo,
		gumroadClient,
		trialService,
		cfg.DefaultMaxActivations,
		cfg.AutoCreateOnActivation,
		cfg.GumroadVerificationEnabled,
		logger,
	)

	// Initialize rate limiters
	ipLimiter := core.NewRateLimiter(cfg.RateLimitPerIPPerMin)
	licenseLimiter := core.NewRateLimiter(cfg.RateLimitPerLicensePerMin)

	// Start cleanup routines
	ipLimiter.StartCleanupRoutine()
	licenseLimiter.StartCleanupRoutine()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	activationsHandler := handlers.NewActivationsHandler(licenseService, logger)
	trialsHandler := handlers.NewTrialsHandler(trialService)
	licensesHandler := handlers.NewLicensesHandler(licenseService, logger)
	adminHandler := handlers.NewAdminHandler(webhooksRepo, logger)

	// Global middleware
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.Security())
	r.Use(middleware.RateLimit(ipLimiter, licenseLimiter))

	// Health check endpoint (no auth required)
	r.Get("/healthz", healthHandler.Health)

	// Metrics endpoint (no auth required)
	r.Handle("/metrics", promhttp.Handler())

	// Public API routes (with HMAC auth)
	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.HMACAuth(authService))
		r.Post("/activate", activationsHandler.Activate)
		r.Post("/trial/start", trialsHandler.Start)
		r.Get("/trial/status", trialsHandler.Status)
	})

	// Admin API routes (with JWT auth)
	r.Route("/v1/admin", func(r chi.Router) {
		r.Use(middleware.JWTAuth(authService))

		// License management
		r.Route("/licenses", func(r chi.Router) {
			r.Get("/", licensesHandler.List)
			r.Post("/", licensesHandler.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", licensesHandler.Get)
				r.Put("/", licensesHandler.Update)
				r.Delete("/", licensesHandler.Delete)
				r.Get("/activations", licensesHandler.ListActivations)
				r.Post("/reset-activations", licensesHandler.ResetActivations)
			})
		})

		// Webhook management
		r.Route("/webhooks", func(r chi.Router) {
			r.Get("/", adminHandler.ListWebhooks)
			r.Post("/", adminHandler.CreateWebhook)
		})

		// Trial management
		r.Route("/trials", func(r chi.Router) {
			r.Get("/", trialsHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Post("/status", trialsHandler.UpdateStatus)
				r.Put("/", trialsHandler.Update)
				r.Delete("/", trialsHandler.Delete)
			})
		})
	})

	return r
}
