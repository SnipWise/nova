#!/bin/bash

# NOVA Redis Vector Store - Quick Start Script
# This script starts Redis Stack for use with NOVA RAG Agent

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.redis.yml"

echo "🚀 Starting NOVA Redis Vector Store..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "   Please start Docker Desktop and try again"
    exit 1
fi

# Start Redis
docker-compose -f "$COMPOSE_FILE" up -d

echo ""
echo "⏳ Waiting for Redis to be ready..."
sleep 3

# Wait for Redis to be healthy
RETRY=0
MAX_RETRY=10
until docker exec nova-redis-vector-store redis-cli ping > /dev/null 2>&1; do
    RETRY=$((RETRY+1))
    if [ $RETRY -eq $MAX_RETRY ]; then
        echo "❌ Error: Redis failed to start"
        echo "   Check logs with: docker-compose -f docker-compose.redis.yml logs"
        exit 1
    fi
    echo "   Waiting... ($RETRY/$MAX_RETRY)"
    sleep 2
done

echo ""
echo "✅ Redis Vector Store is ready!"
echo ""
echo "📊 Connection Info:"
echo "   Redis Server: localhost:6379"
echo "   Redis CLI:    docker exec -it nova-redis-vector-store redis-cli"
echo ""
echo "🔍 Quick Commands:"
echo "   View logs:    docker-compose -f docker-compose.redis.yml logs -f"
echo "   Stop Redis:   docker-compose -f docker-compose.redis.yml stop"
echo "   Remove all:   docker-compose -f docker-compose.redis.yml down -v"
echo ""
echo "📚 See REDIS-SETUP.md for detailed documentation"
echo ""
