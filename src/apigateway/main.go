package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

var (
	services = map[string]string{
		"visualsearch": getEnv("VISUAL_SEARCH_URL", "http://visualsearch:8093"),
		"gamification": getEnv("GAMIFICATION_URL", "http://gamification:8094"),
		"inventory":    getEnv("INVENTORY_URL", "http://inventory:8092"),
		"pwa":          getEnv("PWA_URL", "http://pwa:8095"),
		"search":       getEnv("SEARCH_URL", "http://search:8097"),
		"analytics":    getEnv("ANALYTICS_URL", "http://analytics:8099"),
		"review":       getEnv("REVIEW_URL", "http://review:8096"),
		"wishlist":     getEnv("WISHLIST_URL", "http://wishlist:8098"),
	}

	rateLimiters = make(map[string]*rate.Limiter)
	rateLimiterMu sync.RWMutex
)

type Gateway struct {
	proxies map[string]*httputil.ReverseProxy
}

func NewGateway() *Gateway {
	g := &Gateway{
		proxies: make(map[string]*httputil.ReverseProxy),
	}

	// Create reverse proxies for each service
	for name, serviceURL := range services {
		target, err := url.Parse(serviceURL)
		if err != nil {
			log.Printf("Failed to parse URL for %s: %v", name, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = g.errorHandler

		g.proxies[name] = proxy
		log.Printf("[Gateway] Registered proxy for %s -> %s", name, serviceURL)
	}

	return g
}

func (g *Gateway) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("[Gateway] Proxy error for %s: %v", r.URL.Path, err)

	respondJSON(w, http.StatusBadGateway, map[string]interface{}{
		"error":   "service_unavailable",
		"message": "The requested service is currently unavailable",
		"path":    r.URL.Path,
	})
}

func (g *Gateway) serveProxy(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy, exists := g.proxies[serviceName]
		if !exists {
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error": "service not found",
			})
			return
		}

		// Remove service prefix from path
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+serviceName)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		// Add custom headers
		r.Header.Set("X-Forwarded-By", "api-gateway")
		r.Header.Set("X-Gateway-Version", "1.0")

		// Proxy the request
		proxy.ServeHTTP(w, r)
	}
}

// Rate limiting middleware
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := getClientIP(r)

		// Get or create rate limiter for this IP
		rateLimiterMu.Lock()
		limiter, exists := rateLimiters[clientIP]
		if !exists {
			// 100 requests per minute per IP
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 100)
			rateLimiters[clientIP] = limiter
		}
		rateLimiterMu.Unlock()

		// Check rate limit
		if !limiter.Allow() {
			respondJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
				"limit":   "100 requests per minute",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		log.Printf("[Gateway] %s %s %d %v %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			getClientIP(r),
		)

		// Send metrics to analytics if available
		go sendMetricsToAnalytics(r, wrapped.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Health check aggregator
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	type ServiceHealth struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		URL       string `json:"url"`
		Available bool   `json:"available"`
	}

	results := make([]ServiceHealth, 0, len(services))
	var wg sync.WaitGroup

	for name, serviceURL := range services {
		wg.Add(1)

		go func(name, serviceURL string) {
			defer wg.Done()

			health := ServiceHealth{
				Name:      name,
				URL:       serviceURL,
				Status:    "unknown",
				Available: false,
			}

			// Try to reach service health endpoint
			resp, err := http.Get(serviceURL + "/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				health.Status = "healthy"
				health.Available = true
				resp.Body.Close()
			} else {
				health.Status = "unhealthy"
				health.Available = false
			}

			results = append(results, health)
		}(name, serviceURL)
	}

	wg.Wait()

	// Check overall health
	allHealthy := true
	for _, result := range results {
		if !result.Available {
			allHealthy = false
			break
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	respondJSON(w, status, map[string]interface{}{
		"status":    map[bool]string{true: "healthy", false: "degraded"}[allHealthy],
		"timestamp": time.Now(),
		"services":  results,
		"gateway": map[string]interface{}{
			"version": "1.0",
			"uptime":  time.Since(startTime).String(),
		},
	})
}

// Routes info
func routesHandler(w http.ResponseWriter, r *http.Request) {
	type Route struct {
		Service string `json:"service"`
		Path    string `json:"path"`
		Target  string `json:"target"`
	}

	routes := []Route{}
	for name, url := range services {
		routes = append(routes, Route{
			Service: name,
			Path:    "/" + name + "/*",
			Target:  url,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"routes":    routes,
		"total":     len(routes),
		"timestamp": time.Now(),
	})
}

// Gateway stats
func statsHandler(w http.ResponseWriter, r *http.Request) {
	rateLimiterMu.RLock()
	activeIPs := len(rateLimiters)
	rateLimiterMu.RUnlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"gateway": map[string]interface{}{
			"version":         "1.0",
			"uptime":          time.Since(startTime).String(),
			"active_ips":      activeIPs,
			"services_count":  len(services),
		},
		"rate_limiting": map[string]interface{}{
			"enabled":       true,
			"limit":         "100 requests per minute per IP",
			"tracked_ips":   activeIPs,
		},
	})
}

func sendMetricsToAnalytics(r *http.Request, statusCode int, duration time.Duration) {
	analyticsURL := services["analytics"]
	if analyticsURL == "" {
		return
	}

	event := map[string]interface{}{
		"type":    "request",
		"service": "gateway",
		"data": map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": statusCode,
			"latency_ms":  duration.Milliseconds(),
			"client_ip":   getClientIP(r),
		},
	}

	body, _ := json.Marshal(event)
	http.Post(analyticsURL+"/events", "application/json", strings.NewReader(string(body)))
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

var startTime = time.Now()

func main() {
	port := getEnv("PORT", "8080")

	// Initialize gateway
	gateway := NewGateway()

	// Setup router
	router := mux.NewRouter()

	// Gateway endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/routes", routesHandler).Methods("GET")
	router.HandleFunc("/stats", statsHandler).Methods("GET")

	// Service proxies
	for serviceName := range services {
		router.PathPrefix("/" + serviceName).Handler(gateway.serveProxy(serviceName))
	}

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name":    "API Gateway",
			"version": "1.0",
			"status":  "running",
			"endpoints": map[string]string{
				"health":  "/health",
				"routes":  "/routes",
				"stats":   "/stats",
			},
			"services": services,
		})
	}).Methods("GET")

	// Apply middleware
	handler := loggingMiddleware(rateLimitMiddleware(router))

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	finalHandler := c.Handler(handler)

	// Start server
	log.Printf("[API Gateway] Starting on port %s", port)
	log.Printf("[API Gateway] Registered %d services", len(services))
	log.Printf("[API Gateway] Rate limiting: 100 requests/minute per IP")

	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// WebSocket proxy handler (for services with WebSocket support)
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	serviceName := mux.Vars(r)["service"]

	serviceURL, exists := services[serviceName]
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Parse target URL
	target, err := url.Parse(serviceURL)
	if err != nil {
		http.Error(w, "Invalid service URL", http.StatusInternalServerError)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Update request
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = target.Host

	// Proxy the request
	proxy.ServeHTTP(w, r)
}
