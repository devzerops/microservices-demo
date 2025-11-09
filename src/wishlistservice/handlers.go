package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler wraps the wishlist service
type Handler struct {
	service *WishlistService
	logger  *log.Logger
}

// NewHandler creates a new handler instance
func NewHandler(service *WishlistService, logger *log.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// HealthCheckHandler handles health check requests
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "wishlist-service",
	})
}

// AddItemHandler handles adding an item to wishlist
func (h *Handler) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	item, err := h.service.AddItem(userID, req)
	if err != nil {
		h.logger.Printf("Failed to add item for user %s: %v", userID, err)
		respondError(w, http.StatusBadRequest, "Failed to add item", err.Error())
		return
	}

	h.logger.Printf("Added item %s to wishlist for user %s", item.ItemID, userID)
	respondJSON(w, http.StatusCreated, item)
}

// RemoveItemHandler handles removing an item from wishlist
func (h *Handler) RemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	itemID := vars["item_id"]

	if err := h.service.RemoveItem(userID, itemID); err != nil {
		h.logger.Printf("Failed to remove item %s for user %s: %v", itemID, userID, err)
		respondError(w, http.StatusNotFound, "Failed to remove item", err.Error())
		return
	}

	h.logger.Printf("Removed item %s from wishlist for user %s", itemID, userID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Item removed from wishlist",
	})
}

// GetWishlistHandler handles getting a user's wishlist
func (h *Handler) GetWishlistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	sortBy := r.URL.Query().Get("sort_by")

	wishlist, err := h.service.GetWishlist(userID, sortBy)
	if err != nil {
		h.logger.Printf("Failed to get wishlist for user %s: %v", userID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get wishlist", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

// GetItemHandler handles getting a specific wishlist item
func (h *Handler) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	itemID := vars["item_id"]

	item, err := h.service.GetItem(userID, itemID)
	if err != nil {
		h.logger.Printf("Failed to get item %s for user %s: %v", itemID, userID, err)
		respondError(w, http.StatusNotFound, "Failed to get item", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// UpdateItemHandler handles updating a wishlist item
func (h *Handler) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	itemID := vars["item_id"]

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode update request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	item, err := h.service.UpdateItem(userID, itemID, req)
	if err != nil {
		h.logger.Printf("Failed to update item %s for user %s: %v", itemID, userID, err)
		respondError(w, http.StatusBadRequest, "Failed to update item", err.Error())
		return
	}

	h.logger.Printf("Updated item %s for user %s", itemID, userID)
	respondJSON(w, http.StatusOK, item)
}

// UpdateProductPriceHandler handles updating product prices (admin/system endpoint)
func (h *Handler) UpdateProductPriceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	var req struct {
		NewPrice float64 `json:"new_price"`
		InStock  bool    `json:"in_stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	alerts, err := h.service.UpdateProductPrice(productID, req.NewPrice, req.InStock)
	if err != nil {
		h.logger.Printf("Failed to update price for product %s: %v", productID, err)
		respondError(w, http.StatusInternalServerError, "Failed to update price", err.Error())
		return
	}

	h.logger.Printf("Updated price for product %s, generated %d alerts", productID, len(alerts))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"alerts_created": len(alerts),
		"alerts":         alerts,
	})
}

// GetAlertsHandler handles getting alerts for a user
func (h *Handler) GetAlertsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	unreadOnly := false
	if unreadStr := r.URL.Query().Get("unread_only"); unreadStr != "" {
		if parsed, err := strconv.ParseBool(unreadStr); err == nil {
			unreadOnly = parsed
		}
	}

	alerts, err := h.service.GetAlerts(userID, unreadOnly)
	if err != nil {
		h.logger.Printf("Failed to get alerts for user %s: %v", userID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get alerts", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"alerts":  alerts,
		"total":   len(alerts),
	})
}

// MarkAlertReadHandler handles marking an alert as read
func (h *Handler) MarkAlertReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	alertID := vars["alert_id"]

	if err := h.service.MarkAlertAsRead(userID, alertID); err != nil {
		h.logger.Printf("Failed to mark alert %s as read for user %s: %v", alertID, userID, err)
		respondError(w, http.StatusNotFound, "Failed to mark alert as read", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Alert marked as read",
	})
}

// ShareWishlistHandler handles sharing a wishlist
func (h *Handler) ShareWishlistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req ShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.service.ShareWishlist(userID, req.ShareWithUserID); err != nil {
		h.logger.Printf("Failed to share wishlist for user %s: %v", userID, err)
		respondError(w, http.StatusBadRequest, "Failed to share wishlist", err.Error())
		return
	}

	h.logger.Printf("User %s shared wishlist with user %s", userID, req.ShareWithUserID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Wishlist shared successfully",
	})
}

// UnshareWishlistHandler handles unsharing a wishlist
func (h *Handler) UnshareWishlistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	unshareUserID := vars["unshare_user_id"]

	if err := h.service.UnshareWishlist(userID, unshareUserID); err != nil {
		h.logger.Printf("Failed to unshare wishlist for user %s: %v", userID, err)
		respondError(w, http.StatusNotFound, "Failed to unshare wishlist", err.Error())
		return
	}

	h.logger.Printf("User %s unshared wishlist with user %s", userID, unshareUserID)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Wishlist unshared successfully",
	})
}

// SetPublicHandler handles setting wishlist public/private
func (h *Handler) SetPublicHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req struct {
		IsPublic bool `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.service.SetPublic(userID, req.IsPublic); err != nil {
		h.logger.Printf("Failed to set public status for user %s: %v", userID, err)
		respondError(w, http.StatusNotFound, "Failed to update wishlist", err.Error())
		return
	}

	h.logger.Printf("User %s set wishlist public status to %v", userID, req.IsPublic)
	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Wishlist visibility updated",
	})
}

// GetStatsHandler handles getting wishlist statistics
func (h *Handler) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	stats, err := h.service.GetStats(userID)
	if err != nil {
		h.logger.Printf("Failed to get stats for user %s: %v", userID, err)
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
