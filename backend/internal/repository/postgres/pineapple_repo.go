package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type PineappleRepository struct {
	db *sqlx.DB
}

func NewPineappleRepository(db *sqlx.DB) *PineappleRepository {
	return &PineappleRepository{db: db}
}

// ==================== Classes ====================

func (r *PineappleRepository) CreateClass(ctx context.Context, class *domain.PineappleClass) error {
	query := `
		INSERT INTO pineapple_classes (class_name, instructor, level, style, description, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		class.ClassName, class.Instructor, class.Level, class.Style,
		class.Description, class.DurationMinutes,
	).Scan(&class.ID, &class.CreatedAt, &class.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	return nil
}

func (r *PineappleRepository) GetClassByID(ctx context.Context, id int) (*domain.PineappleClass, error) {
	var class domain.PineappleClass
	query := `SELECT * FROM pineapple_classes WHERE id = $1`

	err := r.db.GetContext(ctx, &class, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class by ID: %w", err)
	}

	return &class, nil
}

func (r *PineappleRepository) GetAllClasses(ctx context.Context) ([]domain.PineappleClass, error) {
	var classes []domain.PineappleClass
	query := `SELECT * FROM pineapple_classes ORDER BY class_name`

	err := r.db.SelectContext(ctx, &classes, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all classes: %w", err)
	}

	return classes, nil
}

func (r *PineappleRepository) UpdateClass(ctx context.Context, class *domain.PineappleClass) error {
	query := `
		UPDATE pineapple_classes
		SET class_name = $1, instructor = $2, level = $3, style = $4,
			description = $5, duration_minutes = $6, updated_at = NOW()
		WHERE id = $7
	`

	_, err := r.db.ExecContext(ctx, query,
		class.ClassName, class.Instructor, class.Level, class.Style,
		class.Description, class.DurationMinutes, class.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update class: %w", err)
	}

	return nil
}

func (r *PineappleRepository) DeleteClass(ctx context.Context, id int) error {
	query := `DELETE FROM pineapple_classes WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return nil
}

// ==================== Schedules ====================

func (r *PineappleRepository) CreateSchedule(ctx context.Context, schedule *domain.PineappleClassSchedule) error {
	query := `
		INSERT INTO pineapple_class_schedules (class_id, day_of_week, start_time, end_time, room, max_capacity, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		schedule.ClassID, schedule.DayOfWeek, schedule.StartTime, schedule.EndTime,
		schedule.Room, schedule.MaxCapacity, schedule.IsActive,
	).Scan(&schedule.ID, &schedule.CreatedAt, &schedule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

func (r *PineappleRepository) GetScheduleByID(ctx context.Context, id int) (*domain.PineappleClassSchedule, error) {
	var schedule domain.PineappleClassSchedule
	query := `SELECT * FROM pineapple_class_schedules WHERE id = $1`

	err := r.db.GetContext(ctx, &schedule, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule by ID: %w", err)
	}

	return &schedule, nil
}

func (r *PineappleRepository) GetSchedulesByClassID(ctx context.Context, classID int) ([]domain.PineappleClassSchedule, error) {
	var schedules []domain.PineappleClassSchedule
	query := `SELECT * FROM pineapple_class_schedules WHERE class_id = $1 AND is_active = true ORDER BY day_of_week, start_time`

	err := r.db.SelectContext(ctx, &schedules, query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules by class ID: %w", err)
	}

	return schedules, nil
}

func (r *PineappleRepository) GetSchedulesByDayOfWeek(ctx context.Context, dayOfWeek int) ([]domain.PineappleClassSchedule, error) {
	var schedules []domain.PineappleClassSchedule
	query := `SELECT * FROM pineapple_class_schedules WHERE day_of_week = $1 AND is_active = true ORDER BY start_time`

	err := r.db.SelectContext(ctx, &schedules, query, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules by day: %w", err)
	}

	return schedules, nil
}

func (r *PineappleRepository) GetAllSchedules(ctx context.Context) ([]domain.PineappleClassSchedule, error) {
	var schedules []domain.PineappleClassSchedule
	query := `SELECT * FROM pineapple_class_schedules WHERE is_active = true ORDER BY day_of_week, start_time`

	err := r.db.SelectContext(ctx, &schedules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all schedules: %w", err)
	}

	return schedules, nil
}

func (r *PineappleRepository) UpdateSchedule(ctx context.Context, schedule *domain.PineappleClassSchedule) error {
	query := `
		UPDATE pineapple_class_schedules
		SET day_of_week = $1, start_time = $2, end_time = $3, room = $4,
			max_capacity = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`

	_, err := r.db.ExecContext(ctx, query,
		schedule.DayOfWeek, schedule.StartTime, schedule.EndTime, schedule.Room,
		schedule.MaxCapacity, schedule.IsActive, schedule.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}

// ==================== Subscriptions ====================

func (r *PineappleRepository) CreateSubscription(ctx context.Context, sub *domain.PineappleSubscription) error {
	query := `
		INSERT INTO pineapple_subscriptions
		(schedule_id, push_endpoint, push_p256dh, push_auth, notification_threshold,
		 notify_2h_before, notify_1h_before, notify_30m_before, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (schedule_id, push_endpoint)
		DO UPDATE SET is_active = true, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		sub.ScheduleID, sub.PushEndpoint, sub.PushP256dh, sub.PushAuth,
		sub.NotificationThreshold, sub.Notify2hBefore, sub.Notify1hBefore,
		sub.Notify30mBefore, sub.IsActive,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *PineappleRepository) GetSubscriptionsBySchedule(ctx context.Context, scheduleID int) ([]domain.PineappleSubscription, error) {
	var subscriptions []domain.PineappleSubscription
	query := `SELECT * FROM pineapple_subscriptions WHERE schedule_id = $1 AND is_active = true`

	err := r.db.SelectContext(ctx, &subscriptions, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *PineappleRepository) GetSubscriptionsByEndpoint(ctx context.Context, endpoint string) ([]domain.PineappleSubscription, error) {
	var subscriptions []domain.PineappleSubscription
	query := `SELECT * FROM pineapple_subscriptions WHERE push_endpoint = $1 AND is_active = true`

	err := r.db.SelectContext(ctx, &subscriptions, query, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by endpoint: %w", err)
	}

	return subscriptions, nil
}

func (r *PineappleRepository) DeleteSubscription(ctx context.Context, id int) error {
	query := `UPDATE pineapple_subscriptions SET is_active = false, updated_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

// ==================== Availability Snapshots ====================

func (r *PineappleRepository) CreateAvailabilitySnapshot(ctx context.Context, snapshot *domain.PineappleAvailabilitySnapshot) error {
	query := `
		INSERT INTO pineapple_availability_snapshots
		(schedule_id, snapshot_date, snapshot_time, seats_available, seats_total,
		 is_fully_booked, scrape_successful, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		snapshot.ScheduleID, snapshot.SnapshotDate, snapshot.SnapshotTime,
		snapshot.SeatsAvailable, snapshot.SeatsTotal, snapshot.IsFullyBooked,
		snapshot.ScrapeSuccessful, snapshot.ErrorMessage,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create availability snapshot: %w", err)
	}

	return nil
}

func (r *PineappleRepository) GetLatestAvailability(ctx context.Context, scheduleID int, classDate time.Time) (*domain.PineappleAvailabilitySnapshot, error) {
	var snapshot domain.PineappleAvailabilitySnapshot
	query := `
		SELECT * FROM pineapple_availability_snapshots
		WHERE schedule_id = $1 AND snapshot_date = $2
		ORDER BY snapshot_time DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &snapshot, query, scheduleID, classDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest availability: %w", err)
	}

	return &snapshot, nil
}

// ==================== Notifications ====================

func (r *PineappleRepository) CreateNotification(ctx context.Context, notif *domain.PineappleNotification) error {
	query := `
		INSERT INTO pineapple_notifications
		(subscription_id, schedule_id, class_date, notification_type, seats_available,
		 delivery_successful, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, sent_at
	`

	err := r.db.QueryRowContext(ctx, query,
		notif.SubscriptionID, notif.ScheduleID, notif.ClassDate, notif.NotificationType,
		notif.SeatsAvailable, notif.DeliverySuccessful, notif.ErrorMessage,
	).Scan(&notif.ID, &notif.SentAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *PineappleRepository) CheckIfNotificationSent(ctx context.Context, subscriptionID int, scheduleID int, classDate time.Time, notifType string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM pineapple_notifications
		WHERE subscription_id = $1 AND schedule_id = $2
		  AND class_date = $3 AND notification_type = $4
	`

	err := r.db.GetContext(ctx, &count, query, subscriptionID, scheduleID, classDate, notifType)
	if err != nil {
		return false, fmt.Errorf("failed to check notification: %w", err)
	}

	return count > 0, nil
}

// ==================== Scrape Jobs ====================

func (r *PineappleRepository) CreateScrapeJob(ctx context.Context, job *domain.PineappleScrapeJob) error {
	query := `
		INSERT INTO pineapple_scrape_jobs (job_type, status)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, job.JobType, job.Status).Scan(&job.ID, &job.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create scrape job: %w", err)
	}

	return nil
}

func (r *PineappleRepository) UpdateScrapeJob(ctx context.Context, job *domain.PineappleScrapeJob) error {
	query := `
		UPDATE pineapple_scrape_jobs
		SET status = $1, classes_scraped = $2, schedules_updated = $3,
			availability_checks = $4, error_message = $5, started_at = $6,
			completed_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		job.Status, job.ClassesScraped, job.SchedulesUpdated,
		job.AvailabilityChecks, job.ErrorMessage, job.StartedAt,
		job.CompletedAt, job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update scrape job: %w", err)
	}

	return nil
}
