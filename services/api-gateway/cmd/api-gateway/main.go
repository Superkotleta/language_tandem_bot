package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"api-gateway/internal/server"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	cfg := config.LoadAPIGateway()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.Info("starting API Gateway", zap.String("port", cfg.HTTPPort))

	// Create proxies
	profileProxy := proxy.New("profile", &cfg.ProfileService)
	botProxy := proxy.New("bot", &cfg.BotService)

	// Create handlers
	profileHandler := handlers.NewProfileProxy(profileProxy, logger)
	botHandler := handlers.NewBotProxy(botProxy, logger)
	healthHandler := handlers.NewHealthHandler(profileHandler, botHandler)

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimitConfig.RequestsPerMinute,
		time.Minute,
	)

	// Setup router
	router := gin.New()

	// Disable all automatic redirects
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Middleware
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))

	if cfg.RateLimitConfig.Enabled {
		router.Use(middleware.RateLimitMiddleware(rateLimiter, logger))
	}

	// Health check endpoints
	router.GET("/healthz", healthHandler.Health)
	router.GET("/readyz", healthHandler.Ready)

	// API routes
	api := router.Group("/api/v1")
	// Profile service routes
	api.Any("/users", profileHandler.Forward)
	api.Any("/users/*path", profileHandler.Forward)
	api.Any("/languages", profileHandler.Forward)
	api.Any("/languages/*path", profileHandler.Forward)
	api.Any("/interests", profileHandler.Forward)
	api.Any("/interests/*path", profileHandler.Forward)
	api.Any("/preferences", profileHandler.Forward)
	api.Any("/preferences/*path", profileHandler.Forward)
	api.Any("/traits", profileHandler.Forward)
	api.Any("/traits/*path", profileHandler.Forward)
	api.Any("/availability", profileHandler.Forward)
	api.Any("/availability/*path", profileHandler.Forward)

	// Bot service routes
	api.Any("/bot", botHandler.Forward)
	api.Any("/bot/*path", botHandler.Forward)
	api.Any("/telegram", botHandler.Forward)
	api.Any("/telegram/*path", botHandler.Forward)
	api.Any("/discord", botHandler.Forward)
	api.Any("/discord/*path", botHandler.Forward)

	// Start server
	srv := server.New(cfg.HTTPPort, router)
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()
	logger.Info("API Gateway started", zap.String("port", cfg.HTTPPort))

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("API Gateway stopped")
}
