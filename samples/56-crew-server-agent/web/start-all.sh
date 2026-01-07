#!/bin/bash

# Nova Crew Server - Complete Startup Script
# Starts backend, CORS proxy, and web interface in one command

set -e  # Exit on error

echo "🚀 Starting Nova Crew Server - Complete Stack"
echo "═══════════════════════════════════════════════"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "${YELLOW}Stopping all services...${NC}"

    # Kill all child processes
    pkill -P $$ || true

    echo "${GREEN}✓ All services stopped${NC}"
    exit 0
}

# Register cleanup on Ctrl+C
trap cleanup INT TERM

echo "📋 Pre-flight checks..."

# Check if main.go exists
if [ ! -f "../main.go" ]; then
    echo "${RED}✗ Error: main.go not found in parent directory${NC}"
    echo "  Please run this script from samples/56-crew-server-agent/web/"
    exit 1
fi

# Check if proxy exists
if [ ! -f "proxy/main.go" ]; then
    echo "${RED}✗ Error: proxy/main.go not found${NC}"
    echo "  Expected in: samples/56-crew-server-agent/web/proxy/main.go"
    exit 1
fi

# Check if ports are available
if check_port 8080; then
    echo "${RED}✗ Error: Port 8080 is already in use${NC}"
    echo "  Please stop the service using port 8080 or change the port"
    exit 1
fi

if check_port 8081; then
    echo "${RED}✗ Error: Port 8081 is already in use${NC}"
    echo "  Please stop the service using port 8081 or change the port"
    exit 1
fi

if check_port 3000; then
    echo "${YELLOW}⚠  Warning: Port 3000 is already in use${NC}"
    echo "  Will try alternative port 3001"
    WEB_PORT=3001
else
    WEB_PORT=3000
fi

echo "${GREEN}✓ All checks passed${NC}"
echo ""

# Start Backend (Go server)
echo "${BLUE}[1/3] Starting Backend Server (port 8080)...${NC}"
cd ..
go run main.go > /tmp/nova-backend.log 2>&1 &
BACKEND_PID=$!
cd web

# Wait for backend to be ready
echo "  Waiting for backend to start..."
for i in {1..30}; do
    if check_port 8080; then
        echo "${GREEN}  ✓ Backend started successfully${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "${RED}  ✗ Backend failed to start${NC}"
        echo "  Check logs: tail /tmp/nova-backend.log"
        cleanup
    fi
    sleep 1
done

# Start CORS Proxy
echo ""
echo "${BLUE}[2/3] Starting CORS Proxy (port 8081)...${NC}"
cd proxy
go run main.go > /tmp/nova-proxy.log 2>&1 &
PROXY_PID=$!
cd ..

# Wait for proxy to be ready
echo "  Waiting for proxy to start..."
for i in {1..10}; do
    if check_port 8081; then
        echo "${GREEN}  ✓ Proxy started successfully${NC}"
        break
    fi
    if [ $i -eq 10 ]; then
        echo "${RED}  ✗ Proxy failed to start${NC}"
        echo "  Check logs: tail /tmp/nova-proxy.log"
        cleanup
    fi
    sleep 1
done

# Start Web Interface
echo ""
echo "${BLUE}[3/3] Starting Web Interface (port $WEB_PORT)...${NC}"

# Detect available HTTP server
if command -v python3 &> /dev/null; then
    echo "  Using Python 3"
    python3 -m http.server $WEB_PORT > /tmp/nova-web.log 2>&1 &
    WEB_PID=$!
elif command -v python &> /dev/null; then
    echo "  Using Python 2"
    python -m SimpleHTTPServer $WEB_PORT > /tmp/nova-web.log 2>&1 &
    WEB_PID=$!
elif command -v php &> /dev/null; then
    echo "  Using PHP"
    php -S localhost:$WEB_PORT > /tmp/nova-web.log 2>&1 &
    WEB_PID=$!
else
    echo "${RED}  ✗ No HTTP server found${NC}"
    echo "  Please install Python, PHP, or Node.js"
    cleanup
fi

sleep 2
echo "${GREEN}  ✓ Web interface started successfully${NC}"

# Display summary
echo ""
echo "═══════════════════════════════════════════════"
echo "${GREEN}✓ All services are running!${NC}"
echo "═══════════════════════════════════════════════"
echo ""
echo "🌐 Open your browser to:"
echo "   ${BLUE}http://localhost:$WEB_PORT${NC}"
echo ""
echo "📡 Service endpoints:"
echo "   • Backend:       http://localhost:8080"
echo "   • CORS Proxy:    http://localhost:8081"
echo "   • Web Interface: http://localhost:$WEB_PORT"
echo ""
echo "📊 Process IDs:"
echo "   • Backend PID:  $BACKEND_PID"
echo "   • Proxy PID:    $PROXY_PID"
echo "   • Web PID:      $WEB_PID"
echo ""
echo "📝 Logs:"
echo "   • tail -f /tmp/nova-backend.log"
echo "   • tail -f /tmp/nova-proxy.log"
echo "   • tail -f /tmp/nova-web.log"
echo ""
echo "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Keep script running
wait
