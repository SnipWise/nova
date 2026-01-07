# Quick Start Guide

## 🚀 One-Command Start (Recommended)

The easiest way to start everything:

```bash
cd samples/56-crew-server-agent/web
./start-all.sh
```

This automatically starts:
- ✅ Backend server (port 8080)
- ✅ CORS proxy (port 8081)
- ✅ Web interface (port 3000)

Then open: **http://localhost:3000**

Press `Ctrl+C` to stop all services.

---

## 📋 Manual Start (3 Terminals)

If you prefer to start services individually:

### Step 1: Start the Nova Crew Server

In terminal 1, start the Go server:

```bash
cd samples/56-crew-server-agent
go run main.go
```

You should see:
```
🚀 Server starting on http://localhost:8080
```

### Step 2: Start the CORS Proxy

In terminal 2, start the CORS proxy:

```bash
cd samples/56-crew-server-agent/web
go run cors-proxy.go
```

You should see:
```
🔄 CORS Proxy starting on http://localhost:8081
```

**Why?** The proxy adds CORS headers to all API endpoints (not just `/completion`).

### Step 3: Serve the Web Interface

### Option A: Python 3 (Recommended - Usually pre-installed)

In terminal 3:

```bash
cd samples/56-crew-server-agent/web
python3 -m http.server 3000
```

Then open: **http://localhost:3000**

### Option B: Node.js

In terminal 2:

```bash
cd samples/56-crew-server-agent/web
npx http-server -p 3000
```

Then open: **http://localhost:3000**

### Option C: PHP

In terminal 2:

```bash
cd samples/56-crew-server-agent/web
php -S localhost:3000
```

Then open: **http://localhost:3000**

### Option D: Go (if you prefer)

In terminal 2:

```bash
cd samples/56-crew-server-agent/web
go run -mod=mod github.com/shurcooL/goexec@latest 'http.ListenAndServe(":3000", http.FileServer(http.Dir(".")))'
```

Then open: **http://localhost:3000**

## Step 3: Use the Interface

1. Type a message in the input field
2. Press Enter or click "Send"
3. Watch the AI response stream in real-time with markdown formatting!

## Testing Different Agents

The system automatically routes questions to specialized agents:

### Coder Agent
```
Write a Python function to calculate fibonacci numbers
```

### Thinker Agent
```
What is the meaning of consciousness?
```

### Cook Agent
```
Give me a recipe for chocolate chip cookies
```

### Tool Calling
```
Say hello to Alice
```

```
Calculate the sum of 42 and 58
```

## Features to Try

- **Markdown**: The AI responses support full markdown including code blocks
- **Syntax Highlighting**: Code blocks are automatically highlighted
- **Streaming**: Responses appear token-by-token in real-time
- **Function Calling**: Try commands that trigger tools and validate them
- **Memory Management**: Use "Clear Memory" to reset the conversation
- **Context Tracking**: Watch the context size update in real-time

## Troubleshooting

### "Failed to connect to server"

**Problem**: Web interface can't reach the Go server

**Solutions**:
1. Make sure the Go server is running on port 8080
2. Check the console for error messages
3. Verify the API URL in `js/api.js` is `http://localhost:8080`

### Code blocks not highlighting

**Problem**: Code appears without colors

**Solutions**:
1. Check your internet connection (CDN resources)
2. Open browser console and look for loading errors
3. Specify language in code blocks (```python, ```go, etc.)

### Streaming stops mid-response

**Problem**: Response cuts off unexpectedly

**Solutions**:
1. Check Go server logs for errors
2. Try the "Stop" button and resend
3. Clear memory and try again

## Development Tips

- **Console Logging**: Open browser DevTools (F12) to see detailed logs
- **Network Tab**: Monitor SSE streaming events in real-time
- **JSON Messages**: Click "View Messages" to see raw conversation data

Enjoy chatting with your Nova Crew! 🚀
