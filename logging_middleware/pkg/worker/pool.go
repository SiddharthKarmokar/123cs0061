package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Task represents a unit of work for the worker pool.
type Task interface {
	Execute(ctx context.Context) error
}

// LogTask is a specific task for sending logs to the evaluation service.
type LogTask struct {
	URL        string
	Payload    interface{}
	MaxRetries int
	HTTPClient *http.Client
	OnSuccess  func(ctx context.Context)
	OnFailure  func(ctx context.Context, err error)
}

func (t *LogTask) Execute(ctx context.Context) error {
	data, err := json.Marshal(t.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal log payload: %w", err)
	}

	backoff := 100 * time.Millisecond
	var lastErr error

	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.URL, bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := t.HTTPClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Success
				if t.OnSuccess != nil {
					t.OnSuccess(ctx)
				}
				return nil
			}
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		// Exponential backoff
		if attempt < t.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	if t.OnFailure != nil {
		t.OnFailure(ctx, lastErr)
	}
	return fmt.Errorf("max retries reached, last error: %w", lastErr)
}

// Pool manages a group of workers processing tasks from a queue.
type Pool struct {
	tasks      chan Task
	wg         sync.WaitGroup
	quit       chan struct{}
	maxWorkers int
}

// NewPool creates a new worker pool.
func NewPool(maxWorkers, queueSize int) *Pool {
	return &Pool{
		tasks:      make(chan Task, queueSize),
		quit:       make(chan struct{}),
		maxWorkers: maxWorkers,
	}
}

// Start initializes the workers and starts processing tasks.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case task := <-p.tasks:
					// Execute the task. Errors could be logged to a dead-letter queue in the future.
					_ = task.Execute(ctx)
				case <-p.quit:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Submit adds a new task to the queue. Does not block unless the queue is full.
func (p *Pool) Submit(task Task) error {
	select {
	case <-p.quit:
		return fmt.Errorf("pool is stopped")
	default:
	}

	select {
	case <-p.quit:
		return fmt.Errorf("pool is stopped")
	case p.tasks <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Stop gracefully shuts down the worker pool.
func (p *Pool) Stop() {
	close(p.quit)
	p.wg.Wait()
}
