#!/bin/bash

echo "🚂 TrainPain Development Setup"
echo "==============================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop."
    exit 1
fi

echo "✅ Docker is running"
echo ""

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Check if database is accessible
until docker-compose exec -T postgres pg_isready -U trainpain_user -d trainpain > /dev/null 2>&1; do
    echo "   Still waiting..."
    sleep 2
done

echo "✅ PostgreSQL is ready"
echo ""

# Run migrations
echo "📊 Running database migrations..."
cd backend
make migrate-up

if [ $? -eq 0 ]; then
    echo "✅ Migrations completed successfully"
else
    echo "⚠️  Migration failed or already up to date"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To start the backend:"
echo "  cd backend && make run"
echo ""
echo "To start the frontend (in another terminal):"
echo "  cd frontend && npm run dev"
echo ""
