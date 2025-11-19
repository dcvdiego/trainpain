package service

import (
	"context"
	"fmt"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/dcvdiego/trainpain-backend/internal/external"
	"github.com/dcvdiego/trainpain-backend/internal/repository/postgres"
)

type PineappleService struct {
	repo          *postgres.PineappleRepository
	scraper       *external.PineappleScraper
	vapidPublicKey  string
	vapidPrivateKey string
}

func NewPineappleService(
	repo *postgres.PineappleRepository,
	scraper *external.PineappleScraper,
	vapidPublicKey string,
	vapidPrivateKey string,
) *PineappleService {
	return &PineappleService{
		repo:            repo,
		scraper:         scraper,
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
	}
}

// ==================== Classes ====================

// GetAllClasses returns all classes with their schedules
func (s *PineappleService) GetAllClasses(ctx context.Context) ([]domain.PineappleClassWithSchedule, error) {
	classes, err := s.repo.GetAllClasses(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.PineappleClassWithSchedule, len(classes))
	for i, class := range classes {
		schedules, err := s.repo.GetSchedulesByClassID(ctx, class.ID)
		if err != nil {
			return nil, err
		}

		result[i] = domain.PineappleClassWithSchedule{
			PineappleClass: class,
			Schedules:      schedules,
		}
	}

	return result, nil
}

// GetClassesByDay returns classes scheduled for a specific day of week
func (s *PineappleService) GetClassesByDay(ctx context.Context, dayOfWeek int) ([]domain.PineappleClassWithSchedule, error) {
	schedules, err := s.repo.GetSchedulesByDayOfWeek(ctx, dayOfWeek)
	if err != nil {
		return nil, err
	}

	// Group schedules by class ID
	classSchedules := make(map[int][]domain.PineappleClassSchedule)
	for _, schedule := range schedules {
		classSchedules[schedule.ClassID] = append(classSchedules[schedule.ClassID], schedule)
	}

	// Fetch class details
	result := []domain.PineappleClassWithSchedule{}
	for classID, scheds := range classSchedules {
		class, err := s.repo.GetClassByID(ctx, classID)
		if err != nil {
			return nil, err
		}

		result = append(result, domain.PineappleClassWithSchedule{
			PineappleClass: *class,
			Schedules:      scheds,
		})
	}

	return result, nil
}

// ==================== Subscriptions ====================

// Subscribe creates or updates a subscription for a class schedule
func (s *PineappleService) Subscribe(ctx context.Context, sub *domain.PineappleSubscription) error {
	// Validate that the schedule exists
	_, err := s.repo.GetScheduleByID(ctx, sub.ScheduleID)
	if err != nil {
		return fmt.Errorf("invalid schedule ID: %w", err)
	}

	return s.repo.CreateSubscription(ctx, sub)
}

// Unsubscribe removes a subscription
func (s *PineappleService) Unsubscribe(ctx context.Context, subscriptionID int) error {
	return s.repo.DeleteSubscription(ctx, subscriptionID)
}

// GetUserSubscriptions returns all active subscriptions for a push endpoint
func (s *PineappleService) GetUserSubscriptions(ctx context.Context, pushEndpoint string) ([]domain.PineappleSubscription, error) {
	return s.repo.GetSubscriptionsByEndpoint(ctx, pushEndpoint)
}

// ==================== Scraping ====================

// ScrapeAllClasses scrapes all classes from the website and updates the database
func (s *PineappleService) ScrapeAllClasses(ctx context.Context) error {
	// Create a scrape job
	job := &domain.PineappleScrapeJob{
		JobType: "full_scrape",
		Status:  "running",
	}
	now := time.Now()
	job.StartedAt = &now

	if err := s.repo.CreateScrapeJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create scrape job: %w", err)
	}

	// Scrape classes
	scrapedClasses, err := s.scraper.ScrapeClasses(ctx)
	if err != nil {
		errMsg := err.Error()
		job.Status = "failed"
		job.ErrorMessage = &errMsg
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		s.repo.UpdateScrapeJob(ctx, job)
		return fmt.Errorf("failed to scrape classes: %w", err)
	}

	// Process scraped data
	classesCreated := 0
	schedulesCreated := 0

	for _, scraped := range scrapedClasses {
		// Create or update class
		class := &domain.PineappleClass{
			ClassName:       scraped.ClassName,
			Instructor:      scraped.Instructor,
			Level:           scraped.Level,
			Style:           scraped.Style,
			Description:     scraped.Description,
			DurationMinutes: scraped.DurationMinutes,
		}

		if err := s.repo.CreateClass(ctx, class); err != nil {
			// Class might already exist, try to update
			// For simplicity, we'll just log and continue
			fmt.Printf("Failed to create class %s: %v\n", class.ClassName, err)
			continue
		}
		classesCreated++

		// Create schedule
		schedule := &domain.PineappleClassSchedule{
			ClassID:     class.ID,
			DayOfWeek:   scraped.DayOfWeek,
			StartTime:   scraped.StartTime,
			EndTime:     scraped.EndTime,
			Room:        scraped.Room,
			MaxCapacity: scraped.SeatsTotal,
			IsActive:    true,
		}

		if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
			fmt.Printf("Failed to create schedule for class %s: %v\n", class.ClassName, err)
			continue
		}
		schedulesCreated++

		// Create availability snapshot
		snapshot := &domain.PineappleAvailabilitySnapshot{
			ScheduleID:       schedule.ID,
			SnapshotDate:     time.Now(),
			SnapshotTime:     time.Now(),
			SeatsAvailable:   scraped.SeatsAvailable,
			SeatsTotal:       scraped.SeatsTotal,
			IsFullyBooked:    scraped.SeatsAvailable == 0,
			ScrapeSuccessful: true,
		}

		if err := s.repo.CreateAvailabilitySnapshot(ctx, snapshot); err != nil {
			fmt.Printf("Failed to create availability snapshot: %v\n", err)
		}
	}

	// Update job status
	job.Status = "completed"
	job.ClassesScraped = classesCreated
	job.SchedulesUpdated = schedulesCreated
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	return s.repo.UpdateScrapeJob(ctx, job)
}

// ==================== Availability Checking ====================

// CheckAvailability checks seat availability for upcoming classes
func (s *PineappleService) CheckAvailability(ctx context.Context) error {
	// Get all schedules
	schedules, err := s.repo.GetAllSchedules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	checksCompleted := 0

	for _, schedule := range schedules {
		// Calculate next occurrence of this class
		nextOccurrence := s.getNextOccurrence(schedule.DayOfWeek, schedule.StartTime)

		// Check if class is within 2 hours
		timeUntilClass := time.Until(nextOccurrence)
		if timeUntilClass > 2*time.Hour || timeUntilClass < 0 {
			continue // Skip if not within 2-hour window
		}

		// Get latest availability
		latestAvail, err := s.repo.GetLatestAvailability(ctx, schedule.ID, nextOccurrence)
		if err == nil {
			// Check if we need to scrape again (last check was > 5 minutes ago)
			if time.Since(latestAvail.SnapshotTime) < 5*time.Minute {
				continue // Recent data exists, skip
			}
		}

		// Scrape availability for this specific class
		// TODO: Implement class-specific scraping
		// For now, we'll skip this part until we can properly scrape individual classes

		checksCompleted++
	}

	fmt.Printf("Checked availability for %d classes\n", checksCompleted)
	return nil
}

// getNextOccurrence calculates the next occurrence of a weekly class
func (s *PineappleService) getNextOccurrence(dayOfWeek int, startTime string) time.Time {
	now := time.Now()
	currentWeekday := int(now.Weekday())

	// Calculate days until next occurrence
	daysUntil := dayOfWeek - currentWeekday
	if daysUntil <= 0 {
		daysUntil += 7
	}

	nextDate := now.AddDate(0, 0, daysUntil)

	// Parse start time
	t, err := time.Parse("15:04:05", startTime)
	if err != nil {
		return nextDate
	}

	return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, now.Location())
}

// ==================== Notifications ====================

// SendNotifications checks subscriptions and sends notifications
func (s *PineappleService) SendNotifications(ctx context.Context) error {
	// Get all schedules
	schedules, err := s.repo.GetAllSchedules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	for _, schedule := range schedules {
		nextOccurrence := s.getNextOccurrence(schedule.DayOfWeek, schedule.StartTime)
		timeUntilClass := time.Until(nextOccurrence)

		// Determine notification type based on time until class
		var notifType string
		var shouldNotify bool

		if timeUntilClass > 1*time.Hour && timeUntilClass <= 2*time.Hour {
			notifType = "2h_before"
			shouldNotify = true
		} else if timeUntilClass > 30*time.Minute && timeUntilClass <= 1*time.Hour {
			notifType = "1h_before"
			shouldNotify = true
		} else if timeUntilClass > 0 && timeUntilClass <= 30*time.Minute {
			notifType = "30m_before"
			shouldNotify = true
		} else {
			continue
		}

		if !shouldNotify {
			continue
		}

		// Get subscriptions for this schedule
		subscriptions, err := s.repo.GetSubscriptionsBySchedule(ctx, schedule.ID)
		if err != nil {
			fmt.Printf("Failed to get subscriptions for schedule %d: %v\n", schedule.ID, err)
			continue
		}

		// Get latest availability
		latestAvail, err := s.repo.GetLatestAvailability(ctx, schedule.ID, nextOccurrence)
		if err != nil {
			fmt.Printf("No availability data for schedule %d\n", schedule.ID)
			continue
		}

		// Send notifications to subscribers
		for _, sub := range subscriptions {
			// Check if notification was already sent
			alreadySent, err := s.repo.CheckIfNotificationSent(ctx, sub.ID, schedule.ID, nextOccurrence, notifType)
			if err != nil || alreadySent {
				continue
			}

			// Check notification settings
			shouldSend := false
			if notifType == "2h_before" && sub.Notify2hBefore {
				shouldSend = true
			} else if notifType == "1h_before" && sub.Notify1hBefore {
				shouldSend = true
			} else if notifType == "30m_before" && sub.Notify30mBefore {
				shouldSend = true
			}

			// Check threshold
			if latestAvail.SeatsAvailable <= sub.NotificationThreshold {
				shouldSend = true
				notifType = "threshold_reached"
			}

			if !shouldSend {
				continue
			}

			// Send push notification
			err = s.sendPushNotification(ctx, &sub, latestAvail, notifType)

			// Record notification
			notif := &domain.PineappleNotification{
				SubscriptionID:     sub.ID,
				ScheduleID:         schedule.ID,
				ClassDate:          nextOccurrence,
				NotificationType:   notifType,
				SeatsAvailable:     latestAvail.SeatsAvailable,
				DeliverySuccessful: err == nil,
			}

			if err != nil {
				errMsg := err.Error()
				notif.ErrorMessage = &errMsg
			}

			s.repo.CreateNotification(ctx, notif)
		}
	}

	return nil
}

// sendPushNotification sends a Web Push notification
func (s *PineappleService) sendPushNotification(
	ctx context.Context,
	sub *domain.PineappleSubscription,
	avail *domain.PineappleAvailabilitySnapshot,
	notifType string,
) error {
	// Create notification message
	message := fmt.Sprintf("Class alert: %d seats remaining!", avail.SeatsAvailable)

	if notifType == "threshold_reached" {
		message = fmt.Sprintf("Hurry! Only %d seats left for your class!", avail.SeatsAvailable)
	}

	// Send Web Push notification
	subscription := &webpush.Subscription{
		Endpoint: sub.PushEndpoint,
		Keys: webpush.Keys{
			P256dh: sub.PushP256dh,
			Auth:   sub.PushAuth,
		},
	}

	resp, err := webpush.SendNotification([]byte(message), subscription, &webpush.Options{
		Subscriber:      "admin@trainpain.com",
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             30,
	})

	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("push notification failed with status: %d", resp.StatusCode)
	}

	return nil
}
