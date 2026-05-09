package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/affordmed/logging_middleware/pkg/auth"
	"github.com/affordmed/logging_middleware/pkg/config"
	"github.com/affordmed/logging_middleware/pkg/database"
	"github.com/affordmed/logging_middleware/pkg/logger"
	"github.com/affordmed/logging_middleware/pkg/outbox"
)

func main() {
	fmt.Println("Starting Stage 0 Logging API...")

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Postgres Database
	dbConfig := database.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DBName:   cfg.Postgres.DB,
	}
	db, err := database.Connect(ctx, dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run bare-minimum migration for Outbox if needed (For testing local setup)
	migrationQuery := `
		CREATE TABLE IF NOT EXISTS outbox_events (
			id UUID PRIMARY KEY,
			aggregate_id UUID NOT NULL,
			aggregate_type VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			payload JSONB NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			retries INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			processed_at TIMESTAMP WITH TIME ZONE
		);
	`
	if _, err := db.ExecContext(ctx, migrationQuery); err != nil {
		log.Fatalf("failed to run outbox migration: %v", err)
	}

	outboxRepo := outbox.NewPostgresRepository(db)

	// Initialize Auth Module
	authClient := auth.NewAuthClient(cfg.Auth.BaseURL)
	authReq := auth.AuthRequest{
		Email:        cfg.Auth.Email,
		Name:         cfg.Auth.Name,
		RollNo:       cfg.Auth.RollNo,
		AccessCode:   cfg.Auth.AccessCode,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
	}

	tokenManager := auth.NewTokenManager(authClient, authReq, cfg.Auth.UseStaticToken, cfg.Auth.StaticToken)

	// Start token manager background refresh
	tokenManager.Start(ctx)
	defer tokenManager.Stop()

	// Initialize Auth-Aware HTTP Client
	authTransport := &auth.Transport{
		Base:    http.DefaultTransport,
		Manager: tokenManager,
	}
	httpClient := &http.Client{
		Transport: authTransport,
	}

	// Initialize Logger
	loggerCfg := logger.Config{
		EvaluationURL: fmt.Sprintf("%s/logs", cfg.Auth.BaseURL),
		MaxWorkers:    cfg.Logger.MaxWorkers,
		QueueSize:     cfg.Logger.QueueSize,
		MaxRetries:    cfg.Logger.MaxRetries,
		HTTPClient:    httpClient,
		OutboxRepo:    outboxRepo,
	}

	logService := logger.NewLogger(loggerCfg)
	logService.Start(ctx)
	defer logService.Stop()

	// Set up simple HTTP server for testing
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/test-log", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Hardcoded log for testing
		err := logService.Log("backend", "info", "handler", "Test log from Postman")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to log: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Log queued successfully"))
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
