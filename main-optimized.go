package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	BLOFIN_API_BASE = "https://openapi.blofin.com"
	DEFAULT_PORT    = "8080"
	
	// Optimized for high concurrency
	MAX_IDLE_CONNS        = 1000
	MAX_CONNS_PER_HOST    = 500
	IDLE_CONN_TIMEOUT     = 90 * time.Second
	TLS_HANDSHAKE_TIMEOUT = 10 * time.Second
	RESPONSE_HEADER_TIMEOUT = 10 * time.Second
)

var (
	// High-performance HTTP client with connection pooling
	httpClient *http.Client
	
	// Request metrics
	requestCount int64
	mu           sync.RWMutex
)

func init() {
	// Set GOMAXPROCS to use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Create optimized HTTP transport
	transport := &http.Transport{
		MaxIdleConns:        MAX_IDLE_CONNS,
		MaxIdleConnsPerHost: MAX_CONNS_PER_HOST,
		IdleConnTimeout:     IDLE_CONN_TIMEOUT,
		TLSHandshakeTimeout: TLS_HANDSHAKE_TIMEOUT,
		ResponseHeaderTimeout: RESPONSE_HEADER_TIMEOUT,
		
		// Enable HTTP/2
		ForceAttemptHTTP2: true,
		
		// Optimize for high throughput
		WriteBufferSize: 32 * 1024,
		ReadBufferSize:  32 * 1024,
	}
	
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	// CORS middleware with connection reuse
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, ACCESS-KEY, ACCESS-SIGN, ACCESS-TIMESTAMP, ACCESS-NONCE, ACCESS-PASSPHRASE, BROKER-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests quickly
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Health check with metrics
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		count := requestCount
		mu.RUnlock()
		
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","timestamp":"%s","requests_served":%d,"goroutines":%d}`, 
			time.Now().UTC().Format(time.RFC3339), count, runtime.NumGoroutine())
	}))

	// Metrics endpoint
	http.HandleFunc("/metrics", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"requests_total": %d,
			"goroutines": %d,
			"memory_alloc_mb": %.2f,
			"memory_sys_mb": %.2f,
			"gc_runs": %d,
			"cpu_cores": %d
		}`, requestCount, runtime.NumGoroutine(), 
		float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024, 
		m.NumGC, runtime.NumCPU())
	}))

	// Optimized Blofin API proxy
	http.HandleFunc("/api/", corsMiddleware(blofinProxyOptimized))

	// Configure server for high concurrency
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		
		// Optimize for high connection count
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Printf("🚀 Optimized Blofin CORS Proxy starting on port %s", port)
	log.Printf("🔗 Proxying requests to: %s", BLOFIN_API_BASE)
	log.Printf("⚡ Max connections per host: %d", MAX_CONNS_PER_HOST)
	log.Printf("🧠 Using %d CPU cores", runtime.NumCPU())
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func blofinProxyOptimized(w http.ResponseWriter, r *http.Request) {
	// Increment request counter
	mu.Lock()
	requestCount++
	mu.Unlock()
	
	// Build target URL - preserve the full path and query parameters
	targetURL, err := url.Parse(BLOFIN_API_BASE + r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	
	// Preserve query parameters
	targetURL.RawQuery = r.URL.RawQuery

	// Create proxy request with context for timeout control
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL.String(), r.Body)
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

	// Make the request to Blofin API using optimized client
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		// Don't log every error in production to avoid log spam
		if os.Getenv("DEBUG") == "true" {
			log.Printf("❌ Proxy request failed: %v", err)
		}
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

	// Copy response body efficiently
	_, err = io.Copy(w, resp.Body)
	if err != nil && os.Getenv("DEBUG") == "true" {
		log.Printf("❌ Failed to copy response body: %v", err)
	}
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
