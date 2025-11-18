-- ============================================
-- STATION & ROUTE METADATA
-- ============================================

CREATE TABLE IF NOT EXISTS stations (
    id SERIAL PRIMARY KEY,
    crs_code VARCHAR(3) UNIQUE,
    tiploc VARCHAR(7),
    station_name VARCHAR(255) NOT NULL,
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    station_type VARCHAR(50),
    tfl_station_id VARCHAR(50),
    zone VARCHAR(10),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stations_crs ON stations(crs_code);
CREATE INDEX IF NOT EXISTS idx_stations_name ON stations(station_name);
CREATE INDEX IF NOT EXISTS idx_stations_type ON stations(station_type);

CREATE TABLE IF NOT EXISTS routes (
    id SERIAL PRIMARY KEY,
    origin_station_id INT REFERENCES stations(id),
    destination_station_id INT REFERENCES stations(id),
    route_hash VARCHAR(64) UNIQUE,
    operators TEXT[],
    typical_duration_minutes INT,
    distance_miles DECIMAL(6,2),
    is_tfl_route BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(origin_station_id, destination_station_id)
);

CREATE INDEX IF NOT EXISTS idx_routes_origin ON routes(origin_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_destination ON routes(destination_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_hash ON routes(route_hash);

-- ============================================
-- HISTORICAL SERVICE PERFORMANCE DATA
-- ============================================

CREATE TABLE IF NOT EXISTS service_records (
    id BIGSERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id),
    rid VARCHAR(50),
    service_date DATE NOT NULL,
    day_of_week INT,
    is_weekday BOOLEAN,

    scheduled_departure TIME NOT NULL,
    scheduled_arrival TIME NOT NULL,
    scheduled_duration_minutes INT,

    actual_departure TIME,
    actual_arrival TIME,
    actual_duration_minutes INT,

    departure_delay_minutes INT DEFAULT 0,
    arrival_delay_minutes INT DEFAULT 0,
    was_cancelled BOOLEAN DEFAULT FALSE,
    cancellation_reason TEXT,
    delay_reason TEXT,

    is_on_time BOOLEAN,
    is_slightly_delayed BOOLEAN,
    is_significantly_delayed BOOLEAN,
    is_severely_delayed BOOLEAN,

    operator_code VARCHAR(10),
    data_source VARCHAR(20),

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_records_route_date ON service_records(route_id, service_date);
CREATE INDEX IF NOT EXISTS idx_service_records_date ON service_records(service_date);
CREATE INDEX IF NOT EXISTS idx_service_records_route_dow ON service_records(route_id, day_of_week);
CREATE INDEX IF NOT EXISTS idx_service_records_scheduled_departure ON service_records(scheduled_departure);
CREATE INDEX IF NOT EXISTS idx_service_records_cancelled ON service_records(was_cancelled);
CREATE INDEX IF NOT EXISTS idx_service_records_delayed ON service_records(is_severely_delayed);

-- ============================================
-- PRE-COMPUTED RELIABILITY METRICS
-- ============================================

CREATE TABLE IF NOT EXISTS route_reliability_metrics (
    id SERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id),

    time_period VARCHAR(50),
    start_time TIME,
    end_time TIME,
    day_filter VARCHAR(20),

    analysis_start_date DATE,
    analysis_end_date DATE,
    total_services_analyzed INT,

    cancellation_rate DECIMAL(5,2),
    on_time_rate DECIMAL(5,2),
    avg_delay_minutes DECIMAL(6,2),
    median_delay_minutes INT,
    p95_delay_minutes INT,
    p99_delay_minutes INT,

    pct_0_5_min_late DECIMAL(5,2),
    pct_5_15_min_late DECIMAL(5,2),
    pct_15_30_min_late DECIMAL(5,2),
    pct_30_plus_min_late DECIMAL(5,2),

    reliability_score DECIMAL(5,2),

    best_day_of_week VARCHAR(10),
    worst_day_of_week VARCHAR(10),
    most_common_delay_reason TEXT,

    computed_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(route_id, time_period, day_filter, start_time, end_time)
);

CREATE INDEX IF NOT EXISTS idx_route_reliability_metrics_route_period ON route_reliability_metrics(route_id, time_period);
CREATE INDEX IF NOT EXISTS idx_route_reliability_metrics_route_day ON route_reliability_metrics(route_id, day_filter);
CREATE INDEX IF NOT EXISTS idx_route_reliability_metrics_reliability_score ON route_reliability_metrics(reliability_score);

-- ============================================
-- API RESPONSE CACHING
-- ============================================

CREATE TABLE IF NOT EXISTS api_cache (
    id SERIAL PRIMARY KEY,
    cache_key VARCHAR(255) UNIQUE NOT NULL,
    api_source VARCHAR(50),
    request_params JSONB,
    response_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    hit_count INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_api_cache_cache_key ON api_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_api_cache_expires ON api_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_api_cache_source ON api_cache(api_source);

-- ============================================
-- USER QUERIES (for analytics)
-- ============================================

CREATE TABLE IF NOT EXISTS user_queries (
    id BIGSERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id),
    query_params JSONB,
    result_cached BOOLEAN,
    response_time_ms INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_queries_route ON user_queries(route_id);
CREATE INDEX IF NOT EXISTS idx_user_queries_created ON user_queries(created_at);

-- ============================================
-- BACKGROUND JOB TRACKING
-- ============================================

CREATE TABLE IF NOT EXISTS data_collection_jobs (
    id SERIAL PRIMARY KEY,
    job_type VARCHAR(50),
    status VARCHAR(20),
    route_id INT REFERENCES routes(id),
    date_range_start DATE,
    date_range_end DATE,
    records_processed INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_collection_jobs_status ON data_collection_jobs(status);
CREATE INDEX IF NOT EXISTS idx_data_collection_jobs_job_type ON data_collection_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_data_collection_jobs_created ON data_collection_jobs(created_at);
