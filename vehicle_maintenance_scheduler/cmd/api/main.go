package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/affordmed/logging_middleware/pkg/auth"
	"github.com/affordmed/logging_middleware/pkg/config"
	"github.com/affordmed/logging_middleware/pkg/logger"
	"github.com/affordmed/vehicle_maintenance_scheduler/internal/scheduler/domain"
	"github.com/affordmed/vehicle_maintenance_scheduler/internal/scheduler/optimizer"
)

func main() {
	fmt.Println("Starting Vehicle Maintenance Scheduler...")

	// Load Configuration using shared SDK
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Shared Auth Module
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
	tokenManager.Start(ctx)
	defer tokenManager.Stop()

	// Initialize Auth-Aware HTTP Client
	authTransport := &auth.Transport{
		Base:    http.DefaultTransport,
		Manager: tokenManager,
	}
	httpClient := &http.Client{
		Transport: authTransport,
		Timeout:   10 * time.Second,
	}

	// Initialize Shared Logger
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

	// Initialize Optimizer
	knapsackOptimizer := optimizer.NewKnapsackOptimizer()

	// Set up HTTP Server
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Scheduler OK"))
	})

	mux.HandleFunc("/schedule", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		_ = logService.Log("backend", "info", "handler", "Received scheduling request")

		// In a real scenario, fetch tasks from an external API using httpClient.
		// For now, we mock some depot/task data
		tasks := []domain.Task{
			{ID: "T1", Duration: 2 * time.Hour, ImpactScore: 50},
			{ID: "T2", Duration: 3 * time.Hour, ImpactScore: 60},
			{ID: "T3", Duration: 4 * time.Hour, ImpactScore: 80},
			{ID: "T4", Duration: 5 * time.Hour, ImpactScore: 90},
		}

		availableHours := 8 // Mocked depot capacity

		res := knapsackOptimizer.Optimize(tasks, availableHours)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)

		_ = logService.Log("backend", "info", "handler", fmt.Sprintf("Scheduling completed, impact: %d", res.TotalImpact))
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port, // Assume same env structure
		Handler: mux,
	}

	go func() {
		log.Printf("Starting Scheduler HTTP server on port %s", cfg.HTTP.Port)
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
