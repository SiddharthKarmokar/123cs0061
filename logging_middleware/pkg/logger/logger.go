package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/affordmed/logging_middleware/pkg/worker"
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

	task := &worker.LogTask{
		URL:        l.config.EvaluationURL,
		Payload:    payload,
		MaxRetries: l.config.MaxRetries,
		HTTPClient: l.httpClient,
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
