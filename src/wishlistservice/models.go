package main

import (
	"time"
)

// WishlistItem represents a single item in a user's wishlist
type WishlistItem struct {
	ItemID            string    `json:"item_id"`
	UserID            string    `json:"user_id"`
	ProductID         string    `json:"product_id"`
	ProductName       string    `json:"product_name,omitempty"`
	ProductImageURL   string    `json:"product_image_url,omitempty"`
	CurrentPrice      float64   `json:"current_price"`
	OriginalPrice     float64   `json:"original_price"`
	TargetPrice       float64   `json:"target_price,omitempty"` // Alert when price drops to this
	Priority          string    `json:"priority"`                // high, medium, low
	Notes             string    `json:"notes,omitempty"`
	NotifyOnPriceDrop bool      `json:"notify_on_price_drop"`
	NotifyOnRestock   bool      `json:"notify_on_restock"`
	InStock           bool      `json:"in_stock"`
	AddedAt           time.Time `json:"added_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Wishlist represents a user's complete wishlist
type Wishlist struct {
	UserID     string          `json:"user_id"`
	Items      []*WishlistItem `json:"items"`
	TotalItems int             `json:"total_items"`
	SharedWith []string        `json:"shared_with,omitempty"` // User IDs who can view this wishlist
	IsPublic   bool            `json:"is_public"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// AddItemRequest represents the request to add an item to wishlist
type AddItemRequest struct {
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name,omitempty"`
	ProductImageURL   string  `json:"product_image_url,omitempty"`
	CurrentPrice      float64 `json:"current_price"`
	TargetPrice       float64 `json:"target_price,omitempty"`
	Priority          string  `json:"priority,omitempty"` // default: medium
	Notes             string  `json:"notes,omitempty"`
	NotifyOnPriceDrop bool    `json:"notify_on_price_drop"`
	NotifyOnRestock   bool    `json:"notify_on_restock"`
	InStock           bool    `json:"in_stock"`
}

// UpdateItemRequest represents the request to update a wishlist item
type UpdateItemRequest struct {
	TargetPrice       *float64 `json:"target_price,omitempty"`
	Priority          *string  `json:"priority,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	NotifyOnPriceDrop *bool    `json:"notify_on_price_drop,omitempty"`
	NotifyOnRestock   *bool    `json:"notify_on_restock,omitempty"`
}

// PriceUpdate represents a price change for a product
type PriceUpdate struct {
	ProductID    string    `json:"product_id"`
	OldPrice     float64   `json:"old_price"`
	NewPrice     float64   `json:"new_price"`
	PriceChange  float64   `json:"price_change"`
	PercentChange float64   `json:"percent_change"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PriceAlert represents an alert for a price drop
type PriceAlert struct {
	AlertID      string       `json:"alert_id"`
	UserID       string       `json:"user_id"`
	ItemID       string       `json:"item_id"`
	ProductID    string       `json:"product_id"`
	ProductName  string       `json:"product_name"`
	PriceUpdate  PriceUpdate  `json:"price_update"`
	TargetPrice  float64      `json:"target_price,omitempty"`
	AlertType    string       `json:"alert_type"` // price_drop, restock, target_reached
	IsRead       bool         `json:"is_read"`
	CreatedAt    time.Time    `json:"created_at"`
}

// ShareRequest represents the request to share a wishlist
type ShareRequest struct {
	ShareWithUserID string `json:"share_with_user_id"`
	Permission      string `json:"permission,omitempty"` // view, edit (future)
}

// WishlistStats represents statistics about a user's wishlist
type WishlistStats struct {
	UserID              string  `json:"user_id"`
	TotalItems          int     `json:"total_items"`
	TotalValue          float64 `json:"total_value"`
	AveragePriceDrop    float64 `json:"average_price_drop"`
	ItemsOnSale         int     `json:"items_on_sale"`
	OutOfStockItems     int     `json:"out_of_stock_items"`
	HighPriorityItems   int     `json:"high_priority_items"`
	MediumPriorityItems int     `json:"medium_priority_items"`
	LowPriorityItems    int     `json:"low_priority_items"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
