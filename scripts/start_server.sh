#!/bin/bash

# remiaq Server Startup Script

echo "🚀 Starting remiaq API Server..."

# Check if pb_data directory exists
if [ ! -d "pb_data" ]; then
    echo "📁 Creating pb_data directory..."
    mkdir -p pb_data
fi

# Check if database exists
if [ ! -f "pb_data/data.db" ]; then
    echo "🗄️  Database not found. Creating database..."
    sqlite3 pb_data/data.db < migrations/001_initial_schema.sql
    echo "✅ Database created successfully!"
fi

# Load environment variables if .env exists
if [ -f ".env" ]; then
    echo "📝 Loading environment variables from .env..."
    export $(cat .env | grep -v '^#' | xargs)
fi

# Start server
echo "🌐 Starting server on ${SERVER_ADDR:-127.0.0.1:8888}..."
go run cmd/server/main.go