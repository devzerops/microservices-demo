package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var (
	analytics *AnalyticsService
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type AnalyticsService struct {
	metrics    *MetricsCollector
	events     *EventTracker
	aggregator *Aggregator
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan interface{}
}

type Event struct {
	Type      string                 `json:"type"`
	Service   string                 `json:"service"`
	UserID    string                 `json:"user_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type MetricValue struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Tags      map[string]string      `json:"tags"`
	Timestamp time.Time              `json:"timestamp"`
}

type DashboardData struct {
	Overview      OverviewMetrics            `json:"overview"`
	Services      map[string]ServiceMetrics  `json:"services"`
	RealtimeStats RealtimeStats              `json:"realtime_stats"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

type OverviewMetrics struct {
	TotalRequests     int64   `json:"total_requests"`
	ActiveUsers       int     `json:"active_users"`
	AverageLatency    float64 `json:"average_latency_ms"`
	ErrorRate         float64 `json:"error_rate"`
	RequestsPerSecond float64 `json:"requests_per_second"`
}

type ServiceMetrics struct {
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	Uptime        float64 `json:"uptime_hours"`
	RequestCount  int64   `json:"request_count"`
	ErrorCount    int64   `json:"error_count"`
	AvgLatency    float64 `json:"avg_latency_ms"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

type RealtimeStats struct {
	RequestsLast1Min   int     `json:"requests_last_1min"`
	ErrorsLast1Min     int     `json:"errors_last_1min"`
	ActiveConnections  int     `json:"active_connections"`
	TopEndpoints       []EndpointStat `json:"top_endpoints"`
	RecentEvents       []Event `json:"recent_events"`
}

type EndpointStat struct {
	Endpoint string  `json:"endpoint"`
	Count    int     `json:"count"`
	AvgTime  float64 `json:"avg_time_ms"`
}

func NewAnalyticsService() *AnalyticsService {
	as := &AnalyticsService{
		metrics:    NewMetricsCollector(),
		events:     NewEventTracker(),
		aggregator: NewAggregator(),
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 100),
	}

	// Start background workers
	go as.broadcastLoop()
	go as.aggregationLoop()

	return as
}

func (as *AnalyticsService) broadcastLoop() {
	for data := range as.broadcast {
		as.clientsMu.RLock()
		for client := range as.clients {
			err := client.WriteJSON(data)
			if err != nil {
				log.Printf("[Analytics] WebSocket write error: %v", err)
				client.Close()
				delete(as.clients, client)
			}
		}
		as.clientsMu.RUnlock()
	}
}

func (as *AnalyticsService) aggregationLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dashboard := as.GetDashboardData()
		as.broadcast <- map[string]interface{}{
			"type": "dashboard_update",
			"data": dashboard,
		}
	}
}

func (as *AnalyticsService) GetDashboardData() DashboardData {
	return DashboardData{
		Overview:      as.metrics.GetOverview(),
		Services:      as.metrics.GetServiceMetrics(),
		RealtimeStats: as.events.GetRealtimeStats(),
		UpdatedAt:     time.Now(),
	}
}

// HTTP Handlers

func trackEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid event data",
		})
		return
	}

	event.Timestamp = time.Now()

	// Track event
	analytics.events.Track(event)

	// Broadcast to connected clients
	analytics.broadcast <- map[string]interface{}{
		"type": "event",
		"data": event,
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "event tracked",
		"id":      event.Type,
	})
}

func recordMetricHandler(w http.ResponseWriter, r *http.Request) {
	var metric MetricValue

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid metric data",
		})
		return
	}

	metric.Timestamp = time.Now()

	// Record metric
	analytics.metrics.Record(metric)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "metric recorded",
		"name":    metric.Name,
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	dashboard := analytics.GetDashboardData()
	respondJSON(w, http.StatusOK, dashboard)
}

func serviceMetricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	metrics := analytics.metrics.GetServiceMetrics()

	if serviceMetrics, exists := metrics[serviceName]; exists {
		respondJSON(w, http.StatusOK, serviceMetrics)
	} else {
		respondJSON(w, http.StatusNotFound, map[string]string{
			"error": "service not found",
		})
	}
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	limit := 100
	eventType := r.URL.Query().Get("type")
	service := r.URL.Query().Get("service")

	events := analytics.events.GetEvents(limit, eventType, service)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
		"limit":  limit,
	})
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string                 `json:"service"`
		Status  string                 `json:"status"`
		Metrics map[string]interface{} `json:"metrics"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid heartbeat data",
		})
		return
	}

	analytics.metrics.RecordHeartbeat(req.Service, req.Status, req.Metrics)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "heartbeat recorded",
		"service": req.Service,
	})
}

func dashboardWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Analytics] WebSocket upgrade error: %v", err)
		return
	}

	log.Printf("[Analytics] Dashboard client connected")

	// Add client
	analytics.clientsMu.Lock()
	analytics.clients[conn] = true
	analytics.clientsMu.Unlock()

	// Send initial snapshot
	dashboard := analytics.GetDashboardData()
	conn.WriteJSON(map[string]interface{}{
		"type": "snapshot",
		"data": dashboard,
	})

	// Keep connection alive (read messages to detect disconnect)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("[Analytics] Client disconnected: %v", err)
			analytics.clientsMu.Lock()
			delete(analytics.clients, conn)
			analytics.clientsMu.Unlock()
			conn.Close()
			break
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "analytics-service",
		"timestamp": time.Now(),
		"stats": map[string]interface{}{
			"connected_clients": len(analytics.clients),
			"total_events":      analytics.events.TotalEvents(),
			"tracked_services":  len(analytics.metrics.GetServiceMetrics()),
		},
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_events":      analytics.events.TotalEvents(),
		"total_metrics":     analytics.metrics.TotalMetrics(),
		"connected_clients": len(analytics.clients),
		"uptime_hours":      time.Since(startTime).Hours(),
		"services":          analytics.metrics.GetServiceMetrics(),
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8099"
	}

	// Initialize analytics service
	analytics = NewAnalyticsService()

	// Setup router
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/events", trackEventHandler).Methods("POST")
	router.HandleFunc("/events", eventsHandler).Methods("GET")
	router.HandleFunc("/metrics", recordMetricHandler).Methods("POST")
	router.HandleFunc("/dashboard", dashboardHandler).Methods("GET")
	router.HandleFunc("/services/{service}/metrics", serviceMetricsHandler).Methods("GET")
	router.HandleFunc("/heartbeat", heartbeatHandler).Methods("POST")
	router.HandleFunc("/ws/dashboard", dashboardWebSocketHandler)
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/stats", statsHandler).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	log.Printf("[Analytics Service] Starting on port %s", port)
	log.Printf("[Analytics Service] Dashboard: GET /dashboard")
	log.Printf("[Analytics Service] WebSocket: ws://localhost:%s/ws/dashboard", port)
	log.Printf("[Analytics Service] Track Event: POST /events")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
