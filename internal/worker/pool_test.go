package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWorkerPool(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("BasicExecution", func(t *testing.T) {
		pool := NewPool(2, 10, logger)
		pool.Start()
		defer pool.Stop()

		var counter int32
		
		for i := 0; i < 5; i++ {
			pool.Submit(func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})
		}

		time.Sleep(100 * time.Millisecond)

		if atomic.LoadInt32(&counter) != 5 {
			t.Errorf("Expected 5 tasks executed, got %d", counter)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		pool := NewPool(2, 10, logger)
		pool.Start()
		defer pool.Stop()

		pool.Submit(func(ctx context.Context) error {
			return errors.New("test error")
		})

		time.Sleep(100 * time.Millisecond)
		// Should not crash on error
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		pool := NewPool(2, 10, logger)
		pool.Start()

		pool.Submit(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		})

		pool.Stop()
		// Should cancel tasks on stop
	})
}

func BenchmarkWorkerPool(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	pool := NewPool(10, 1000, logger)
	pool.Start()
	defer pool.Stop()

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		pool.Submit(func(ctx context.Context) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		})
	}
}

