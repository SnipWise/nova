// CORS Proxy for Nova Crew Server
// This is a simple reverse proxy that adds CORS headers to all requests
//
// Usage:
//   1. Start your crew server on port 8080: go run main.go
//   2. Start this proxy on port 8081: go run cors-proxy.go
//   3. Update web/js/api.js to use http://localhost:8081
//
// This is a development-only solution. For production, add CORS to the SDK itself.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	// Port to listen on
	PROXY_PORT = "8081"
	// Backend server URL
	BACKEND_URL = "http://localhost:8080"
)

func main() {
	http.HandleFunc("/", handleProxy)

	fmt.Printf("🔄 CORS Proxy starting on http://localhost:%s\n", PROXY_PORT)
	fmt.Printf("📡 Proxying to: %s\n", BACKEND_URL)
	fmt.Printf("🌐 Update API_BASE_URL in web/js/api.js to: http://localhost:%s\n", PROXY_PORT)

	if err := http.ListenAndServe(":"+PROXY_PORT, nil); err != nil {
		log.Fatal(err)
	}
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Build backend URL
	backendURL := BACKEND_URL + r.URL.Path
	if r.URL.RawQuery != "" {
		backendURL += "?" + r.URL.RawQuery
	}

	// Create new request to backend
	proxyReq, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send request to backend
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error forwarding request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (except CORS - we set our own)
	for key, values := range resp.Header {
		// Skip CORS headers from backend
		if strings.HasPrefix(key, "Access-Control-") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Check if this is a streaming response (SSE)
	isSSE := strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")

	if isSSE {
		// For SSE, we need to flush after each chunk
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Stream the response with immediate flushing
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := w.Write(buf[:n])
				if writeErr != nil {
					log.Printf("Error writing response: %v", writeErr)
					return
				}
				flusher.Flush() // Flush immediately for SSE
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error reading response: %v", err)
				return
			}
		}
	} else {
		// For regular responses, use io.Copy
		io.Copy(w, resp.Body)
	}
}
