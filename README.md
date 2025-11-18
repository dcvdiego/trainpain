# TrainPain

> **Is your commute reliable? Find out before you move.**

A data-driven web application that analyzes historical UK train reliability to help people make informed decisions about where to live.

## What is TrainPain?

TrainPain uses real historical performance data from National Rail and TfL to answer the question: **"If I live here and commute there, how reliable will my journey actually be?"**

Unlike typical rail apps that show schedules and real-time updates, TrainPain focuses on long-term reliability patterns - cancellation rates, typical delays, worst-case scenarios, and day-to-day consistency.

## ✨ Features

- **📊 Route Reliability Analysis:** Enter any UK train route and see how reliable it is based on historical data
- **🎯 Reliability Score:** 0-100 score with clear interpretation (Excellent/Good/Fair/Poor)
- **📈 Comprehensive Metrics:** On-time performance, cancellation rates, delay distributions
- **📱 Mobile-First Design:** Fully responsive, works great on phones and tablets
- **⚡ Fast & Free:** Optimized for deployment on free tier services

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 15+
- National Rail HSP API credentials ([Register here](https://opendata.nationalrail.co.uk/))

### Local Development

```bash
# 1. Clone repository
git clone https://github.com/dcvdiego/trainpain.git
cd trainpain

# 2. Start database
docker-compose up -d

# 3. Set up backend
cd backend
cp .env.example .env
# Edit .env and add your HSP API credentials
make migrate-up
make run

# 4. Set up frontend (in new terminal)
cd frontend
npm install
npm run dev

# 5. Open http://localhost:5173
```

**See [SETUP.md](./SETUP.md) for detailed setup instructions, deployment guides, and troubleshooting.**

## 📖 Documentation

- **[SETUP.md](./SETUP.md)** - Complete setup guide for local development and deployment
- **[PROJECT_PLAN.md](./PROJECT_PLAN.md)** - Technical specification and architecture

## 🏗️ Tech Stack

**Frontend:**
- React 18 + TypeScript + Vite
- Ant Design UI components
- Axios for API calls
- Mobile-first responsive design

**Backend:**
- Go 1.22+
- Gin web framework
- PostgreSQL database
- National Rail HSP API integration

**Deployment (Free Tier):**
- **Frontend:** Cloudflare Pages - $0
- **Backend:** Railway - $0 (500 hours/month)
- **Database:** Neon PostgreSQL - $0 (0.5GB)
- **Total:** $0/month

## 🎯 Example Routes

Try these popular commutes:

- **Kingston → Waterloo** (South West London)
- **Brighton → Victoria** (South Coast)
- **Reading → Paddington** (Thames Valley)
- **Cambridge → Kings Cross** (East Anglia)

## 📊 What You'll See

**Reliability Metrics Include:**
- 🎯 Overall reliability score (0-100)
- ✅ On-time performance percentage
- ❌ Cancellation rate
- ⏱️ Average delay time
- 📈 Delay distribution breakdown
- 🚨 Worst-case scenario (95th percentile)

## 🔐 API Credentials

**Required:**
- National Rail HSP API - [Register](https://opendata.nationalrail.co.uk/)

**Optional:**
- TfL API (for London routes) - [Register](https://api-portal.tfl.gov.uk/)

## 🚢 Deployment

### Cloudflare Pages (Frontend)

```bash
cd frontend
npm run build
# Deploy dist/ folder to Cloudflare Pages
```

### Railway (Backend)

1. Connect GitHub repository
2. Set root directory to `/backend`
3. Add environment variables (see SETUP.md)
4. Deploy!

**See [SETUP.md](./SETUP.md#part-4-deployment) for detailed deployment instructions.**

## 📱 Mobile-Friendly

TrainPain is built with mobile-first design using Ant Design's responsive grid system. Works perfectly on:
- 📱 Phones (all sizes)
- 📲 Tablets
- 💻 Desktops

## 🗺️ Roadmap

- [x] MVP with route reliability analysis
- [x] Mobile-first responsive design
- [x] Deployment configs for free hosting
- [ ] Route comparison feature
- [ ] Background data collection for popular routes
- [ ] TfL integration for London routes
- [ ] User accounts and saved routes
- [ ] Email alerts for reliability changes

## 📄 License

This project is for educational and informational purposes.

**Data Attribution:**
- Contains National Rail data © 2025 under Open Government License 2.0
- Powered by TfL Open Data (when applicable)

## 🐛 Troubleshooting

See [SETUP.md - Troubleshooting](./SETUP.md#troubleshooting) for common issues and solutions.

## 🤝 Contributing

Contributions welcome! Please open an issue first to discuss proposed changes.

## 💡 Support

- 📖 Check [SETUP.md](./SETUP.md) for detailed guides
- 🐛 [Open an issue](https://github.com/dcvdiego/trainpain/issues) for bugs
- 💬 Discussions welcome in Issues

---

**Built with ❤️ for UK commuters**
