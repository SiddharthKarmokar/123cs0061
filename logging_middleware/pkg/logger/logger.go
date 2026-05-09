package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"

	"github.com/affordmed/logging_middleware/pkg/outbox"
	"github.com/affordmed/logging_middleware/pkg/worker"
	"github.com/google/uuid"
)

// Logger is the core interface for the logging platform.
type Logger interface {
	Start(ctx context.Context)
	Log(stack, level, pkg, message string) error
	Stop()
}

// Config holds the configuration for the Logger.
type Config struct {
	EvaluationURL string
	MaxWorkers    int
	QueueSize     int
	MaxRetries    int
	HTTPClient    *http.Client
	OutboxRepo    outbox.Repository
}

type defaultLogger struct {
	pool       *worker.Pool
	config     Config
	httpClient *http.Client
}

// NewLogger creates and initializes a new asynchronous Logger.
func NewLogger(cfg Config) Logger {
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 5
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1000
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	pool := worker.NewPool(cfg.MaxWorkers, cfg.QueueSize)

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	return &defaultLogger{
		pool:       pool,
		config:     cfg,
		httpClient: httpClient,
	}
}

// Start should be called to spin up the worker pool.
func (l *defaultLogger) Start(ctx context.Context) {
	l.pool.Start(ctx)
}

// Log validates the payload, constructs it, and submits it for async processing.
func (l *defaultLogger) Log(stack, level, pkg, message string) error {
	payload := &LogPayload{
		Stack:   stack,
		Level:   level,
		Package: pkg,
		Message: message,
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid log payload: %w", err)
	}

	var outboxID uuid.UUID
	if l.config.OutboxRepo != nil {
		outboxID = uuid.New()
		payloadBytes, _ := json.Marshal(payload)
		event := outbox.Event{
			ID:            outboxID,
			AggregateID:   uuid.New(),
			AggregateType: "Logger",
			EventType:     "LogCreated",
			Payload:       payloadBytes,
			Status:        "PENDING",
			Retries:       0,
			CreatedAt:     time.Now(),
		}
		if err := l.config.OutboxRepo.SaveEvent(context.Background(), event); err != nil {
			return fmt.Errorf("failed to save log to outbox: %w", err)
		}
	}

	task := &worker.LogTask{
		URL:        l.config.EvaluationURL,
		Payload:    payload,
		MaxRetries: l.config.MaxRetries,
		HTTPClient: l.httpClient,
		OnSuccess: func(ctx context.Context) {
			if l.config.OutboxRepo != nil && outboxID != uuid.Nil {
				_ = l.config.OutboxRepo.MarkProcessed(context.Background(), outboxID)
			}
		},
		OnFailure: func(ctx context.Context, err error) {
			if l.config.OutboxRepo != nil && outboxID != uuid.Nil {
				_ = l.config.OutboxRepo.MarkFailed(context.Background(), outboxID)
			}
		},
	}

	if err := l.pool.Submit(task); err != nil {
		return fmt.Errorf("failed to enqueue log task: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the logger and its internal worker pool.
func (l *defaultLogger) Stop() {
	l.pool.Stop()
}
