package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/affordmed/logging_middleware/internal/platform/auth"
	"github.com/affordmed/logging_middleware/internal/platform/config"
	"github.com/affordmed/logging_middleware/internal/platform/logger"
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

	tokenManager := auth.NewTokenManager(authClient, authReq)

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
	}

	logService := logger.NewLogger(loggerCfg)
	logService.Start(ctx)
	defer logService.Stop()

	// In a real app we'd start the HTTP/gRPC servers here
	// and inject logService into them.

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
}
