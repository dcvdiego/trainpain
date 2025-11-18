# TrainPain - Setup Guide

Complete setup guide for getting TrainPain running locally and deploying to production.

## Prerequisites

- **Go 1.22+** - [Download](https://golang.org/dl/)
- **Node.js 20+** - [Download](https://nodejs.org/)
- **Docker** & **Docker Compose** - [Download](https://www.docker.com/products/docker-desktop)
- **PostgreSQL 15+** (or use Docker)
- **National Rail API Credentials** - Register at https://opendata.nationalrail.co.uk/
- (Optional) **TfL API Credentials** - Register at https://api-portal.tfl.gov.uk/

## Part 1: Get API Credentials

### National Rail HSP API (Required)

1. Go to https://opendata.nationalrail.co.uk/
2. Create an account
3. Subscribe to the **HSP** (Historical Service Performance) feed
4. Note your username and password

### TfL API (Optional - for London routes)

1. Go to https://api-portal.tfl.gov.uk/
2. Register for an account
3. Create an API key
4. Note your `app_id` and `app_key`

## Part 2: Local Development Setup

### 1. Clone and Navigate

```bash
git clone https://github.com/dcvdiego/trainpain.git
cd trainpain
```

### 2. Backend Setup

```bash
cd backend

# Copy environment config
cp .env.example .env

# Edit .env and add your credentials:
nano .env  # or use your preferred editor
```

Edit the following values in `.env`:
```
HSP_USERNAME=your_username_here
HSP_PASSWORD=your_password_here
TFL_APP_ID=your_app_id_here  # optional
TFL_APP_KEY=your_app_key_here  # optional
```

### 3. Start Database

```bash
# From project root
docker-compose up -d

# Wait for PostgreSQL to be ready (about 10 seconds)
```

### 4. Run Database Migrations

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Windows
choco install migrate
```

Run migrations:
```bash
cd backend
make migrate-up

# Or manually:
migrate -path migrations -database "postgresql://trainpain_user:trainpain_dev_password@localhost:5432/trainpain?sslmode=disable" up
```

### 5. Start Backend Server

```bash
cd backend
go run cmd/api/main.go

# Or use make:
make run
```

Backend should now be running on http://localhost:8080

Test it:
```bash
curl http://localhost:8080/health
# Should return: {"status":"ok","service":"trainpain-api"}
```

### 6. Start Frontend

Open a new terminal:

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

Frontend should now be running on http://localhost:5173

## Part 3: Testing the Application

1. Open http://localhost:5173 in your browser
2. Try a sample route:
   - **From:** Kingston
   - **To:** Waterloo
   - **Day Filter:** Weekdays
   - **Analysis Period:** Last 90 days
3. Click "Check Reliability"
4. You should see reliability metrics after 5-10 seconds

### Sample Routes to Test

- **Kingston → Waterloo** (South West London commute)
- **Brighton → London Victoria** (South Coast commute)
- **Reading → London Paddington** (Thames Valley commute)
- **Cambridge → London Kings Cross** (East Anglia commute)

## Part 4: Deployment

### Option 1: Free Tier Deployment (Recommended for 2 users)

#### Database: Neon (Free PostgreSQL)

1. Go to https://neon.tech/
2. Create a free account
3. Create a new project
4. Copy the connection string
5. Run migrations:
   ```bash
   migrate -path backend/migrations -database "YOUR_NEON_CONNECTION_STRING" up
   ```

#### Backend: Railway (Free Tier)

1. Go to https://railway.app/
2. Create account and new project
3. Click "Deploy from GitHub repo"
4. Select your `trainpain` repository
5. Set root directory to `/backend`
6. Add environment variables:
   ```
   DATABASE_HOST=your-neon-host
   DATABASE_PORT=5432
   DATABASE_USER=your-neon-user
   DATABASE_PASSWORD=your-neon-password
   DATABASE_DBNAME=your-neon-dbname
   DATABASE_SSLMODE=require
   HSP_USERNAME=your_hsp_username
   HSP_PASSWORD=your_hsp_password
   SERVER_PORT=8080
   SERVER_ENV=production
   ```
7. Deploy!
8. Note your Railway backend URL (e.g., `https://trainpain-production.up.railway.app`)

#### Frontend: Cloudflare Pages (Free)

1. Go to https://dash.cloudflare.com/
2. Go to Pages → Create a project
3. Connect to your GitHub repository
4. Configure build settings:
   - **Build command:** `npm run build`
   - **Build output directory:** `dist`
   - **Root directory:** `frontend`
5. Add environment variable:
   ```
   VITE_API_URL=https://your-railway-backend-url.railway.app/api/v1
   ```
6. Deploy!
7. Your site will be live at `https://trainpain.pages.dev`

### Option 2: Self-Hosted (VPS)

If you have a VPS (DigitalOcean, Hetzner, etc.):

1. Install Docker and Docker Compose
2. Clone repository
3. Create `.env` files with production credentials
4. Run:
   ```bash
   docker-compose up -d
   cd backend && make migrate-up
   ```
5. Set up Nginx reverse proxy
6. Configure SSL with Let's Encrypt

## Troubleshooting

### Backend won't start

- Check database is running: `docker ps`
- Check environment variables are set
- Check logs: `docker-compose logs postgres`

### Frontend can't connect to backend

- Verify `VITE_API_URL` in `frontend/.env.local`
- Check CORS settings in `backend/internal/api/middleware/cors.go`
- Check backend is running: `curl http://localhost:8080/health`

### "Station not found" errors

- Database needs to be seeded with stations
- Run migrations: `make migrate-up`
- Check stations exist: `psql -d trainpain -c "SELECT count(*) FROM stations;"`

### HSP API errors

- Verify credentials in `.env`
- Check HSP API is accessible: `curl https://hsp-prod.rockshore.net`
- API may have rate limits - wait a few minutes and try again

### No data returned

- HSP API might not have data for that specific route
- Try a major route like Kingston → Waterloo
- Check date range isn't too far in the past

## Development Commands

### Backend

```bash
cd backend

make run          # Run API server
make build        # Build binary
make test         # Run tests
make docker-up    # Start database
make docker-down  # Stop database
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
```

### Frontend

```bash
cd frontend

npm run dev    # Start dev server
npm run build  # Build for production
npm run preview # Preview production build
npm run lint   # Run ESLint
```

## Cost Estimates

### Free Tier (Perfect for 2 users)

- **Frontend:** Cloudflare Pages - $0
- **Backend:** Railway - $0 (500 hours/month free)
- **Database:** Neon - $0 (0.5GB free)
- **Total:** $0/month

### Low Traffic Production

- **Frontend:** Cloudflare Pages - $0
- **Backend:** Railway Hobby - $5/month
- **Database:** Neon - $0
- **Total:** ~$5/month

## Next Steps

- Add user authentication (optional)
- Implement background data collection for popular routes
- Add caching layer (Redis) for better performance
- Create comparison feature for multiple routes
- Add email alerts for route reliability changes

## Support

For issues or questions:
- Check the [main documentation](./PROJECT_PLAN.md)
- Open an issue on GitHub
- Check logs: `docker-compose logs`

## License

See LICENSE file for details.
