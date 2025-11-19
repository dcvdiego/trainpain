package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/dcvdiego/trainpain-backend/internal/service"
	"go.uber.org/zap"
)

// PineappleScheduler handles periodic tasks for Pineapple classes
type PineappleScheduler struct {
	service *service.PineappleService
	logger  *zap.Logger
	stopCh  chan struct{}
}

// NewPineappleScheduler creates a new scheduler
func NewPineappleScheduler(service *service.PineappleService, logger *zap.Logger) *PineappleScheduler {
	return &PineappleScheduler{
		service: service,
		logger:  logger,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *PineappleScheduler) Start(ctx context.Context) {
	s.logger.Info("starting Pineapple scheduler")

	// Check availability every 5 minutes
	availabilityTicker := time.NewTicker(5 * time.Minute)
	defer availabilityTicker.Stop()

	// Send notifications every 1 minute (to catch time windows)
	notificationTicker := time.NewTicker(1 * time.Minute)
	defer notificationTicker.Stop()

	// Scrape all classes once per day at 3 AM
	scrapeTicker := time.NewTicker(24 * time.Hour)
	defer scrapeTicker.Stop()

	// Run initial scrape in background
	go func() {
		if err := s.service.ScrapeAllClasses(ctx); err != nil {
			s.logger.Error("initial scrape failed", zap.Error(err))
		}
	}()

	for {
		select {
		case <-availabilityTicker.C:
			s.logger.Info("checking class availability")
			if err := s.service.CheckAvailability(ctx); err != nil {
				s.logger.Error("failed to check availability", zap.Error(err))
			}

		case <-notificationTicker.C:
			s.logger.Debug("checking notifications")
			if err := s.service.SendNotifications(ctx); err != nil {
				s.logger.Error("failed to send notifications", zap.Error(err))
			}

		case <-scrapeTicker.C:
			s.logger.Info("running daily class scrape")
			if err := s.service.ScrapeAllClasses(ctx); err != nil {
				s.logger.Error("daily scrape failed", zap.Error(err))
			}

		case <-s.stopCh:
			s.logger.Info("stopping Pineapple scheduler")
			return

		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping scheduler")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *PineappleScheduler) Stop() {
	close(s.stopCh)
}

// calculateTimeUntilNextRun calculates time until next run at targetHour
func calculateTimeUntilNextRun(targetHour int) time.Duration {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, now.Location())

	// If we've passed the target hour today, schedule for tomorrow
	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	return time.Until(nextRun)
}

// RunDailyScrapeAtTime runs the scrape job daily at a specific time
func (s *PineappleScheduler) RunDailyScrapeAtTime(ctx context.Context, targetHour int) {
	for {
		// Calculate time until next run
		timeUntilRun := calculateTimeUntilNextRun(targetHour)
		s.logger.Info(fmt.Sprintf("next daily scrape scheduled in %v", timeUntilRun))

		select {
		case <-time.After(timeUntilRun):
			s.logger.Info("running scheduled daily scrape")
			if err := s.service.ScrapeAllClasses(ctx); err != nil {
				s.logger.Error("scheduled scrape failed", zap.Error(err))
			}

		case <-s.stopCh:
			s.logger.Info("stopping daily scrape scheduler")
			return

		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping daily scrape")
			return
		}
	}
}
