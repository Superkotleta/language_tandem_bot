package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/delivery/telegram"
	"language-exchange-bot/internal/repository"
	"language-exchange-bot/internal/service"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to Database
	dbPool, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 3. Initialize Layers
	userRepo := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepo)
	
	bot, err := telegram.NewBot(cfg.TelegramToken, userService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// 4. Graceful Shutdown Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	// 5. Start Bot
	log.Println("Starting application...")
	bot.Start(ctx)
	log.Println("Application stopped.")
}


