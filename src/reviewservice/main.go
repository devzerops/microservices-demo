package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const defaultPort = "8096"

func main() {
	// Setup logger
	logger := log.New(os.Stdout, "[REVIEW-SERVICE] ", log.LstdFlags)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	logger.Printf("Starting Review Service on port %s", port)

	// Initialize service
	reviewService := NewReviewService()

	// Initialize handler
	handler := NewHandler(reviewService, logger)

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Review endpoints
	router.HandleFunc("/reviews", handler.CreateReviewHandler).Methods("POST")
	router.HandleFunc("/reviews/{review_id}", handler.GetReviewHandler).Methods("GET")
	router.HandleFunc("/reviews/{review_id}", handler.UpdateReviewHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/reviews/{review_id}", handler.DeleteReviewHandler).Methods("DELETE")

	// Product reviews
	router.HandleFunc("/products/{product_id}/reviews", handler.GetProductReviewsHandler).Methods("GET")
	router.HandleFunc("/products/{product_id}/stats", handler.GetProductStatsHandler).Methods("GET")

	// Reactions
	router.HandleFunc("/reviews/{review_id}/reactions", handler.AddReactionHandler).Methods("POST")
	router.HandleFunc("/reviews/{review_id}/reactions", handler.RemoveReactionHandler).Methods("DELETE")

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

	logger.Printf("Review Service listening on port %s", port)
	logger.Printf("Endpoints:")
	logger.Printf("  - GET    /health")
	logger.Printf("  - GET    /")
	logger.Printf("  - POST   /reviews")
	logger.Printf("  - GET    /reviews/{review_id}")
	logger.Printf("  - PUT    /reviews/{review_id}")
	logger.Printf("  - DELETE /reviews/{review_id}")
	logger.Printf("  - GET    /products/{product_id}/reviews")
	logger.Printf("  - GET    /products/{product_id}/stats")
	logger.Printf("  - POST   /reviews/{review_id}/reactions")
	logger.Printf("  - DELETE /reviews/{review_id}/reactions")

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "Review & Rating Service",
		"version": "1.0.0",
		"features": []string{
			"Product reviews with 1-5 star ratings",
			"Review CRUD operations",
			"Helpful/Report reactions",
			"Product rating statistics",
			"Review filtering and sorting",
			"Verified purchase indicator",
		},
		"endpoints": map[string]string{
			"health":          "GET /health",
			"create_review":   "POST /reviews",
			"get_review":      "GET /reviews/{review_id}",
			"update_review":   "PUT /reviews/{review_id}",
			"delete_review":   "DELETE /reviews/{review_id}",
			"product_reviews": "GET /products/{product_id}/reviews",
			"product_stats":   "GET /products/{product_id}/stats",
			"add_reaction":    "POST /reviews/{review_id}/reactions",
			"remove_reaction": "DELETE /reviews/{review_id}/reactions",
		},
	})
}
