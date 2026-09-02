package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

	"lotto-journal/api/internal/client"
	"lotto-journal/api/internal/config"
	"lotto-journal/api/internal/database"
	"lotto-journal/api/internal/handler"
	"lotto-journal/api/internal/repository"
	"lotto-journal/api/internal/service"
	"lotto-journal/api/middlewares"

	_ "lotto-journal/api/docs"

	"github.com/gofiber/contrib/v3/swaggo"
)

// @title Lotto Journal API
// @version 0.2
// @description Backend API for the Lotto Journal LINE Bot and verification engine.
// @host localhost:3000
// @BasePath /
func main() {
	// Load config
	cfg := config.LoadConfig()

	// Enforce CRON_SECRET is configured in production and staging to fail fast
	if (cfg.APP_ENV == "production" || cfg.APP_ENV == "staging") && cfg.CronSecret == "" {
		log.Fatal("[main] CRON_SECRET environment variable is required in production and staging environments.")
	} else if cfg.CronSecret == "" {
		log.Println("[main] WARNING: CRON_SECRET is empty. Job routes (/jobs/*) will return 401 Unauthorized for all requests. Please configure CRON_SECRET in your environment or .env file.")
	}

	db, err := database.ConnectDatabase(
		cfg.DB_DSN,
		database.PoolConfig{
			MaxIdleConns:    cfg.DBMaxIdleConns,
			MaxOpenConns:    cfg.DBMaxOpenConns,
			ConnMaxLifetime: cfg.DBConnMaxLifetime,
			ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// LINE bot client
	bot, err := messaging_api.NewMessagingApiAPI(cfg.LineChannelAccessToken)
	if err != nil {
		log.Fatalf("Failed to create LINE bot client: %v", err)
	}

	blobBot, err := messaging_api.NewMessagingApiBlobAPI(cfg.LineChannelAccessToken)
	if err != nil {
		log.Fatalf("Failed to create LINE bot blob client: %v", err)
	}

	// GLO Lottery client
	lotteryClient := client.NewLotteryClient("")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	drawRepo := repository.NewDrawRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	webhookRepo := repository.NewWebhookEventRepository(db)
	drawResultRepo := repository.NewDrawResultRepository(db)
	winningRepo := repository.NewUserWinningRepository(db)
	fileRepo := repository.NewFileRepository(db)
	ocrSessionRepo := repository.NewOcrSessionRepository(db)

	// Services
	userSvc := service.NewUserService(userRepo)
	drawSvc := service.NewDrawService(drawRepo, lotteryClient)
	ticketSvc := service.NewTicketService(ticketRepo, drawRepo, drawSvc)
	notificationSvc := service.NewNotificationService(db, bot, ticketRepo, winningRepo, drawResultRepo)
	resultSvc := service.NewResultService(db, lotteryClient, drawRepo, drawResultRepo, ticketRepo, winningRepo, notificationSvc)

	storageSvc, err := service.NewStorageService(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2BucketName)
	if err != nil {
		log.Fatalf("failed to initialize R2 storage service: %v", err)
	}
	ocrSvc := service.NewOcrService(cfg.OpenAIAPIKey)
	ocrSessionSvc := service.NewOcrSessionService(ocrSessionRepo, ticketRepo, drawSvc)

	// Start startup draw schedule sync (only in development/non-production for convenience)
	if cfg.APP_ENV != "production" {
		go func() {
			log.Println("[main] Executing startup draw schedule sync...")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := drawSvc.SyncDrawSchedule(ctx); err != nil {
				log.Printf("[main] Startup draw schedule sync failed: %v", err)
			}
		}()
	}

	// Handlers
	healthHandler := handler.NewHealthHandler(db)
	jobHandler := handler.NewJobHandler(drawSvc, resultSvc, cfg.CronSecret)
	lineHandler := handler.NewLineHandler(
		cfg.LineChannelSecret,
		bot,
		blobBot,
		userSvc,
		ticketSvc,
		notificationSvc,
		webhookRepo,
		storageSvc,
		ocrSvc,
		ocrSessionSvc,
		fileRepo,
	)

	// Fiber app
	app := fiber.New()

	// Global middleware — order matters:
	//   1. recoverer : catch panics before anything else runs, return 500 + stack trace
	//   2. requestid : assign a trace ID so subsequent middleware can read it
	//   3. Logging   : log after the handler finishes (reads requestid + status code)
	app.Use(recoverer.New(recoverer.Config{EnableStackTrace: true}))
	app.Use(requestid.New())
	app.Use(middlewares.Logging)

	// Routes
	app.Get("/health", healthHandler.Handle)

	// Cron / Scheduled Job Trigger Routes
	app.Post("/jobs/sync-schedule", jobHandler.SyncSchedule)
	app.Post("/jobs/verify-results", jobHandler.VerifyResults)

	if cfg.APP_ENV != "production" {
		log.Println("[main] Swagger UI available at /swagger/index.html")
		app.Get("/swagger/*", swaggo.HandlerDefault)
	}

	// timeout.New wraps the handler: if it does not return within 25 s, Fiber responds 408.
	// Applied only to /webhook — other routes (e.g. /health) have no artificial deadline.
	app.Post("/webhook", timeout.New(lineHandler.Handle, timeout.Config{
		Timeout: 25 * time.Second,
	}))

	// Start server
	log.Fatal(app.Listen(cfg.PORT))
}
