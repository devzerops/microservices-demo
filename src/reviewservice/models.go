package main

import (
	"time"
)

// Review represents a product review
type Review struct {
	ReviewID         string    `json:"review_id"`
	ProductID        string    `json:"product_id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	Rating           int       `json:"rating"` // 1-5 stars
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	VerifiedPurchase bool      `json:"verified_purchase"`
	Images           []string  `json:"images,omitempty"`
	HelpfulCount     int       `json:"helpful_count"`
	ReportCount      int       `json:"report_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ReviewStats represents aggregate statistics for a product
type ReviewStats struct {
	ProductID      string             `json:"product_id"`
	TotalReviews   int                `json:"total_reviews"`
	AverageRating  float64            `json:"average_rating"`
	RatingBreakdown map[int]int       `json:"rating_breakdown"` // star -> count
	LastUpdated    time.Time          `json:"last_updated"`
}

// CreateReviewRequest represents the request to create a review
type CreateReviewRequest struct {
	ProductID        string   `json:"product_id"`
	UserID           string   `json:"user_id"`
	Username         string   `json:"username"`
	Rating           int      `json:"rating"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	VerifiedPurchase bool     `json:"verified_purchase"`
	Images           []string `json:"images,omitempty"`
}

// UpdateReviewRequest represents the request to update a review
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty"`
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

// ReviewReaction represents user reactions to reviews (helpful/report)
type ReviewReaction struct {
	ReviewID     string    `json:"review_id"`
	UserID       string    `json:"user_id"`
	ReactionType string    `json:"reaction_type"` // "helpful" or "report"
	CreatedAt    time.Time `json:"created_at"`
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
