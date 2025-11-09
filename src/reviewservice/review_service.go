package main

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ReviewService handles all review-related operations
type ReviewService struct {
	reviews       map[string]*Review        // reviewID -> Review
	productStats  map[string]*ReviewStats   // productID -> ReviewStats
	userReactions map[string]map[string]string // reviewID -> userID -> reactionType
	mu            sync.RWMutex
}

// NewReviewService creates a new review service instance
func NewReviewService() *ReviewService {
	return &ReviewService{
		reviews:       make(map[string]*Review),
		productStats:  make(map[string]*ReviewStats),
		userReactions: make(map[string]map[string]string),
	}
}

// CreateReview creates a new review
func (rs *ReviewService) CreateReview(req CreateReviewRequest) (*Review, error) {
	// Validate input
	if err := validateCreateReviewRequest(req); err != nil {
		return nil, err
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Generate review ID
	reviewID := uuid.New().String()

	// Create review
	review := &Review{
		ReviewID:         reviewID,
		ProductID:        req.ProductID,
		UserID:           req.UserID,
		Username:         req.Username,
		Rating:           req.Rating,
		Title:            req.Title,
		Content:          req.Content,
		VerifiedPurchase: req.VerifiedPurchase,
		Images:           req.Images,
		HelpfulCount:     0,
		ReportCount:      0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	rs.reviews[reviewID] = review

	// Update product stats
	rs.updateProductStats(req.ProductID)

	return review, nil
}

// GetReview retrieves a review by ID
func (rs *ReviewService) GetReview(reviewID string) (*Review, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	review, exists := rs.reviews[reviewID]
	if !exists {
		return nil, errors.New("review not found")
	}

	return review, nil
}

// GetReviewsByProduct retrieves all reviews for a product with optional filters
func (rs *ReviewService) GetReviewsByProduct(productID string, minRating int, sortBy string) ([]*Review, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	var productReviews []*Review

	for _, review := range rs.reviews {
		if review.ProductID == productID {
			if minRating > 0 && review.Rating < minRating {
				continue
			}
			productReviews = append(productReviews, review)
		}
	}

	// Sort reviews
	rs.sortReviews(productReviews, sortBy)

	return productReviews, nil
}

// UpdateReview updates an existing review
func (rs *ReviewService) UpdateReview(reviewID string, req UpdateReviewRequest) (*Review, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	review, exists := rs.reviews[reviewID]
	if !exists {
		return nil, errors.New("review not found")
	}

	// Update fields if provided
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, errors.New("rating must be between 1 and 5")
		}
		review.Rating = *req.Rating
	}

	if req.Title != nil {
		review.Title = *req.Title
	}

	if req.Content != nil {
		review.Content = *req.Content
	}

	review.UpdatedAt = time.Now()

	// Update product stats if rating changed
	if req.Rating != nil {
		rs.updateProductStats(review.ProductID)
	}

	return review, nil
}

// DeleteReview deletes a review
func (rs *ReviewService) DeleteReview(reviewID string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	review, exists := rs.reviews[reviewID]
	if !exists {
		return errors.New("review not found")
	}

	productID := review.ProductID
	delete(rs.reviews, reviewID)

	// Update product stats
	rs.updateProductStats(productID)

	// Clean up reactions
	delete(rs.userReactions, reviewID)

	return nil
}

// AddReaction adds a helpful or report reaction to a review
func (rs *ReviewService) AddReaction(reviewID, userID, reactionType string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	review, exists := rs.reviews[reviewID]
	if !exists {
		return errors.New("review not found")
	}

	if reactionType != "helpful" && reactionType != "report" {
		return errors.New("invalid reaction type")
	}

	// Initialize reactions map for this review if needed
	if rs.userReactions[reviewID] == nil {
		rs.userReactions[reviewID] = make(map[string]string)
	}

	// Check if user already reacted
	previousReaction, hadReaction := rs.userReactions[reviewID][userID]

	// Remove previous reaction count
	if hadReaction {
		if previousReaction == "helpful" {
			review.HelpfulCount--
		} else if previousReaction == "report" {
			review.ReportCount--
		}
	}

	// Add new reaction
	rs.userReactions[reviewID][userID] = reactionType

	if reactionType == "helpful" {
		review.HelpfulCount++
	} else if reactionType == "report" {
		review.ReportCount++
	}

	return nil
}

// RemoveReaction removes a user's reaction from a review
func (rs *ReviewService) RemoveReaction(reviewID, userID string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	review, exists := rs.reviews[reviewID]
	if !exists {
		return errors.New("review not found")
	}

	if rs.userReactions[reviewID] == nil {
		return nil // No reactions to remove
	}

	reactionType, exists := rs.userReactions[reviewID][userID]
	if !exists {
		return nil // User hasn't reacted
	}

	// Remove reaction count
	if reactionType == "helpful" {
		review.HelpfulCount--
	} else if reactionType == "report" {
		review.ReportCount--
	}

	delete(rs.userReactions[reviewID], userID)

	return nil
}

// GetProductStats retrieves statistics for a product
func (rs *ReviewService) GetProductStats(productID string) (*ReviewStats, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	stats, exists := rs.productStats[productID]
	if !exists {
		// Return empty stats
		return &ReviewStats{
			ProductID:       productID,
			TotalReviews:    0,
			AverageRating:   0.0,
			RatingBreakdown: make(map[int]int),
			LastUpdated:     time.Now(),
		}, nil
	}

	return stats, nil
}

// updateProductStats recalculates statistics for a product (caller must hold lock)
func (rs *ReviewService) updateProductStats(productID string) {
	var totalReviews int
	var totalRating int
	ratingBreakdown := make(map[int]int)

	for _, review := range rs.reviews {
		if review.ProductID == productID {
			totalReviews++
			totalRating += review.Rating
			ratingBreakdown[review.Rating]++
		}
	}

	var averageRating float64
	if totalReviews > 0 {
		averageRating = float64(totalRating) / float64(totalReviews)
	}

	rs.productStats[productID] = &ReviewStats{
		ProductID:       productID,
		TotalReviews:    totalReviews,
		AverageRating:   averageRating,
		RatingBreakdown: ratingBreakdown,
		LastUpdated:     time.Now(),
	}
}

// sortReviews sorts reviews based on the specified criteria
func (rs *ReviewService) sortReviews(reviews []*Review, sortBy string) {
	switch sortBy {
	case "rating_desc":
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Rating > reviews[j].Rating
		})
	case "rating_asc":
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Rating < reviews[j].Rating
		})
	case "helpful":
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].HelpfulCount > reviews[j].HelpfulCount
		})
	case "recent":
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
		})
	case "oldest":
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].CreatedAt.Before(reviews[j].CreatedAt)
		})
	default:
		// Default to most recent
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
		})
	}
}

// validateCreateReviewRequest validates the create review request
func validateCreateReviewRequest(req CreateReviewRequest) error {
	if req.ProductID == "" {
		return errors.New("product_id is required")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Content == "" {
		return errors.New("content is required")
	}
	if len(req.Content) < 10 {
		return errors.New("content must be at least 10 characters")
	}
	if len(req.Content) > 5000 {
		return errors.New("content must be less than 5000 characters")
	}

	return nil
}
