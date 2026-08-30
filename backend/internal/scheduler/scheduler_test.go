package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_RunsTask(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	s := New(10*time.Millisecond, func(ctx context.Context) {
		calls.Add(1)
	})
	s.Start(ctx)
	if calls.Load() < 1 {
		t.Fatalf("task not called, calls=%d", calls.Load())
	}
}

func TestScheduler_Cancel(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	s := New(10*time.Millisecond, func(ctx context.Context) {
		calls.Add(1)
	})
	go s.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	if calls.Load() < 1 {
		t.Fatalf("task not called before cancel, calls=%d", calls.Load())
	}
}
