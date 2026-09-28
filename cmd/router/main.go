package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"

	"github.com/username/adaptive-request-fabric/internal/cache"
	"github.com/username/adaptive-request-fabric/internal/worker"
)

var backends []string
var counter uint64
var fabricCache *cache.MultiLevelCache
var requestPool *worker.Pool

func init() {
	b := os.Getenv("BACKENDS")
	if b == "" {
		backends = []string{
			"http://localhost:8081",
			"http://localhost:8082",
			"http://localhost:8083",
		}
	} else {
		backends = strings.Split(b, ",")
	}
}

func main() {
	port := "8000"
	
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	pgConn := os.Getenv("POSTGRES_CONN")
	if pgConn == "" {
		pgConn = "postgres://postgres:postgres@localhost:5432/fabric?sslmode=disable"
	}

	var err error
	fabricCache, err = cache.NewMultiLevelCache(redisAddr, pgConn)
	if err != nil {
		log.Printf("Warning: Failed to init cache (using fallback memory mode or expecting failure later): %v", err)
	}

	// Phase 4 & 5: Initialize a bounded worker pool
	// 10 concurrent workers, max 50 requests waiting in queue.
	requestPool = worker.NewPool(10, 50)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := requestPool.Submit(w, r, handleRequest)
		if err == worker.ErrQueueFull {
			http.Error(w, `{"error": "Service Overloaded (Backpressure)"}`, http.StatusServiceUnavailable)
		}
	})

	log.Printf("Starting router on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	cacheKey := r.URL.Path

	// Try Cache
	cachedVal, level, err := fabricCache.Get(r.Context(), cacheKey)
	if err == nil {
		// Parse and inject cache level
		var resp map[string]interface{}
		json.Unmarshal([]byte(cachedVal), &resp)
		resp["cache"] = level
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Simple round robin routing
	idx := atomic.AddUint64(&counter, 1) % uint64(len(backends))
	target := backends[idx]

	targetURL, _ := url.Parse(target + r.URL.Path)

	proxyReq, _ := http.NewRequest(r.Method, targetURL.String(), r.Body)
	for h, val := range r.Header {
		proxyReq.Header[h] = val
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Backend error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	// Inject cache "miss" and populate cache
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResp); err == nil {
		jsonResp["cache"] = "miss"
		
		// Save original without "miss" tag
		fabricCache.Set(context.Background(), cacheKey, string(bodyBytes))
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResp)
		return
	}

	for h, val := range resp.Header {
		w.Header()[h] = val
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)
}
