package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type mockTask struct {
	executeFunc func(ctx context.Context) error
	calls       int32
}

func (m *mockTask) Execute(ctx context.Context) error {
	atomic.AddInt32(&m.calls, 1)
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil
}

func TestPool_SubmitAndExecute(t *testing.T) {
	pool := NewPool(2, 5)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	task := &mockTask{}

	if err := pool.Submit(task); err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	// Wait briefly for worker to process
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&task.calls) != 1 {
		t.Errorf("expected task to be executed exactly once, got %d", task.calls)
	}

	pool.Stop()
}

func TestPool_QueueFull(t *testing.T) {
	pool := NewPool(1, 1) // 1 worker, queue size 1

	// Submit tasks to fill the queue without starting the pool
	err1 := pool.Submit(&mockTask{})
	if err1 != nil {
		t.Fatalf("expected first submit to succeed, got %v", err1)
	}

	// Queue is full, submit should fail
	err2 := pool.Submit(&mockTask{})
	if err2 == nil {
		t.Fatalf("expected submit to fail because queue is full")
	}
}

func TestPool_SubmitAfterStop(t *testing.T) {
	pool := NewPool(1, 1)
	pool.Stop()

	err := pool.Submit(&mockTask{})
	if err == nil {
		t.Fatalf("expected submit to fail because pool is stopped")
	}
	if err.Error() != "pool is stopped" {
		t.Errorf("expected pool is stopped error, got %v", err)
	}
}
