package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler wraps the review service
type Handler struct {
	service *ReviewService
	logger  *log.Logger
}

// NewHandler creates a new handler instance
func NewHandler(service *ReviewService, logger *log.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// HealthCheckHandler handles health check requests
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "review-service",
	})
}

// CreateReviewHandler handles review creation
func (h *Handler) CreateReviewHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReviewRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	review, err := h.service.CreateReview(req)
	if err != nil {
		h.logger.Printf("Failed to create review: %v", err)
		respondError(w, http.StatusBadRequest, "Failed to create review", err.Error())
		return
	}

	h.logger.Printf("Created review %s for product %s", review.ReviewID, review.ProductID)
	respondJSON(w, http.StatusCreated, review)
}

// GetReviewHandler handles getting a single review
func (h *Handler) GetReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID := vars["review_id"]

	review, err := h.service.GetReview(reviewID)
	if err != nil {
		h.logger.Printf("Review not found: %s", reviewID)
		respondError(w, http.StatusNotFound, "Review not found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, review)
}

// GetProductReviewsHandler handles getting all reviews for a product
func (h *Handler) GetProductReviewsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	// Parse query parameters
	minRating := 0
	if minRatingStr := r.URL.Query().Get("min_rating"); minRatingStr != "" {
		if parsed, err := strconv.Atoi(minRatingStr); err == nil {
			minRating = parsed
		}
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "recent"
	}

	reviews, err := h.service.GetReviewsByProduct(productID, minRating, sortBy)
	if err != nil {
		h.logger.Printf("Failed to get reviews for product %s: %v", productID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get reviews", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"product_id": productID,
		"reviews":    reviews,
		"total":      len(reviews),
	})
}

// UpdateReviewHandler handles review updates
func (h *Handler) UpdateReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID := vars["review_id"]

	var req UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode update request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	review, err := h.service.UpdateReview(reviewID, req)
	if err != nil {
		h.logger.Printf("Failed to update review %s: %v", reviewID, err)
		respondError(w, http.StatusBadRequest, "Failed to update review", err.Error())
		return
	}

	h.logger.Printf("Updated review %s", reviewID)
	respondJSON(w, http.StatusOK, review)
}

// DeleteReviewHandler handles review deletion
func (h *Handler) DeleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID := vars["review_id"]

	if err := h.service.DeleteReview(reviewID); err != nil {
		h.logger.Printf("Failed to delete review %s: %v", reviewID, err)
		respondError(w, http.StatusNotFound, "Failed to delete review", err.Error())
		return
	}

	h.logger.Printf("Deleted review %s", reviewID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Review deleted successfully",
	})
}

// AddReactionHandler handles adding reactions (helpful/report)
func (h *Handler) AddReactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID := vars["review_id"]

	var req struct {
		UserID       string `json:"user_id"`
		ReactionType string `json:"reaction_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.service.AddReaction(reviewID, req.UserID, req.ReactionType); err != nil {
		h.logger.Printf("Failed to add reaction: %v", err)
		respondError(w, http.StatusBadRequest, "Failed to add reaction", err.Error())
		return
	}

	h.logger.Printf("Added %s reaction to review %s by user %s", req.ReactionType, reviewID, req.UserID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Reaction added successfully",
	})
}

// RemoveReactionHandler handles removing reactions
func (h *Handler) RemoveReactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID := vars["review_id"]

	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.service.RemoveReaction(reviewID, req.UserID); err != nil {
		h.logger.Printf("Failed to remove reaction: %v", err)
		respondError(w, http.StatusBadRequest, "Failed to remove reaction", err.Error())
		return
	}

	h.logger.Printf("Removed reaction from review %s by user %s", reviewID, req.UserID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Reaction removed successfully",
	})
}

// GetProductStatsHandler handles getting product review statistics
func (h *Handler) GetProductStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	stats, err := h.service.GetProductStats(productID)
	if err != nil {
		h.logger.Printf("Failed to get stats for product %s: %v", productID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, error string, message string) {
	respondJSON(w, status, ErrorResponse{
		Error:   error,
		Message: message,
	})
}
