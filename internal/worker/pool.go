package worker

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Task represents a work task
type Task func(ctx context.Context) error

// Pool represents a worker pool
type Pool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
}

// NewPool creates a new worker pool
func NewPool(workers, queueSize int, logger *zap.Logger) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Pool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	p.logger.Info("starting worker pool", zap.Int("workers", p.workers))
	
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is a single worker goroutine
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	
	p.logger.Debug("worker started", zap.Int("worker_id", id))
	
	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("worker stopped", zap.Int("worker_id", id))
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				p.logger.Debug("task queue closed, worker exiting", zap.Int("worker_id", id))
				return
			}
			
			// Execute task
			if err := task(p.ctx); err != nil {
				p.logger.Error("task execution failed",
					zap.Int("worker_id", id),
					zap.Error(err),
				)
			}
		}
	}
}

// Submit submits a task to the pool
func (p *Pool) Submit(task Task) error {
	select {
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	case p.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Stop gracefully stops the worker pool
func (p *Pool) Stop() {
	p.logger.Info("stopping worker pool")
	
	// Close task queue
	close(p.taskQueue)
	
	// Cancel context to signal workers
	p.cancel()
	
	// Wait for all workers to finish
	p.wg.Wait()
	
	p.logger.Info("worker pool stopped")
}

// WaitWithTimeout waits for all workers with a timeout
func (p *Pool) WaitWithTimeout(ctx context.Context) error {
	done := make(chan struct{})
	
	go func() {
		p.wg.Wait()
		close(done)
	}()
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

