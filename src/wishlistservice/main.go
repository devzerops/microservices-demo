package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const defaultPort = "8098"

func main() {
	// Setup logger
	logger := log.New(os.Stdout, "[WISHLIST-SERVICE] ", log.LstdFlags)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	logger.Printf("Starting Wishlist Service on port %s", port)

	// Initialize service
	wishlistService := NewWishlistService()

	// Initialize handler
	handler := NewHandler(wishlistService, logger)

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Wishlist endpoints
	router.HandleFunc("/users/{user_id}/wishlist", handler.GetWishlistHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/wishlist/items", handler.AddItemHandler).Methods("POST")
	router.HandleFunc("/users/{user_id}/wishlist/items/{item_id}", handler.GetItemHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/wishlist/items/{item_id}", handler.UpdateItemHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/users/{user_id}/wishlist/items/{item_id}", handler.RemoveItemHandler).Methods("DELETE")

	// Statistics
	router.HandleFunc("/users/{user_id}/wishlist/stats", handler.GetStatsHandler).Methods("GET")

	// Alerts
	router.HandleFunc("/users/{user_id}/alerts", handler.GetAlertsHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/alerts/{alert_id}/read", handler.MarkAlertReadHandler).Methods("POST")

	// Sharing
	router.HandleFunc("/users/{user_id}/wishlist/share", handler.ShareWishlistHandler).Methods("POST")
	router.HandleFunc("/users/{user_id}/wishlist/share/{unshare_user_id}", handler.UnshareWishlistHandler).Methods("DELETE")
	router.HandleFunc("/users/{user_id}/wishlist/public", handler.SetPublicHandler).Methods("PUT")

	// Product price updates (system/admin endpoint)
	router.HandleFunc("/products/{product_id}/price", handler.UpdateProductPriceHandler).Methods("PUT")

	// CORS configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Printf("Wishlist Service listening on port %s", port)
	logger.Printf("Endpoints:")
	logger.Printf("  - GET    /health")
	logger.Printf("  - GET    /")
	logger.Printf("  - GET    /users/{user_id}/wishlist")
	logger.Printf("  - POST   /users/{user_id}/wishlist/items")
	logger.Printf("  - GET    /users/{user_id}/wishlist/items/{item_id}")
	logger.Printf("  - PUT    /users/{user_id}/wishlist/items/{item_id}")
	logger.Printf("  - DELETE /users/{user_id}/wishlist/items/{item_id}")
	logger.Printf("  - GET    /users/{user_id}/wishlist/stats")
	logger.Printf("  - GET    /users/{user_id}/alerts")
	logger.Printf("  - POST   /users/{user_id}/alerts/{alert_id}/read")
	logger.Printf("  - POST   /users/{user_id}/wishlist/share")
	logger.Printf("  - DELETE /users/{user_id}/wishlist/share/{unshare_user_id}")
	logger.Printf("  - PUT    /users/{user_id}/wishlist/public")
	logger.Printf("  - PUT    /products/{product_id}/price")

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "Wishlist Service",
		"version": "1.0.0",
		"features": []string{
			"Save favorite products",
			"Price drop notifications",
			"Stock availability alerts",
			"Target price tracking",
			"Priority levels (high/medium/low)",
			"Personal notes for items",
			"Wishlist sharing",
			"Public/private wishlists",
			"Wishlist statistics",
			"Sort by price, priority, discount",
		},
		"endpoints": map[string]string{
			"health":          "GET /health",
			"get_wishlist":    "GET /users/{user_id}/wishlist",
			"add_item":        "POST /users/{user_id}/wishlist/items",
			"get_item":        "GET /users/{user_id}/wishlist/items/{item_id}",
			"update_item":     "PUT /users/{user_id}/wishlist/items/{item_id}",
			"remove_item":     "DELETE /users/{user_id}/wishlist/items/{item_id}",
			"get_stats":       "GET /users/{user_id}/wishlist/stats",
			"get_alerts":      "GET /users/{user_id}/alerts",
			"mark_alert_read": "POST /users/{user_id}/alerts/{alert_id}/read",
			"share_wishlist":  "POST /users/{user_id}/wishlist/share",
			"unshare":         "DELETE /users/{user_id}/wishlist/share/{unshare_user_id}",
			"set_public":      "PUT /users/{user_id}/wishlist/public",
			"update_price":    "PUT /products/{product_id}/price",
		},
	})
}
