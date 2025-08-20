#!/bin/bash

# Quick Start Script for URL Shortener Service
# This script helps you get the service running locally

echo "🚀 URL Shortener Service - Quick Start"
echo "======================================"

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not in PATH. Please add Go to your PATH:"
    echo "   export PATH=\$PATH:/usr/local/go/bin"
    echo "   Then run this script again."
    exit 1
fi

echo "✅ Go found: $(go version)"

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 &> /dev/null; then
    echo "⚠️  PostgreSQL is not running on localhost:5432"
    echo "   You can start it with:"
    echo "   brew services start postgresql"
    echo "   Or use Docker: docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15"
fi

# Check if Redis is running
if ! redis-cli ping &> /dev/null; then
    echo "⚠️  Redis is not running on localhost:6379"
    echo "   You can start it with:"
    echo "   brew services start redis"
    echo "   Or use Docker: docker run -d --name redis -p 6379:6379 redis:7"
fi

echo ""
echo "🔧 Setting up the service..."

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

# Build the service
echo "🔨 Building the service..."
go build -o bin/urlshortener ./cmd/api

if [ $? -eq 0 ]; then
    echo "✅ Service built successfully!"
    echo ""
    echo "🚀 To run the service:"
    echo "   ./bin/urlshortener"
    echo ""
    echo "📋 Environment variables you can set:"
    echo "   export URLSHORTENER_DATABASE_HOST=localhost"
    echo "   export URLSHORTENER_DATABASE_PORT=5432"
    echo "   export URLSHORTENER_DATABASE_USER=urlshortener"
    echo "   export URLSHORTENER_DATABASE_PASSWORD=password"
    echo "   export URLSHORTENER_DATABASE_DBNAME=urlshortener"
    echo "   export URLSHORTENER_REDIS_HOST=localhost"
    echo "   export URLSHORTENER_REDIS_PORT=6379"
    echo ""
    echo "🌐 The service will be available at: http://localhost:8080"
    echo "📊 Health check: http://localhost:8080/api/v1/healthz"
    echo "📈 Metrics: http://localhost:8080/metrics"
else
    echo "❌ Build failed. Please check the error messages above."
    exit 1
fi
