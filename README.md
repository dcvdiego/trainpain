# TrainPain

> **Is your commute reliable? Find out before you move.**

A data-driven web application that analyzes historical UK train reliability to help people make informed decisions about where to live.

## What is TrainPain?

TrainPain uses real historical performance data from National Rail and TfL to answer the question: **"If I live here and commute there, how reliable will my journey actually be?"**

Unlike typical rail apps that show schedules and real-time updates, TrainPain focuses on long-term reliability patterns - cancellation rates, typical delays, worst-case scenarios, and day-to-day consistency.

## Key Features

- **Route Reliability Analysis:** Enter any UK train route and see how reliable it actually is based on historical data
- **Comprehensive Metrics:** On-time performance, cancellation rates, delay distributions, and more
- **Commute Comparison:** Compare multiple potential living locations to find the most reliable commute
- **Data-Driven Decisions:** Help house hunters and renters understand commute quality before making a commitment

## Tech Stack

**Frontend:**
- React + TypeScript + Vite
- Modern UI with interactive data visualizations

**Backend:**
- Go 1.22+
- PostgreSQL database
- Integration with National Rail HSP API and TfL Unified API

**Data Sources:**
- National Rail Historical Service Performance (HSP) API - up to 1 year of historical data
- TfL (Transport for London) Unified API - London Underground, Overground, Elizabeth Line, DLR

## Project Status

**Current Phase:** Planning & Design Complete

This repository contains a comprehensive project plan ready for implementation. See [PROJECT_PLAN.md](./PROJECT_PLAN.md) for the complete technical specification.

## Quick Start (Future)

Once implemented, the development setup will be:

```bash
# Clone repository
git clone https://github.com/yourusername/trainpain.git
cd trainpain

# Backend setup
cd backend
docker-compose up -d
go run cmd/api/main.go

# Frontend setup
cd ../frontend
npm install
npm run dev
```

## Documentation

- **[PROJECT_PLAN.md](./PROJECT_PLAN.md)** - Complete technical specification including:
  - System architecture
  - Database schema design
  - Feature specifications and user flows
  - API integration details
  - Implementation phases and timeline
  - Critical considerations and risk mitigation

## Data Attribution

This application uses:
- National Rail data under Open Government License 2.0
- TfL data under TfL Open Data license

Required attribution:
- "Contains National Rail data © 2025"
- "Powered by TfL Open Data"

## Next Steps

See the [Next Steps section](./PROJECT_PLAN.md#next-steps) in the project plan for detailed implementation guidance.

**Week 1 priorities:**
1. Register for National Rail Data Portal (HSP API access)
2. Register for TfL API credentials
3. Set up project structure (frontend + backend)
4. Create proof of concept with sample route query

## Contributing

This project is currently in the planning phase. Contributions welcome once implementation begins.

## License

[To be determined]

## Contact

[Your contact information]

---

**Estimated MVP Timeline:** 6-7 weeks (part-time) or 3-4 weeks (full-time)

**Estimated Monthly Cost (Production):** $32/month (low traffic) to $140/month (scaling)
