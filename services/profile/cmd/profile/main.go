package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"profile/internal/config"
	"profile/internal/db"
	"profile/internal/handlers"
	"profile/internal/repository"
	"profile/internal/server"
	"profile/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg := config.LoadProfile()

	// Connect to database
	pool, err := db.Connect(ctxWithTimeout(10*time.Second), cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(cfg); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	languageRepo := repository.NewLanguageRepository(pool)
	interestRepo := repository.NewInterestRepository(pool)
	userLanguageRepo := repository.NewUserLanguageRepository(pool)
	userInterestRepo := repository.NewUserInterestRepository(pool)
	preferenceRepo := repository.NewUserPreferenceRepository(pool)
	traitRepo := repository.NewUserTraitRepository(pool)
	availabilityRepo := repository.NewUserTimeAvailabilityRepository(pool)

	// Initialize services
	profileService := service.NewProfileService(
		userRepo, languageRepo, interestRepo,
		userLanguageRepo, userInterestRepo, preferenceRepo,
		traitRepo, availabilityRepo,
	)

	// Initialize validator
	validator := validator.New()

	// Initialize handlers
	profileHandler := handlers.NewProfileHandler(profileService, validator)
	healthHandler := handlers.NewHealthHandler(pool)

	// Initialize router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	setupRoutes(router, profileHandler, healthHandler)

	// HTTP server
	srv := server.NewWithRouter(cfg.HTTPPort, router)
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()
	logger.Info("profile service started", zap.String("port", cfg.HTTPPort))

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("profile service stopped")
}

func ctxWithTimeout(d time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), d)
	return ctx
}

func setupRoutes(router *gin.Engine, profileHandler *handlers.ProfileHandler, healthHandler *handlers.HealthHandler) {
	// Health check endpoints
	router.GET("/healthz", healthHandler.Health)
	router.GET("/readyz", healthHandler.Ready)

	// API routes
	api := router.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("", profileHandler.CreateUser)
			users.GET("", profileHandler.ListUsers)
			users.GET("/:id", profileHandler.GetUser)
			users.PUT("/:id", profileHandler.UpdateUser)
			users.DELETE("/:id", profileHandler.DeleteUser)
			users.PUT("/:id/last-seen", profileHandler.UpdateLastSeen)
			users.GET("/:id/completion", profileHandler.GetUserProfileCompletion)
			users.GET("/telegram/:telegram_id", profileHandler.GetUserByTelegramID)
			users.GET("/discord/:discord_id", profileHandler.GetUserByDiscordID)
		}
	}
}
