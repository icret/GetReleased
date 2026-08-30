package scheduler

import (
	"context"
	"time"
)

type Task func(ctx context.Context)

type Scheduler struct {
	interval time.Duration
	task     Task
}

func New(interval time.Duration, task Task) *Scheduler {
	return &Scheduler{interval: interval, task: task}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.task(ctx)
		}
	}
}
