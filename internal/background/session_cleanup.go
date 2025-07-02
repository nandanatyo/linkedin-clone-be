package background

import (
	"context"
	"fmt"
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type SessionCleanupService struct {
	jwtService auth.JWTService
	logger     logger.StructuredLogger
	ticker     *time.Ticker
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool

	totalRuns       int64
	failedRuns      int64
	lastRunTime     time.Time
	lastRunStatus   string
	cleanupInterval time.Duration
}

func NewSessionCleanupService(jwtService auth.JWTService, logger logger.StructuredLogger) *SessionCleanupService {
	return &SessionCleanupService{
		jwtService:    jwtService,
		logger:        logger,
		stopChan:      make(chan struct{}),
		lastRunStatus: "never_run",
	}
}

func (s *SessionCleanupService) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.logger.Warn("Session cleanup service already running")
		return
	}

	s.cleanupInterval = s.getCleanupInterval()

	s.ticker = time.NewTicker(s.cleanupInterval)
	s.running = true
	s.wg.Add(1)

	s.logger.Info("Starting session cleanup service",
		"cleanup_interval", s.cleanupInterval.String())

	go func() {
		defer s.wg.Done()
		defer s.logger.Info("Session cleanup service stopped")

		s.performCleanup(ctx)

		for {
			select {
			case <-s.ticker.C:
				s.performCleanup(ctx)
			case <-s.stopChan:
				s.logger.Info("Session cleanup service received stop signal")
				return
			case <-ctx.Done():
				s.logger.Info("Session cleanup service context cancelled")
				return
			}
		}
	}()

	s.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
		Event:   "session_cleanup_service_started",
		Entity:  "service",
		Success: true,
		Details: map[string]interface{}{
			"cleanup_interval_minutes": s.cleanupInterval.Minutes(),
		},
	})
}

func (s *SessionCleanupService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping session cleanup service...")

	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	s.wg.Wait()

	s.logger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "session_cleanup_service_stopped",
		Entity:  "service",
		Success: true,
		Details: map[string]interface{}{
			"total_runs":     atomic.LoadInt64(&s.totalRuns),
			"failed_runs":    atomic.LoadInt64(&s.failedRuns),
			"uptime_minutes": time.Since(s.lastRunTime).Minutes(),
		},
	})
}

func (s *SessionCleanupService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *SessionCleanupService) GetMetrics() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextRun := s.lastRunTime.Add(s.cleanupInterval)
	if s.lastRunTime.IsZero() {
		nextRun = time.Now().Add(s.cleanupInterval)
	}

	return map[string]interface{}{
		"status": func() string {
			if s.running {
				return "running"
			} else {
				return "stopped"
			}
		}(),
		"total_runs":               atomic.LoadInt64(&s.totalRuns),
		"failed_runs":              atomic.LoadInt64(&s.failedRuns),
		"success_rate":             s.calculateSuccessRate(),
		"last_run":                 s.lastRunTime.Format(time.RFC3339),
		"last_run_status":          s.lastRunStatus,
		"next_run":                 nextRun.Format(time.RFC3339),
		"cleanup_interval":         s.cleanupInterval.String(),
		"cleanup_interval_minutes": s.cleanupInterval.Minutes(),
	}
}

func (s *SessionCleanupService) performCleanup(ctx context.Context) {
	start := time.Now()
	atomic.AddInt64(&s.totalRuns, 1)

	s.logger.Debug("Starting session cleanup task",
		"run_number", atomic.LoadInt64(&s.totalRuns))

	err := s.jwtService.CleanupExpiredSessions(ctx)

	duration := time.Since(start)
	s.mu.Lock()
	s.lastRunTime = start
	s.mu.Unlock()

	if err != nil {
		atomic.AddInt64(&s.failedRuns, 1)
		s.mu.Lock()
		s.lastRunStatus = "failed"
		s.mu.Unlock()

		s.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
			Event:    "session_cleanup_failed",
			Entity:   "session",
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Details: map[string]interface{}{
				"run_number":        atomic.LoadInt64(&s.totalRuns),
				"total_failed_runs": atomic.LoadInt64(&s.failedRuns),
			},
		})

		s.logger.Error("Session cleanup failed",
			"error", err,
			"duration", duration,
			"run_number", atomic.LoadInt64(&s.totalRuns))
		return
	}

	s.mu.Lock()
	s.lastRunStatus = "success"
	s.mu.Unlock()

	s.logger.LogBusinessEvent(ctx, logger.BusinessEventLog{
		Event:    "session_cleanup_completed",
		Entity:   "session",
		Success:  true,
		Duration: duration,
		Details: map[string]interface{}{
			"cleanup_duration_ms": duration.Milliseconds(),
			"run_number":          atomic.LoadInt64(&s.totalRuns),
			"success_rate":        s.calculateSuccessRate(),
		},
	})

	s.logger.Debug("Session cleanup completed successfully",
		"duration", duration,
		"run_number", atomic.LoadInt64(&s.totalRuns),
		"success_rate", s.calculateSuccessRate())
}

func (s *SessionCleanupService) calculateSuccessRate() float64 {
	totalRuns := atomic.LoadInt64(&s.totalRuns)
	failedRuns := atomic.LoadInt64(&s.failedRuns)

	if totalRuns == 0 {
		return 0.0
	}

	successRuns := totalRuns - failedRuns
	return (float64(successRuns) / float64(totalRuns)) * 100.0
}

func (s *SessionCleanupService) getCleanupInterval() time.Duration {

	defaultInterval := 5 * time.Minute

	envValue := os.Getenv("SESSION_CLEANUP_INTERVAL_MINUTES")
	if envValue == "" {
		return defaultInterval
	}

	minutes, err := strconv.Atoi(envValue)
	if err != nil {
		s.logger.Warn("Invalid SESSION_CLEANUP_INTERVAL_MINUTES value, using default",
			"env_value", envValue,
			"default_minutes", int(defaultInterval.Minutes()))
		return defaultInterval
	}

	if minutes <= 0 {
		s.logger.Warn("SESSION_CLEANUP_INTERVAL_MINUTES must be positive, using default",
			"env_value", minutes,
			"default_minutes", int(defaultInterval.Minutes()))
		return defaultInterval
	}

	interval := time.Duration(minutes) * time.Minute
	s.logger.Info("Using custom cleanup interval from environment",
		"interval_minutes", minutes,
		"interval", interval.String())

	return interval
}

func (s *SessionCleanupService) ForceCleanup(ctx context.Context) error {
	if !s.IsRunning() {
		return fmt.Errorf("cleanup service is not running")
	}

	s.logger.Info("Manual cleanup triggered")

	go func() {
		s.performCleanup(ctx)
	}()

	return nil
}
