#!/bin/bash

# Nova Crew Server - Web Interface Launcher
# This script starts a simple HTTP server to serve the web interface

echo "🚀 Starting Nova Crew Server Web Interface..."
echo ""
echo "Make sure the Go server is running on http://localhost:8080"
echo "If not, run: cd .. && go run main.go"
echo ""
echo "Starting web server on http://localhost:3000"
echo "Press Ctrl+C to stop"
echo ""

# Try Python3 first (most common)
if command -v python3 &> /dev/null; then
    echo "✓ Using Python 3"
    python3 -m http.server 3000
# Try Python 2 as fallback
elif command -v python &> /dev/null; then
    echo "✓ Using Python 2"
    python -m SimpleHTTPServer 3000
# Try PHP
elif command -v php &> /dev/null; then
    echo "✓ Using PHP"
    php -S localhost:3000
# Try Node.js npx
elif command -v npx &> /dev/null; then
    echo "✓ Using Node.js http-server"
    npx http-server -p 3000
else
    echo "❌ Error: No suitable HTTP server found"
    echo ""
    echo "Please install one of the following:"
    echo "  - Python 3: https://www.python.org/"
    echo "  - Node.js: https://nodejs.org/"
    echo "  - PHP: https://www.php.net/"
    exit 1
fi
