# CORS Proxy

Simple HTTP proxy that adds CORS headers to all responses from the Nova Crew Server backend.

## Why?

The Nova Crew Server backend (port 8080) only adds CORS headers to the `/completion` endpoint. Other endpoints like `/health`, `/memory/reset`, `/models`, etc. don't have CORS headers, which causes the web interface to fail when making requests.

This proxy:
- Listens on port **8081**
- Forwards all requests to the backend (port 8080)
- Adds CORS headers to all responses
- Handles OPTIONS preflight requests

## Usage

```bash
# Start the proxy
cd samples/56-crew-server-agent/web/proxy
go run main.go
```

The proxy will start on `http://localhost:8081`

## Architecture

```
Web Interface (port 3000)
         ↓
    Fetch API
         ↓
CORS Proxy (port 8081)  ← Adds CORS headers
         ↓
Backend (port 8080)
```

## Configuration

Edit `main.go` to change:

```go
const (
    PROXY_PORT = "8081"           // Proxy listening port
    BACKEND_URL = "http://localhost:8080"  // Backend URL
)
```

## Notes

- This is a **development-only** solution
- For production, modify the Nova SDK to add CORS headers directly
- See `../FIX-CORS.md` for alternative solutions
