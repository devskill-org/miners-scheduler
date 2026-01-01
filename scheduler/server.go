package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sixdouglas/suncalc"
)

// WebServer provides HTTP endpoints for health checking, monitoring, and web UI
type WebServer struct {
	scheduler *MinerScheduler
	server    *http.Server
	port      int
	startTime time.Time
	upgrader  websocket.Upgrader
	clients   sync.Map
	broadcast chan []byte
	done      chan struct{}
}

// StatusResponse represents the health check response
type StatusResponse struct {
	Status    string          `json:"status"`
	Timestamp string          `json:"timestamp"`
	Version   string          `json:"version,omitempty"`
	Scheduler SchedulerHealth `json:"scheduler"`
	System    SystemHealth    `json:"system"`
	EMS       EMSHealth       `json:"ems"`
	Sun       SunInfo         `json:"sun"`
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

// EMSHealth represents energy management system health information
type EMSHealth struct {
	CurrentPVPower        float64 `json:"current_pv_power"`
	ESSPower              float64 `json:"ess_power"`
	ESSSOC                float64 `json:"ess_soc"`
	GridSensorStatus      uint16  `json:"grid_sensor_status"`
	GridSensorActivePower float64 `json:"grid_sensor_active_power"`
	PlantActivePower      float64 `json:"plant_active_power"`
	DCChargerOutputPower  float64 `json:"dc_charger_output_power"`
	DCChargerVehicleSOC   float64 `json:"dc_charger_vehicle_soc"`
}

// SunInfo represents solar position and timing information
type SunInfo struct {
	SolarAngle float64 `json:"solar_angle"`
	Sunrise    string  `json:"sunrise"`
	Sunset     string  `json:"sunset"`
}

// NewWebServer creates a new web server with health endpoints and static file serving
func NewWebServer(scheduler *MinerScheduler, port int) *WebServer {
	if port <= 0 {
		return nil // Health server disabled
	}

	mux := http.NewServeMux()
	hs := &WebServer{
		scheduler: scheduler,
		port:      port,
		startTime: time.Now(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		broadcast: make(chan []byte, 256),
		done:      make(chan struct{}),
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Register API routes
	mux.HandleFunc("/api/health", hs.healthHandler)
	mux.HandleFunc("/api/ready", hs.readinessHandler)
	mux.HandleFunc("/api/ws", hs.wsHandler)

	// Serve static files from web folder
	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", fs)

	return hs
}

// Start starts the web server
func (hs *WebServer) Start() error {
	if hs == nil {
		return nil // Web server disabled
	}

	// Start the broadcast handler
	go hs.handleBroadcasts()

	// Start periodic status broadcaster
	go hs.broadcastStatus()

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash the main application
			fmt.Printf("Web server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully stops the web server
func (hs *WebServer) Stop(ctx context.Context) error {
	if hs == nil {
		return nil // Web server disabled
	}

	// Signal goroutines to stop
	close(hs.done)

	// Close all WebSocket connections
	hs.clients.Range(func(key, value any) bool {
		if conn, ok := key.(*websocket.Conn); ok {
			conn.Close() //nolint:gosec
		}
		return true
	})

	return hs.server.Shutdown(ctx)
}

// healthHandler handles the /api/health endpoint
func (hs *WebServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := hs.scheduler.GetStatus()

	response := StatusResponse{
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
			Uptime:     formatUptime(time.Since(hs.startTime)),
			Goroutines: 0, // Placeholder - would need runtime.NumGoroutine()
		},
	}

	// Determine overall health status
	if !status.IsRunning {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// readinessHandler handles the /api/ready endpoint
func (hs *WebServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := json.NewEncoder(w).Encode(ready); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// wsHandler handles WebSocket connections
func (hs *WebServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := hs.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	// Register new client
	hs.clients.Store(conn, true)

	clientCount := 0
	hs.clients.Range(func(key, value any) bool {
		clientCount++
		return true
	})
	fmt.Printf("New WebSocket client connected. Total clients: %d\n", clientCount)

	// Send initial data immediately
	hs.sendStatusToClient(conn)

	// Handle client disconnection
	defer func() {
		hs.clients.Delete(conn)
		conn.Close() //nolint:gosec

		clientCount := 0
		hs.clients.Range(func(key, value any) bool {
			clientCount++
			return true
		})
		fmt.Printf("WebSocket client disconnected. Total clients: %d\n", clientCount)
	}()

	// Read messages from client (ping/pong, close)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
	}
}

// handleBroadcasts sends messages to all connected clients
func (hs *WebServer) handleBroadcasts() {
	for {
		select {
		case message := <-hs.broadcast:
			hs.clients.Range(func(key, value any) bool {
				conn, ok := key.(*websocket.Conn)
				if !ok {
					return true
				}

				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					fmt.Printf("WebSocket write error: %v\n", err)
					conn.Close() //nolint:gosec
					hs.clients.Delete(conn)
				}
				return true
			})
		case <-hs.done:
			return
		}
	}
}

// broadcastStatus periodically broadcasts status updates
func (hs *WebServer) broadcastStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hasClients := false
			hs.clients.Range(func(key, value any) bool {
				hasClients = true
				return false // Stop after finding first client
			})

			if hasClients {
				data := hs.buildStatusData()
				message, err := json.Marshal(data)
				if err != nil {
					fmt.Printf("Failed to marshal status data: %v\n", err)
					continue
				}
				hs.broadcast <- message
			}
		case <-hs.done:
			return
		}
	}
}

// sendStatusToClient sends status data to a specific client
func (hs *WebServer) sendStatusToClient(conn *websocket.Conn) {
	data := hs.buildStatusData()
	if err := conn.WriteJSON(data); err != nil {
		fmt.Printf("Failed to send initial data: %v\n", err)
	}
}

// buildStatusData builds combined health and status data
func (hs *WebServer) buildStatusData() map[string]any {
	status := hs.scheduler.GetStatus()
	miners := hs.scheduler.GetDiscoveredMiners()
	doc := hs.scheduler.GetPricesMarketData()

	// Build miners list with detailed status
	minersList := make([]map[string]any, 0, len(miners))
	minersHealthy := true

	for _, miner := range miners {
		minerStatus := "Unknown"

		if miner.LastStats != nil {
			state := miner.LastStats.State.String()
			workMode := miner.LastStats.WorkMode.String()
			minerStatus = fmt.Sprintf("%s (%s)", state, workMode)
		} else if miner.LastStatsError != nil {
			minerStatus = "Error"
			minersHealthy = false
		}

		minersList = append(minersList, map[string]any{
			"ip":     miner.Address,
			"status": minerStatus,
		})
	}

	// Determine overall health status
	overallStatus := "healthy"
	if !status.IsRunning {
		overallStatus = "unhealthy"
	} else if len(miners) > 0 && !minersHealthy {
		overallStatus = "degraded"
	}

	health := StatusResponse{
		Status:    overallStatus,
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
			Uptime:     formatUptime(time.Since(hs.startTime)),
			Goroutines: 0,
		},
	}

	info := hs.scheduler.GetPlantRunningInfo()
	if info != nil {
		health.EMS = EMSHealth{
			CurrentPVPower:        info.PhotovoltaicPower,
			ESSPower:              info.ESSPower,
			ESSSOC:                info.ESSSOC,
			GridSensorStatus:      info.GridSensorStatus,
			GridSensorActivePower: info.GridSensorActivePower,
			PlantActivePower:      info.PlantActivePower,
			DCChargerOutputPower:  info.DCChargerOutputPower,
			DCChargerVehicleSOC:   info.DCChargerVehicleSOC,
		}
	}

	// Calculate sun information
	config := hs.scheduler.GetConfig()
	now := time.Now()
	sunTimes := suncalc.GetTimes(now, config.Latitude, config.Longitude)
	sunPos := suncalc.GetPosition(now, config.Latitude, config.Longitude)

	health.Sun = SunInfo{
		SolarAngle: sunPos.Altitude * 180 / math.Pi, // Convert radians to degrees
		Sunrise:    sunTimes["sunrise"].Value.Format(time.RFC3339),
		Sunset:     sunTimes["sunset"].Value.Format(time.RFC3339),
	}

	priceData := map[string]any{
		"has_document": doc != nil,
	}

	if doc != nil {
		priceData["document_id"] = doc.MRID
		priceData["created_at"] = doc.CreatedDateTime

		if price, found := doc.LookupAveragePriceInHourByTime(time.Now()); found {
			priceData["current_avg_price"] = price
			priceData["current"] = price
			priceData["limit"] = hs.scheduler.GetConfig().PriceLimit
		}
	}

	return map[string]any{
		"type":   "status_update",
		"health": health,
		"status": map[string]any{
			"scheduler_status": status,
			"miners": map[string]any{
				"count": len(miners),
				"list":  minersList,
			},
			"price_data": priceData,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// Helper functions

// formatUptime formats a duration as a string with seconds rounded to integer
func formatUptime(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
