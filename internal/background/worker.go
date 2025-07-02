package background

import (
	"context"
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"sync"
	"time"
)

type BackgroundWorker struct {
	jwtService auth.JWTService
	logger     logger.StructuredLogger
	ticker     *time.Ticker
	done       chan bool
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

type WorkerConfig struct {
	CleanupInterval time.Duration
	JWTService      auth.JWTService
	Logger          logger.StructuredLogger
}

func NewBackgroundWorker(config WorkerConfig) *BackgroundWorker {
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	return &BackgroundWorker{
		jwtService: config.JWTService,
		logger:     config.Logger,
		ticker:     time.NewTicker(config.CleanupInterval),
		done:       make(chan bool),
	}
}

func (w *BackgroundWorker) Start(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		w.logger.Warn("Background worker already running")
		return
	}

	w.running = true
	w.wg.Add(1)

	w.logger.Info("Starting background worker for session cleanup",
		"cleanup_interval", w.ticker.C)

	go func() {
		defer w.wg.Done()
		defer w.logger.Info("Background worker stopped")

		w.cleanupExpiredSessions(ctx)

		for {
			select {
			case <-w.ticker.C:
				w.cleanupExpiredSessions(ctx)
			case <-w.done:
				w.logger.Info("Background worker received stop signal")
				return
			case <-ctx.Done():
				w.logger.Info("Background worker context cancelled")
				return
			}
		}
	}()

	w.logger.Info("Background worker started successfully")
}

func (w *BackgroundWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.logger.Info("Stopping background worker...")

	w.running = false
	w.ticker.Stop()
	close(w.done)
	w.wg.Wait()

	w.logger.Info("Background worker stopped gracefully")
}

func (w *BackgroundWorker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

func (w *BackgroundWorker) cleanupExpiredSessions(ctx context.Context) {
	start := time.Now()

	w.logger.Debug("Starting session cleanup", "timestamp", start)

	err := w.jwtService.CleanupExpiredSessions(ctx)

	duration := time.Since(start)

	if err != nil {
		w.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
			Event:    "session_cleanup_failed",
			Entity:   "session",
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
		})

		w.logger.Error("Failed to cleanup expired sessions",
			"error", err,
			"duration", duration)
		return
	}

	w.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
		Event:    "session_cleanup_completed",
		Entity:   "session",
		Success:  true,
		Duration: duration,
	})

	w.logger.Debug("Session cleanup completed successfully",
		"duration", duration)
}

func (w *BackgroundWorker) cleanupWithMetrics(ctx context.Context) error {
	start := time.Now()

	err := w.jwtService.CleanupExpiredSessions(ctx)

	duration := time.Since(start)

	w.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
		Event:    "session_cleanup_with_metrics",
		Entity:   "session",
		Success:  err == nil,
		Duration: duration,
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
		Details: map[string]interface{}{
			"cleanup_duration_ms": duration.Milliseconds(),
			"cleanup_timestamp":   start.Unix(),
		},
	})

	return err
}
