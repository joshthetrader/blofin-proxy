package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	BLOFIN_API_BASE = "https://openapi.blofin.com"
	DEFAULT_PORT    = "8080"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	// CORS middleware
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, ACCESS-KEY, ACCESS-SIGN, ACCESS-TIMESTAMP, ACCESS-NONCE, ACCESS-PASSPHRASE, BROKER-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Health check endpoint
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	}))

	// Root endpoint for debugging
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message":"Blofin CORS Proxy","version":"1.0","endpoints":["/health","/api/*"],"timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
			return
		}
		// Handle all /api/* routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("🔍 API route detected: %s", r.URL.Path)
			blofinProxy(w, r)
			return
		}
		// 404 for other paths
		http.NotFound(w, r)
	}))

	log.Printf("🚀 Blofin CORS Proxy starting on port %s", port)
	log.Printf("🔗 Proxying requests to: %s", BLOFIN_API_BASE)
	log.Printf("🌐 Health check: http://localhost:%s/health", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func blofinProxy(w http.ResponseWriter, r *http.Request) {
	// Keep the full path including /api prefix (BloFin expects it)
	apiPath := r.URL.Path
	
	// Build target URL - use full path as BloFin expects /api prefix
	targetURL, err := url.Parse(BLOFIN_API_BASE + apiPath)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	
	// Preserve query parameters
	targetURL.RawQuery = r.URL.RawQuery

	// Create proxy request with same method and body
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Forward all headers (including authentication headers)
	for name, values := range r.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Set a reasonable timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request to Blofin API
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("❌ Proxy request failed: %v", err)
		http.Error(w, "Proxy request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (except hop-by-hop)
	for name, values := range resp.Header {
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set response status
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("❌ Failed to copy response body: %v", err)
	}

	// Log requests for debugging (like Netlify proxy)
	log.Printf("🔗 %s %s -> %s (Status: %d)", r.Method, r.URL.Path, targetURL.String(), resp.StatusCode)
}

// HTTP hop-by-hop headers that should not be forwarded
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
	
	header = strings.ToLower(header)
	for _, h := range hopByHopHeaders {
		if strings.ToLower(h) == header {
			return true
		}
	}
	return false
}
