package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matcher/internal/config"
	"matcher/internal/db"
	"matcher/internal/server"
)

func main() {
	cfg := config.LoadMatcher()

	pool, err := db.Connect(ctxWithTimeout(10*time.Second), cfg)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	// Run migrations (matching schema: match_queue, tasks)
	if err := db.RunMigrations(cfg); err != nil {
		pool.Close() // Close pool before fatal exit
		log.Fatalf("migrations error: %v", err)
	}

	// HTTP server
	srv := server.New(cfg.HTTPPort, pool)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()
	log.Printf("matcher service is up on :%s", cfg.HTTPPort)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Printf("matcher service stopped")
}

func ctxWithTimeout(d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return ctx
}
