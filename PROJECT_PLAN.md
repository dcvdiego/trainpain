# TrainPain - Project Plan

> **Is your commute reliable? Find out before you move.**

A comprehensive web application to help UK residents make informed decisions about where to live by analyzing historical train reliability data using National Rail and TfL APIs.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [System Architecture](#system-architecture)
3. [Database Schema](#database-schema)
4. [Features & User Flows](#features--user-flows)
5. [Technical Stack](#technical-stack)
6. [Implementation Phases](#implementation-phases)
7. [API Integration](#api-integration)
8. [Critical Considerations](#critical-considerations)
9. [Next Steps](#next-steps)

---

## Project Overview

### Mission
Provide transparent, data-driven insights into UK train commute reliability using real historical performance data to help people make better housing and lifestyle decisions.

### Target Users
- **House Hunters:** People considering moving to new areas and want to understand commute quality
- **Current Commuters:** People checking if their current route is typical or problematic
- **Data Enthusiasts:** People interested in UK rail performance analytics

### Core Value Proposition
Unlike general rail apps that show schedules and real-time status, TrainPain focuses on **historical reliability** to answer: "If I live here and work there, what will my commute actually be like over time?"

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         FRONTEND                             │
│          (React + TypeScript + Vite)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Route Input  │  │  Metrics     │  │   Map View   │     │
│  │   Component  │  │  Dashboard   │  │   (optional) │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            ↕ REST API
┌─────────────────────────────────────────────────────────────┐
│                      BACKEND (Go)                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              API Gateway Layer                        │  │
│  │  (Rate limiting, Auth, Request validation)           │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Query       │  │  Analytics   │  │  Data        │     │
│  │  Service     │  │  Engine      │  │  Collector   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         External API Integration Layer               │  │
│  │  ┌───────────────┐      ┌───────────────┐          │  │
│  │  │ National Rail │      │   TfL API     │          │  │
│  │  │  HSP Client   │      │    Client     │          │  │
│  │  └───────────────┘      └───────────────┘          │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                   DATABASE (PostgreSQL)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Historical  │  │   API Cache  │  │  Stations    │     │
│  │  Performance │  │              │  │  Metadata    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│              BACKGROUND JOBS (Cron/Scheduler)                │
│  • Daily historical data collection                         │
│  • Metrics pre-computation                                  │
│  • Cache warming                                            │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow Strategy

**Two-Phase Approach:**

**Phase 1: On-Demand (MVP)**
- User queries → Backend queries HSP/TfL APIs in real-time
- Cache results in database for 24 hours
- Good for initial launch with low user volume

**Phase 2: Background Collection (Scale)**
- Scheduled jobs collect data for popular routes
- Pre-compute reliability metrics
- Serve from database (much faster, no API limits)

---

## Database Schema

### Core Tables

#### stations
```sql
CREATE TABLE stations (
    id SERIAL PRIMARY KEY,
    crs_code VARCHAR(3) UNIQUE,              -- National Rail CRS code (e.g., 'WAT')
    tiploc VARCHAR(7),                        -- TIPLOC code for National Rail
    station_name VARCHAR(255) NOT NULL,
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    station_type VARCHAR(50),                 -- 'national_rail', 'tube', 'overground', 'dlr'
    tfl_station_id VARCHAR(50),               -- TfL station ID if applicable
    zone VARCHAR(10),                          -- London zones (1-9)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### routes
```sql
CREATE TABLE routes (
    id SERIAL PRIMARY KEY,
    origin_station_id INT REFERENCES stations(id),
    destination_station_id INT REFERENCES stations(id),
    route_hash VARCHAR(64) UNIQUE,            -- MD5 hash for quick lookups
    operators TEXT[],                          -- Array of operator codes
    typical_duration_minutes INT,
    distance_miles DECIMAL(6,2),
    is_tfl_route BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(origin_station_id, destination_station_id)
);
```

#### service_records
```sql
CREATE TABLE service_records (
    id BIGSERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id),
    rid VARCHAR(50),                          -- Rail RID (unique train identifier)
    service_date DATE NOT NULL,
    day_of_week INT,                          -- 0=Sunday, 1=Monday, etc.
    is_weekday BOOLEAN,

    -- Scheduled times
    scheduled_departure TIME NOT NULL,
    scheduled_arrival TIME NOT NULL,
    scheduled_duration_minutes INT,

    -- Actual times
    actual_departure TIME,
    actual_arrival TIME,
    actual_duration_minutes INT,

    -- Performance metrics
    departure_delay_minutes INT DEFAULT 0,
    arrival_delay_minutes INT DEFAULT 0,
    was_cancelled BOOLEAN DEFAULT FALSE,
    cancellation_reason TEXT,
    delay_reason TEXT,

    -- Classification
    is_on_time BOOLEAN,                       -- Within 5 mins
    is_slightly_delayed BOOLEAN,              -- 5-15 mins
    is_significantly_delayed BOOLEAN,         -- 15-30 mins
    is_severely_delayed BOOLEAN,              -- 30+ mins

    -- Metadata
    operator_code VARCHAR(10),
    data_source VARCHAR(20),                  -- 'hsp_api', 'tfl_api'
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### route_reliability_metrics
```sql
CREATE TABLE route_reliability_metrics (
    id SERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id),

    -- Time window filters
    time_period VARCHAR(50),                  -- 'weekday_morning_peak', 'weekend_all_day'
    start_time TIME,
    end_time TIME,
    day_filter VARCHAR(20),                   -- 'weekday', 'weekend', 'friday', 'all'

    -- Analysis period
    analysis_start_date DATE,
    analysis_end_date DATE,
    total_services_analyzed INT,

    -- Core reliability metrics
    cancellation_rate DECIMAL(5,2),           -- Percentage (0-100)
    on_time_rate DECIMAL(5,2),                -- Within 5 mins
    avg_delay_minutes DECIMAL(6,2),
    median_delay_minutes INT,
    p95_delay_minutes INT,                    -- 95th percentile delay
    p99_delay_minutes INT,                    -- 99th percentile delay

    -- Delay distribution
    pct_0_5_min_late DECIMAL(5,2),
    pct_5_15_min_late DECIMAL(5,2),
    pct_15_30_min_late DECIMAL(5,2),
    pct_30_plus_min_late DECIMAL(5,2),

    -- Reliability score (0-100)
    reliability_score DECIMAL(5,2),

    computed_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(route_id, time_period, day_filter, start_time, end_time)
);
```

---

## Features & User Flows

### Core Features (MVP)

1. **Route Query Interface**
   - Origin/destination station search (autocomplete)
   - Time window selection (morning peak, evening peak, custom)
   - Day filter (weekdays, weekends, specific days)
   - Analysis period (30/60/90 days, 6 months, 1 year)

2. **Reliability Dashboard**
   - Reliability Score (0-100)
   - On-time performance percentage
   - Cancellation rate
   - Average delay, median delay, 95th percentile delay
   - Delay breakdown visualization (bar chart)
   - Most common delay reasons

3. **Key Metrics Display**
   ```
   🎯 RELIABILITY SCORE: 67/100 (Fair - Expect moderate delays)

   📊 KEY METRICS:
   • On-Time Performance: 58% (within 5 mins)
   • Cancellation Rate: 4.2%
   • Average Delay: 8.5 minutes
   • Worst Case (95th %ile): 25 minutes

   📈 DELAY BREAKDOWN:
   58% On-time (0-5 min)
   22% Slightly delayed (5-15 min)
   12% Significantly delayed (15-30 min)
   8% Severely delayed (30+ min)
   ```

### Enhanced Features (Phase 2)

4. **Route Comparison Tool**
   - Compare multiple routes side-by-side
   - "Should I Move Here?" calculator
   - Life impact metrics (days late per year, hours wasted)

5. **Visual Analytics**
   - Delay distribution histogram
   - Time-of-day reliability heatmap
   - Trend graphs (performance over time)

6. **Alternative Routes**
   - Suggest alternative departure stations
   - Compare different times of day

### User Flows

**Flow 1: First-Time Visitor**
```
Homepage → Enter origin/destination → Select time window →
View reliability dashboard → Explore insights → Save route (optional)
```

**Flow 2: House Hunter**
```
Compare Areas → Enter workplace → Add 3-5 potential home areas →
View comparison table → Sort by reliability → Deep dive on best option
```

**Flow 3: Regular Commuter**
```
Saved route (bookmark) → View updated metrics →
Check recent changes → Note trends
```

---

## Technical Stack

### Frontend
- **Framework:** React 18 + TypeScript
- **Build Tool:** Vite 5
- **UI Library:** Ant Design or Material-UI
- **Charts:** Recharts or D3.js
- **State Management:** Zustand or React Query
- **HTTP Client:** Axios
- **Routing:** React Router v6

### Backend
- **Language:** Go 1.22+
- **Web Framework:** Gin
- **Database Driver:** pgx/v5, sqlx
- **Migrations:** golang-migrate
- **Configuration:** Viper
- **HTTP Client:** Resty
- **Logging:** Zap
- **Scheduling:** robfig/cron

### Database
- **Primary:** PostgreSQL 15+
- **Optional Cache:** Redis 7

### DevOps
- **Containerization:** Docker + Docker Compose
- **CI/CD:** GitHub Actions
- **Hosting (MVP):** Railway/Render (backend), Vercel (frontend)
- **Hosting (Production):** DigitalOcean, AWS, or Hetzner

---

## Implementation Phases

### Phase 0: Foundation & Setup (Week 1-2)

**Goals:** API access, project initialization, development environment

**Tasks:**
- [ ] Register for National Rail Data Portal (HSP API access)
- [ ] Register for TfL API credentials
- [ ] Test API endpoints
- [ ] Initialize Git repo and project structure
- [ ] Set up Docker Compose (Postgres)
- [ ] Create database schema and migrations
- [ ] Seed stations table

**Deliverables:**
- Working API credentials
- Skeleton React + Go projects
- Development environment ready

---

### Phase 1: MVP Development (Week 3-6)

**Goals:** Launch basic working product with route reliability queries

**Frontend:**
- [ ] Route search form with autocomplete
- [ ] Time window and day filter selectors
- [ ] Reliability dashboard (metrics display)
- [ ] Delay breakdown visualization
- [ ] Responsive design

**Backend:**
- [ ] HSP API client with rate limiting
- [ ] Station search endpoint
- [ ] Route reliability endpoint
- [ ] Data transformation logic
- [ ] Reliability score calculation
- [ ] Database caching layer

**Data Pipeline:**
- [ ] On-demand API queries
- [ ] 24-hour cache TTL
- [ ] Basic analytics logging

**Testing:**
- [ ] Unit tests (Go)
- [ ] Integration tests (API endpoints)
- [ ] Manual testing (10+ routes)

**Success Criteria:**
- ✅ Users can query any UK route
- ✅ Accurate metrics (validated against National Rail reports)
- ✅ < 10s response time
- ✅ Deployed to public URL

---

### Phase 2: Enhanced Features (Week 7-10)

**Goals:** Add visualizations, TfL integration, comparison tools

**Frontend:**
- [ ] Interactive charts (delay distribution, heatmap, trends)
- [ ] Multi-route comparison page
- [ ] "Should I Move Here?" calculator
- [ ] Route favorites (localStorage)
- [ ] Export/share functionality

**Backend:**
- [ ] TfL API client
- [ ] Hybrid route queries (National Rail + TfL)
- [ ] Comparison endpoint
- [ ] Background worker for data collection
- [ ] Cron jobs (daily collection)
- [ ] Pre-computed metrics for top routes

**Deliverables:**
- Enhanced visualizations
- TfL support (London routes)
- Comparison tool
- Background data pipeline

---

### Phase 3: Production & Scaling (Week 11-12)

**Goals:** Optimize, monitor, scale

**Tasks:**
- [ ] Database query optimization
- [ ] Comprehensive caching strategy
- [ ] Monitoring (Sentry, Prometheus)
- [ ] Production deployment
- [ ] CI/CD pipeline
- [ ] SSL/HTTPS setup
- [ ] Automated backups

**Deliverables:**
- Production-ready app
- 99%+ uptime
- Monitoring dashboards

---

### Phase 4: Future Enhancements (Month 4+)

- [ ] User accounts
- [ ] Email alerts
- [ ] Real-time updates
- [ ] Mobile app
- [ ] Advanced analytics
- [ ] API for third parties

---

## API Integration

### National Rail HSP API

**Endpoints:**
- `POST https://hsp-prod.rockshore.net/api/v1/serviceMetrics`
- `POST https://hsp-prod.rockshore.net/api/v1/serviceDetails`

**Authentication:** Basic Auth (username/password)

**Rate Limit:** 5M requests per 4-week period (~178k/day)

**Data Available:**
- Historical performance up to 1 year back
- Actual vs scheduled times
- Cancellation and delay reasons
- By route, time window, day filter

**Parameters:**
```json
{
  "from_loc": "KNG",
  "to_loc": "WAT",
  "from_time": "07:00",
  "to_time": "10:00",
  "from_date": "2025-01-01",
  "to_date": "2025-03-01",
  "days": "WEEKDAY"
}
```

**Strategy:**
- Cache aggressively (24-48h TTL)
- Background collection for popular routes
- On-demand for long tail

---

### TfL Unified API

**Base URL:** `https://api.tfl.gov.uk`

**Authentication:** Query params `app_id` and `app_key`

**Rate Limit:** Generous (500+ requests/day, possibly higher)

**Key Endpoints:**
- `GET /Line/{ids}/Arrivals` - Real-time predictions
- `GET /Line/Mode/{modes}/Disruption` - Current disruptions
- `GET /Line/{id}/Status` - Line status

**Strategy:**
- Use for London-specific routes (Underground, Overground, Elizabeth, DLR)
- Complement National Rail data
- Real-time status (future feature)

---

### Data Coverage

| Service | API | Coverage |
|---------|-----|----------|
| National Rail | HSP | All UK intercity, regional, commuter trains |
| London Underground | TfL | All Tube lines |
| London Overground | TfL | Full coverage |
| Elizabeth Line | TfL | Full coverage |
| DLR | TfL | Full coverage |

**Overlap Handling:** Routes like Kingston → Waterloo involve National Rail but connect to TfL services. Handle both APIs for comprehensive coverage.

---

## Critical Considerations

### 1. Reliability Score Algorithm

**Proposed Formula:**
```
Reliability Score (0-100) =
  (On-time % × 0.4) +
  ((100 - Cancellation %) × 0.3) +
  ((100 - Avg Delay / 30) × 100 × 0.2) +
  ((100 - P95 Delay / 60) × 100 × 0.1)
```

**Weights:**
- 40% on-time performance (most important)
- 30% cancellation rate
- 20% average delay
- 10% worst-case delay (95th percentile)

**Interpretation:**
- 90-100: Excellent
- 75-89: Good
- 60-74: Fair
- 40-59: Poor
- 0-39: Very Poor

### 2. API Rate Limit Management

**HSP API:**
- Budget: 178k requests/day
- With smart caching, supports ~5k unique queries/day
- Strategy: 24-48h cache, prioritize popular routes

**TfL API:**
- Generally generous limits
- Strategy: Respect fair use, implement exponential backoff

### 3. Data Freshness vs Cost

**Approach:**
- **Top 100 routes:** Daily updates (background jobs)
- **Top 1000 routes:** Weekly updates
- **Long tail:** On-demand with 48h cache
- **User override:** Force refresh (rate-limited)

### 4. Legal & Compliance

**National Rail:**
- Open Government License 2.0
- Attribution: "Contains National Rail data © 2025"
- Commercial use allowed

**TfL:**
- Attribution: "Powered by TfL Open Data"
- Cannot imply endorsement

**User Disclaimer:**
- Data is historical and informational
- Real-time conditions may vary
- Verify with official sources

### 5. HSP API Migration (2026 Risk)

**Issue:** National Rail Data Portal retiring early 2026

**Mitigation:**
- Monitor Rail Data Marketplace announcements
- Plan migration path
- HSP API likely to continue via new portal
- Have backup plan (alternative data sources)

---

## Next Steps

### Week 1 Immediate Actions

**Day 1: API Registration**
1. Go to https://opendata.nationalrail.co.uk/
2. Create account and subscribe to HSP feed
3. Go to https://api-portal.tfl.gov.uk/
4. Register and get app_id/app_key
5. Test both APIs with Postman/cURL

**Day 2: Project Setup**
```bash
# Create project structure
mkdir trainpain && cd trainpain
mkdir frontend backend

# Frontend
cd frontend
npm create vite@latest . -- --template react-ts
npm install

# Backend
cd ../backend
go mod init github.com/yourusername/trainpain-backend
mkdir -p cmd/api internal/api/handlers internal/domain

# Database
docker-compose up -d postgres
```

**Day 3-5: Proof of Concept**
1. Write simple Go script to query HSP API
2. Fetch data for Kingston → Waterloo (last 30 days)
3. Parse response and calculate metrics
4. Validate data quality
5. **Goal:** Confirm API works and data is usable

**Day 6-7: Database**
1. Create migrations for core tables
2. Download National Rail station list
3. Seed stations table
4. Test queries

### Week 2: First Endpoint

**Goal:** End-to-end working prototype

1. Implement `/api/v1/routes/reliability` in Go
2. Create React form to call endpoint
3. Display results in basic UI
4. **Milestone:** User can query one route and see metrics

---

## Project Structure

### Frontend

```
frontend/
├── src/
│   ├── components/
│   │   ├── common/
│   │   ├── search/
│   │   ├── dashboard/
│   │   └── charts/
│   ├── pages/
│   ├── hooks/
│   ├── services/
│   ├── types/
│   ├── utils/
│   ├── App.tsx
│   └── main.tsx
├── package.json
└── vite.config.ts
```

### Backend

```
backend/
├── cmd/
│   ├── api/main.go
│   ├── worker/main.go
│   └── migrate/main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   └── middleware/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   ├── external/
│   │   ├── hsp_client.go
│   │   └── tfl_client.go
│   └── worker/
├── migrations/
├── go.mod
└── Dockerfile
```

---

## Resources

### API Documentation
- HSP API: https://hsp-prod.rockshore.net/
- Open Rail Data Wiki: https://wiki.openraildata.com/
- TfL API Portal: https://api-portal.tfl.gov.uk/
- TfL API Docs: https://api.tfl.gov.uk

### Community
- OpenRailData Forum: https://groups.google.com/g/openraildata-talk
- TfL Tech Forum: https://techforum.tfl.gov.uk/

### Data Sources
- National Rail Data Portal: https://opendata.nationalrail.co.uk/
- London Datastore: https://data.london.gov.uk/
- Rail Data Marketplace: (replacing NRDP in 2026)

---

## Timeline Summary

```
Week 1-2:   Foundation & API Setup
Week 3-6:   MVP Development
Week 7:     MVP Launch & Testing
Week 8-10:  Enhanced Features
Week 11-12: Production Deployment
Month 4+:   Ongoing Improvements
```

**Estimated Time to MVP:** 6-7 weeks (part-time) or 3-4 weeks (full-time)

---

## Cost Estimates

**Free Tier (MVP):**
- Frontend: Vercel - $0
- Backend: Railway - $0
- Database: Neon - $0
- **Total: $0/month**

**Production (Low Traffic):**
- Frontend: CloudFront + S3 - $5
- Backend: DigitalOcean Droplet - $12
- Database: Managed Postgres - $15
- **Total: ~$32/month**

**Production (Scaling):**
- All services scaled up
- **Total: ~$140/month**

---

## Success Metrics

**Product:**
- Routes queried per week
- Unique visitors per month
- User engagement (session duration)

**Technical:**
- API response time < 3s (cached), < 10s (uncached)
- Uptime > 99.5%
- Error rate < 1%

**Data Quality:**
- Coverage: % of UK routes with data
- Freshness: % of popular routes updated within 7 days
- Accuracy: Validated against official reports

---

## License & Attribution

This application uses:
- National Rail data under Open Government License 2.0
- TfL data under TfL Open Data license

Required attribution:
- "Contains National Rail data © 2025"
- "Powered by TfL Open Data"

---

**Created:** November 2025
**Version:** 1.0
**Status:** Planning Phase
