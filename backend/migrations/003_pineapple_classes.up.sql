-- ============================================
-- PINEAPPLE DANCE STUDIOS CLASSES
-- ============================================

-- Table to store dance classes
CREATE TABLE IF NOT EXISTS pineapple_classes (
    id SERIAL PRIMARY KEY,
    class_name VARCHAR(255) NOT NULL,
    instructor VARCHAR(255),
    level VARCHAR(50),
    style VARCHAR(100),
    description TEXT,
    duration_minutes INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pineapple_classes_style ON pineapple_classes(style);
CREATE INDEX IF NOT EXISTS idx_pineapple_classes_level ON pineapple_classes(level);

-- Table to store class schedules (when classes occur)
CREATE TABLE IF NOT EXISTS pineapple_class_schedules (
    id SERIAL PRIMARY KEY,
    class_id INT REFERENCES pineapple_classes(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL, -- 0=Sunday, 1=Monday, ..., 6=Saturday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room VARCHAR(100),
    max_capacity INT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(class_id, day_of_week, start_time)
);

CREATE INDEX IF NOT EXISTS idx_pineapple_class_schedules_class ON pineapple_class_schedules(class_id);
CREATE INDEX IF NOT EXISTS idx_pineapple_class_schedules_day ON pineapple_class_schedules(day_of_week);
CREATE INDEX IF NOT EXISTS idx_pineapple_class_schedules_active ON pineapple_class_schedules(is_active);

-- Table to store user subscriptions (using Web Push API subscription info)
CREATE TABLE IF NOT EXISTS pineapple_subscriptions (
    id SERIAL PRIMARY KEY,
    schedule_id INT REFERENCES pineapple_class_schedules(id) ON DELETE CASCADE,
    push_endpoint TEXT NOT NULL, -- Web Push subscription endpoint (acts as user identifier)
    push_p256dh TEXT NOT NULL, -- Public key for encryption
    push_auth TEXT NOT NULL, -- Auth secret for encryption
    notification_threshold INT DEFAULT 5, -- Notify when seats left <= this number
    notify_2h_before BOOLEAN DEFAULT TRUE,
    notify_1h_before BOOLEAN DEFAULT FALSE,
    notify_30m_before BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(schedule_id, push_endpoint)
);

CREATE INDEX IF NOT EXISTS idx_pineapple_subscriptions_schedule ON pineapple_subscriptions(schedule_id);
CREATE INDEX IF NOT EXISTS idx_pineapple_subscriptions_active ON pineapple_subscriptions(is_active);
CREATE INDEX IF NOT EXISTS idx_pineapple_subscriptions_endpoint ON pineapple_subscriptions(push_endpoint);

-- Table to track seat availability over time
CREATE TABLE IF NOT EXISTS pineapple_availability_snapshots (
    id BIGSERIAL PRIMARY KEY,
    schedule_id INT REFERENCES pineapple_class_schedules(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    snapshot_time TIMESTAMP NOT NULL,
    seats_available INT,
    seats_total INT,
    is_fully_booked BOOLEAN DEFAULT FALSE,
    scrape_successful BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pineapple_availability_schedule_date ON pineapple_availability_snapshots(schedule_id, snapshot_date);
CREATE INDEX IF NOT EXISTS idx_pineapple_availability_snapshot_time ON pineapple_availability_snapshots(snapshot_time);
CREATE INDEX IF NOT EXISTS idx_pineapple_availability_date ON pineapple_availability_snapshots(snapshot_date);

-- Table to track notification deliveries
CREATE TABLE IF NOT EXISTS pineapple_notifications (
    id BIGSERIAL PRIMARY KEY,
    subscription_id INT REFERENCES pineapple_subscriptions(id) ON DELETE CASCADE,
    schedule_id INT REFERENCES pineapple_class_schedules(id) ON DELETE CASCADE,
    class_date DATE NOT NULL,
    notification_type VARCHAR(50), -- '2h_before', '1h_before', '30m_before', 'threshold_reached'
    seats_available INT,
    sent_at TIMESTAMP DEFAULT NOW(),
    delivery_successful BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_pineapple_notifications_subscription ON pineapple_notifications(subscription_id);
CREATE INDEX IF NOT EXISTS idx_pineapple_notifications_schedule_date ON pineapple_notifications(schedule_id, class_date);
CREATE INDEX IF NOT EXISTS idx_pineapple_notifications_sent ON pineapple_notifications(sent_at);

-- Table to track scraping jobs
CREATE TABLE IF NOT EXISTS pineapple_scrape_jobs (
    id SERIAL PRIMARY KEY,
    job_type VARCHAR(50), -- 'full_scrape', 'availability_check'
    status VARCHAR(20), -- 'pending', 'running', 'completed', 'failed'
    classes_scraped INT DEFAULT 0,
    schedules_updated INT DEFAULT 0,
    availability_checks INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pineapple_scrape_jobs_status ON pineapple_scrape_jobs(status);
CREATE INDEX IF NOT EXISTS idx_pineapple_scrape_jobs_type ON pineapple_scrape_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_pineapple_scrape_jobs_created ON pineapple_scrape_jobs(created_at);
