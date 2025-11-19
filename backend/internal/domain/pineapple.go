package domain

import "time"

// PineappleClass represents a dance class at Pineapple Studios
type PineappleClass struct {
	ID               int       `db:"id" json:"id"`
	ClassName        string    `db:"class_name" json:"class_name"`
	Instructor       string    `db:"instructor" json:"instructor"`
	Level            string    `db:"level" json:"level"`
	Style            string    `db:"style" json:"style"`
	Description      string    `db:"description" json:"description"`
	DurationMinutes  int       `db:"duration_minutes" json:"duration_minutes"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// PineappleClassSchedule represents when a class occurs
type PineappleClassSchedule struct {
	ID          int       `db:"id" json:"id"`
	ClassID     int       `db:"class_id" json:"class_id"`
	DayOfWeek   int       `db:"day_of_week" json:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime   string    `db:"start_time" json:"start_time"`   // HH:MM:SS format
	EndTime     string    `db:"end_time" json:"end_time"`       // HH:MM:SS format
	Room        string    `db:"room" json:"room"`
	MaxCapacity int       `db:"max_capacity" json:"max_capacity"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// PineappleClassWithSchedule combines class and schedule information
type PineappleClassWithSchedule struct {
	PineappleClass
	Schedules []PineappleClassSchedule `json:"schedules"`
}

// PineappleSubscription represents a user's subscription to a class
type PineappleSubscription struct {
	ID                    int       `db:"id" json:"id"`
	ScheduleID            int       `db:"schedule_id" json:"schedule_id"`
	PushEndpoint          string    `db:"push_endpoint" json:"push_endpoint"`
	PushP256dh            string    `db:"push_p256dh" json:"push_p256dh"`
	PushAuth              string    `db:"push_auth" json:"push_auth"`
	NotificationThreshold int       `db:"notification_threshold" json:"notification_threshold"`
	Notify2hBefore        bool      `db:"notify_2h_before" json:"notify_2h_before"`
	Notify1hBefore        bool      `db:"notify_1h_before" json:"notify_1h_before"`
	Notify30mBefore       bool      `db:"notify_30m_before" json:"notify_30m_before"`
	IsActive              bool      `db:"is_active" json:"is_active"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

// PineappleAvailabilitySnapshot represents seat availability at a point in time
type PineappleAvailabilitySnapshot struct {
	ID               int64     `db:"id" json:"id"`
	ScheduleID       int       `db:"schedule_id" json:"schedule_id"`
	SnapshotDate     time.Time `db:"snapshot_date" json:"snapshot_date"`
	SnapshotTime     time.Time `db:"snapshot_time" json:"snapshot_time"`
	SeatsAvailable   int       `db:"seats_available" json:"seats_available"`
	SeatsTotal       int       `db:"seats_total" json:"seats_total"`
	IsFullyBooked    bool      `db:"is_fully_booked" json:"is_fully_booked"`
	ScrapeSuccessful bool      `db:"scrape_successful" json:"scrape_successful"`
	ErrorMessage     *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// PineappleNotification represents a sent notification
type PineappleNotification struct {
	ID                 int64     `db:"id" json:"id"`
	SubscriptionID     int       `db:"subscription_id" json:"subscription_id"`
	ScheduleID         int       `db:"schedule_id" json:"schedule_id"`
	ClassDate          time.Time `db:"class_date" json:"class_date"`
	NotificationType   string    `db:"notification_type" json:"notification_type"`
	SeatsAvailable     int       `db:"seats_available" json:"seats_available"`
	SentAt             time.Time `db:"sent_at" json:"sent_at"`
	DeliverySuccessful bool      `db:"delivery_successful" json:"delivery_successful"`
	ErrorMessage       *string   `db:"error_message" json:"error_message,omitempty"`
}

// PineappleScrapeJob represents a scraping job
type PineappleScrapeJob struct {
	ID                  int        `db:"id" json:"id"`
	JobType             string     `db:"job_type" json:"job_type"`
	Status              string     `db:"status" json:"status"`
	ClassesScraped      int        `db:"classes_scraped" json:"classes_scraped"`
	SchedulesUpdated    int        `db:"schedules_updated" json:"schedules_updated"`
	AvailabilityChecks  int        `db:"availability_checks" json:"availability_checks"`
	ErrorMessage        *string    `db:"error_message" json:"error_message,omitempty"`
	StartedAt           *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt         *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}

// ScrapedClassData represents raw class data from scraping
type ScrapedClassData struct {
	ClassName       string
	Instructor      string
	Level           string
	Style           string
	Description     string
	DurationMinutes int
	DayOfWeek       int    // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime       string // HH:MM format
	EndTime         string // HH:MM format
	Room            string
	SeatsAvailable  int
	SeatsTotal      int
}
