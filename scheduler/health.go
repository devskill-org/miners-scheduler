package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HealthServer provides HTTP endpoints for health checking and monitoring
type HealthServer struct {
	scheduler *MinerScheduler
	server    *http.Server
	port      int
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string          `json:"status"`
	Timestamp string          `json:"timestamp"`
	Version   string          `json:"version,omitempty"`
	Scheduler SchedulerHealth `json:"scheduler"`
	System    SystemHealth    `json:"system"`
}

// SchedulerHealth represents scheduler-specific health information
type SchedulerHealth struct {
	IsRunning          bool       `json:"is_running"`
	MinersCount        int        `json:"miners_count"`
	LastCheck          *time.Time `json:"last_check,omitempty"`
	HasMarketData      bool       `json:"has_market_data"`
	LastDocumentTime   *time.Time `json:"last_document_time,omitempty"`
	PriceLimit         float64    `json:"price_limit"`
	Network            string     `json:"network"`
	CheckPriceInterval string     `json:"check_price_interval"`
}

// SystemHealth represents system-level health information
type SystemHealth struct {
	Uptime     string `json:"uptime"`
	Memory     string `json:"memory,omitempty"`
	Goroutines int    `json:"goroutines,omitempty"`
}

// NewHealthServer creates a new health check server
func NewHealthServer(scheduler *MinerScheduler, port int) *HealthServer {
	if port <= 0 {
		return nil // Health server disabled
	}

	mux := http.NewServeMux()
	hs := &HealthServer{
		scheduler: scheduler,
		port:      port,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Register routes
	mux.HandleFunc("/health", hs.healthHandler)
	mux.HandleFunc("/ready", hs.readinessHandler)
	mux.HandleFunc("/status", hs.statusHandler)
	mux.HandleFunc("/", hs.rootHandler)

	return hs
}

// Start starts the health check server
func (hs *HealthServer) Start() error {
	if hs == nil {
		return nil // Health server disabled
	}

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash the main application
			fmt.Printf("Health server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully stops the health check server
func (hs *HealthServer) Stop(ctx context.Context) error {
	if hs == nil {
		return nil // Health server disabled
	}

	return hs.server.Shutdown(ctx)
}

// healthHandler handles the /health endpoint
func (hs *HealthServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := hs.scheduler.GetStatus()

	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Scheduler: SchedulerHealth{
			IsRunning:     status.IsRunning,
			MinersCount:   status.MinersCount,
			HasMarketData: status.HasMarketData,
			PriceLimit:    hs.scheduler.GetConfig().PriceLimit,
			Network:       hs.scheduler.GetConfig().Network,
		},
		System: SystemHealth{
			Uptime:     time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder
			Goroutines: 0,                                               // Placeholder - would need runtime.NumGoroutine()
		},
	}

	// Determine overall health status
	if !status.IsRunning {
		health.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// readinessHandler handles the /ready endpoint
func (hs *HealthServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := hs.scheduler.GetStatus()

	ready := map[string]any{
		"ready":     status.IsRunning,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")

	if !status.IsRunning {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(ready)
}

// statusHandler handles the /status endpoint (detailed status)
func (hs *HealthServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := hs.scheduler.GetStatus()
	miners := hs.scheduler.GetDiscoveredMiners()
	doc, _ := hs.scheduler.GetMarketData(r.Context())

	response := map[string]any{
		"scheduler_status": status,
		"miners": map[string]any{
			"count": len(miners),
			"list":  miners,
		},
		"price_data": map[string]any{
			"has_document": doc != nil,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add document info if available
	if doc != nil {
		response["price_data"].(map[string]any)["document_id"] = doc.MRID
		response["price_data"].(map[string]any)["created_at"] = doc.CreatedDateTime

		// Add current price if available
		if price, found := doc.LookupAveragePriceInHourByTime(time.Now()); found {
			response["price_data"].(map[string]any)["current_avg_price"] = price
			response["price_data"].(map[string]any)["price_vs_limit"] = map[string]any{
				"current": price,
				"limit":   hs.scheduler.GetConfig().PriceLimit,
				"action":  getPriceAction(price, hs.scheduler.GetConfig().PriceLimit),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// rootHandler handles the root endpoint
func (hs *HealthServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	endpoints := map[string]string{
		"health": "Health check endpoint",
		"ready":  "Readiness check endpoint",
		"status": "Detailed status endpoint",
	}

	response := map[string]any{
		"service":     "miners-scheduler",
		"version":     "1.0.0",
		"description": "Avalon miners scheduler based on electricity prices",
		"endpoints":   endpoints,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func getPriceAction(currentPrice, priceLimit float64) string {
	if currentPrice <= priceLimit {
		return "wake_up_standby_miners"
	}
	return "put_miners_to_standby"
}
