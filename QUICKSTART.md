# 🚀 Quick Start Guide

Get TrainPain running in 5 minutes!

## Prerequisites

- **Docker Desktop** must be running
- **Go 1.22+** and **Node.js 20+** installed
- **National Rail API credentials** ([Register here](https://opendata.nationalrail.co.uk/))

## Option 1: Automated Setup (Recommended)

```bash
# From project root
./dev-setup.sh

# Then start backend
cd backend
make run

# In another terminal, start frontend
cd frontend
npm run dev

# Open http://localhost:5173
```

## Option 2: Manual Setup

### 1. Start Docker Desktop

Make sure Docker Desktop is running on your Mac.

### 2. Start Database

```bash
# From project root
docker-compose up -d

# Wait 10 seconds for PostgreSQL to start
sleep 10
```

### 3. Run Migrations

```bash
cd backend

# Option A: Using golang-migrate (if installed)
make migrate-up

# Option B: Manual install of migrate
brew install golang-migrate
make migrate-up
```

**Don't have golang-migrate?** Install it:
```bash
brew install golang-migrate
```

### 4. Configure Backend

The backend will use these defaults (no .env file needed):
- Database: `localhost:5432`
- User: `trainpain_user`
- Password: `trainpain_dev_password`
- Database: `trainpain`

**Optional:** Add your National Rail credentials
```bash
cd backend
export HSP_USERNAME=your_username
export HSP_PASSWORD=your_password
```

Or create a `.env` file:
```bash
cp .env.example .env
# Edit .env and add your credentials
```

### 5. Start Backend

```bash
cd backend
make run

# Or directly:
go run cmd/api/main.go
```

You should see:
```
INFO    starting trainpain API server   {"env": "development", "port": "8080"}
INFO    database connection established
INFO    server listening    {"address": "0.0.0.0:8080"}
```

### 6. Start Frontend

Open a **new terminal**:

```bash
cd frontend
npm install  # First time only
npm run dev
```

You should see:
```
  VITE v5.x.x  ready in X ms

  ➜  Local:   http://localhost:5173/
```

### 7. Open Browser

Navigate to **http://localhost:5173**

## Testing It Works

Try this route:
- **From:** Kingston
- **To:** Waterloo
- **Day Filter:** Weekdays
- **Analysis Period:** Last 90 days

Click "Check Reliability" - you should see metrics after 5-10 seconds.

## Troubleshooting

### "failed to connect to database"

**Problem:** Database not running

**Solution:**
```bash
# Check if Docker is running
docker ps

# If no containers, start them:
docker-compose up -d

# Wait and retry
sleep 10
cd backend && make run
```

### "Station not found"

**Problem:** Database not migrated

**Solution:**
```bash
cd backend
make migrate-up

# Or manually:
brew install golang-migrate
migrate -path migrations -database "postgresql://trainpain_user:trainpain_dev_password@localhost:5432/trainpain?sslmode=disable" up
```

### "HSP API error"

**Problem:** No National Rail credentials

**Solution:** You can still test the UI, but won't get real data. To fix:
1. Register at https://opendata.nationalrail.co.uk/
2. Subscribe to HSP feed
3. Add credentials:
   ```bash
   export HSP_USERNAME=your_username
   export HSP_PASSWORD=your_password
   ```

### Port 5173 or 8080 already in use

**Solution:**
```bash
# Find and kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Find and kill process on port 5173
lsof -ti:5173 | xargs kill -9
```

### Frontend can't reach backend

**Problem:** CORS or backend not running

**Solution:**
1. Make sure backend is running on port 8080
2. Check `frontend/.env.local` has:
   ```
   VITE_API_URL=http://localhost:8080/api/v1
   ```
3. Restart frontend: `npm run dev`

## Sample Routes to Test

Once running, try these UK routes:

**London Commutes:**
- Kingston → Waterloo
- Wimbledon → Waterloo
- Richmond → Waterloo

**Intercity:**
- Brighton → Victoria
- Reading → Paddington
- Cambridge → Kings Cross

## Development Commands

```bash
# Backend
cd backend
make run          # Start server
make migrate-up   # Run migrations
make docker-up    # Start PostgreSQL
make docker-down  # Stop PostgreSQL

# Frontend
cd frontend
npm run dev       # Start dev server
npm run build     # Build for production
npm run preview   # Preview production build
```

## Next Steps

- **Get API Credentials:** Register at https://opendata.nationalrail.co.uk/
- **Add More Stations:** Edit `backend/migrations/002_seed_stations.up.sql`
- **Deploy:** See [SETUP.md](./SETUP.md) for deployment to Cloudflare + Railway

## Need Help?

- **Full setup guide:** [SETUP.md](./SETUP.md)
- **Technical docs:** [PROJECT_PLAN.md](./PROJECT_PLAN.md)
- **Issues:** Check the troubleshooting section above

---

**Happy coding! 🚂**
