package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/affordmed/logging_middleware/pkg/auth"
	"github.com/affordmed/logging_middleware/pkg/config"
)

func main() {
	fmt.Println("Starting Stage 0 Background Worker...")

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Auth Module for worker
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

	// Note: The worker currently does not use the authTransport natively.
	// It relies on tokenManager for manual token retrieval or outbox publishing later.
	// The outbox publisher would use authTransport for external syncs if needed.

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker gracefully...")
}
