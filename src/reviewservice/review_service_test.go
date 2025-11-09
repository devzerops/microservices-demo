package main

import (
	"testing"
	"time"
)

func TestCreateReview(t *testing.T) {
	service := NewReviewService()

	req := CreateReviewRequest{
		ProductID:        "PROD001",
		UserID:           "user123",
		Username:         "John Doe",
		Rating:           5,
		Title:            "Great product!",
		Content:          "This product exceeded my expectations. Highly recommended!",
		VerifiedPurchase: true,
	}

	review, err := service.CreateReview(req)

	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	if review.ProductID != req.ProductID {
		t.Errorf("Expected product ID %s, got %s", req.ProductID, review.ProductID)
	}

	if review.Rating != req.Rating {
		t.Errorf("Expected rating %d, got %d", req.Rating, review.Rating)
	}

	if review.HelpfulCount != 0 {
		t.Errorf("Expected initial helpful count 0, got %d", review.HelpfulCount)
	}
}

func TestCreateReview_Validation(t *testing.T) {
	service := NewReviewService()

	tests := []struct {
		name        string
		req         CreateReviewRequest
		expectError bool
	}{
		{
			name: "missing product ID",
			req: CreateReviewRequest{
				UserID:   "user123",
				Username: "John",
				Rating:   5,
				Title:    "Title",
				Content:  "Content here with enough characters",
			},
			expectError: true,
		},
		{
			name: "invalid rating (too low)",
			req: CreateReviewRequest{
				ProductID: "PROD001",
				UserID:    "user123",
				Username:  "John",
				Rating:    0,
				Title:     "Title",
				Content:   "Content here with enough characters",
			},
			expectError: true,
		},
		{
			name: "invalid rating (too high)",
			req: CreateReviewRequest{
				ProductID: "PROD001",
				UserID:    "user123",
				Username:  "John",
				Rating:    6,
				Title:     "Title",
				Content:   "Content here with enough characters",
			},
			expectError: true,
		},
		{
			name: "content too short",
			req: CreateReviewRequest{
				ProductID: "PROD001",
				UserID:    "user123",
				Username:  "John",
				Rating:    5,
				Title:     "Title",
				Content:   "Short",
			},
			expectError: true,
		},
		{
			name: "valid review",
			req: CreateReviewRequest{
				ProductID: "PROD001",
				UserID:    "user123",
				Username:  "John",
				Rating:    5,
				Title:     "Great!",
				Content:   "This is a valid review with enough content.",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateReview(tt.req)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetReview(t *testing.T) {
	service := NewReviewService()

	// Create a review
	req := CreateReviewRequest{
		ProductID: "PROD001",
		UserID:    "user123",
		Username:  "John",
		Rating:    4,
		Title:     "Good product",
		Content:   "Pretty good overall, would buy again.",
	}

	created, err := service.CreateReview(req)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Get the review
	retrieved, err := service.GetReview(created.ReviewID)
	if err != nil {
		t.Fatalf("Failed to get review: %v", err)
	}

	if retrieved.ReviewID != created.ReviewID {
		t.Errorf("Expected review ID %s, got %s", created.ReviewID, retrieved.ReviewID)
	}

	// Try to get non-existent review
	_, err = service.GetReview("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent review")
	}
}

func TestUpdateReview(t *testing.T) {
	service := NewReviewService()

	// Create a review
	req := CreateReviewRequest{
		ProductID: "PROD001",
		UserID:    "user123",
		Username:  "John",
		Rating:    3,
		Title:     "Okay product",
		Content:   "It's okay, nothing special.",
	}

	created, err := service.CreateReview(req)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Update the review
	newRating := 5
	newTitle := "Actually amazing!"
	updateReq := UpdateReviewRequest{
		Rating: &newRating,
		Title:  &newTitle,
	}

	updated, err := service.UpdateReview(created.ReviewID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update review: %v", err)
	}

	if updated.Rating != newRating {
		t.Errorf("Expected rating %d, got %d", newRating, updated.Rating)
	}

	if updated.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, updated.Title)
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("UpdatedAt should be after original CreatedAt")
	}
}

func TestDeleteReview(t *testing.T) {
	service := NewReviewService()

	// Create a review
	req := CreateReviewRequest{
		ProductID: "PROD001",
		UserID:    "user123",
		Username:  "John",
		Rating:    4,
		Title:     "Good",
		Content:   "This is a good product.",
	}

	created, err := service.CreateReview(req)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Delete the review
	err = service.DeleteReview(created.ReviewID)
	if err != nil {
		t.Fatalf("Failed to delete review: %v", err)
	}

	// Try to get deleted review
	_, err = service.GetReview(created.ReviewID)
	if err == nil {
		t.Error("Expected error when getting deleted review")
	}

	// Try to delete non-existent review
	err = service.DeleteReview("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent review")
	}
}

func TestAddReaction(t *testing.T) {
	service := NewReviewService()

	// Create a review
	req := CreateReviewRequest{
		ProductID: "PROD001",
		UserID:    "user123",
		Username:  "John",
		Rating:    5,
		Title:     "Excellent",
		Content:   "Excellent product!",
	}

	review, err := service.CreateReview(req)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Add helpful reaction
	err = service.AddReaction(review.ReviewID, "user456", "helpful")
	if err != nil {
		t.Fatalf("Failed to add helpful reaction: %v", err)
	}

	// Get review and check count
	updated, _ := service.GetReview(review.ReviewID)
	if updated.HelpfulCount != 1 {
		t.Errorf("Expected helpful count 1, got %d", updated.HelpfulCount)
	}

	// Add report reaction from another user
	err = service.AddReaction(review.ReviewID, "user789", "report")
	if err != nil {
		t.Fatalf("Failed to add report reaction: %v", err)
	}

	updated, _ = service.GetReview(review.ReviewID)
	if updated.ReportCount != 1 {
		t.Errorf("Expected report count 1, got %d", updated.ReportCount)
	}

	// Change reaction for same user
	err = service.AddReaction(review.ReviewID, "user456", "report")
	if err != nil {
		t.Fatalf("Failed to change reaction: %v", err)
	}

	updated, _ = service.GetReview(review.ReviewID)
	if updated.HelpfulCount != 0 {
		t.Errorf("Expected helpful count 0 after change, got %d", updated.HelpfulCount)
	}
	if updated.ReportCount != 2 {
		t.Errorf("Expected report count 2 after change, got %d", updated.ReportCount)
	}
}

func TestRemoveReaction(t *testing.T) {
	service := NewReviewService()

	// Create review and add reaction
	req := CreateReviewRequest{
		ProductID: "PROD001",
		UserID:    "user123",
		Username:  "John",
		Rating:    5,
		Title:     "Great",
		Content:   "Great product!",
	}

	review, _ := service.CreateReview(req)
	service.AddReaction(review.ReviewID, "user456", "helpful")

	// Remove reaction
	err := service.RemoveReaction(review.ReviewID, "user456")
	if err != nil {
		t.Fatalf("Failed to remove reaction: %v", err)
	}

	updated, _ := service.GetReview(review.ReviewID)
	if updated.HelpfulCount != 0 {
		t.Errorf("Expected helpful count 0 after removal, got %d", updated.HelpfulCount)
	}
}

func TestGetProductStats(t *testing.T) {
	service := NewReviewService()

	productID := "PROD001"

	// Create multiple reviews
	reviews := []CreateReviewRequest{
		{ProductID: productID, UserID: "user1", Username: "User1", Rating: 5, Title: "Perfect", Content: "Perfect product!"},
		{ProductID: productID, UserID: "user2", Username: "User2", Rating: 4, Title: "Good", Content: "Good product overall."},
		{ProductID: productID, UserID: "user3", Username: "User3", Rating: 5, Title: "Excellent", Content: "Excellent quality!"},
		{ProductID: productID, UserID: "user4", Username: "User4", Rating: 3, Title: "OK", Content: "It's okay, nothing special."},
	}

	for _, req := range reviews {
		_, err := service.CreateReview(req)
		if err != nil {
			t.Fatalf("Failed to create review: %v", err)
		}
	}

	// Get stats
	stats, err := service.GetProductStats(productID)
	if err != nil {
		t.Fatalf("Failed to get product stats: %v", err)
	}

	if stats.TotalReviews != 4 {
		t.Errorf("Expected 4 total reviews, got %d", stats.TotalReviews)
	}

	expectedAvg := (5.0 + 4.0 + 5.0 + 3.0) / 4.0
	if stats.AverageRating != expectedAvg {
		t.Errorf("Expected average rating %f, got %f", expectedAvg, stats.AverageRating)
	}

	if stats.RatingBreakdown[5] != 2 {
		t.Errorf("Expected 2 five-star reviews, got %d", stats.RatingBreakdown[5])
	}

	if stats.RatingBreakdown[4] != 1 {
		t.Errorf("Expected 1 four-star review, got %d", stats.RatingBreakdown[4])
	}

	if stats.RatingBreakdown[3] != 1 {
		t.Errorf("Expected 1 three-star review, got %d", stats.RatingBreakdown[3])
	}
}

func TestGetReviewsByProduct_Filtering(t *testing.T) {
	service := NewReviewService()

	productID := "PROD001"

	// Create reviews with different ratings
	ratings := []int{5, 4, 3, 2, 1}
	for i, rating := range ratings {
		req := CreateReviewRequest{
			ProductID: productID,
			UserID:    "user" + string(rune(i)),
			Username:  "User",
			Rating:    rating,
			Title:     "Review",
			Content:   "This is a review with enough content.",
		}
		service.CreateReview(req)
	}

	// Get reviews with minimum rating 4
	reviews, err := service.GetReviewsByProduct(productID, 4, "")
	if err != nil {
		t.Fatalf("Failed to get reviews: %v", err)
	}

	if len(reviews) != 2 {
		t.Errorf("Expected 2 reviews with rating >= 4, got %d", len(reviews))
	}

	for _, review := range reviews {
		if review.Rating < 4 {
			t.Errorf("Expected all reviews to have rating >= 4, got %d", review.Rating)
		}
	}
}

func TestGetReviewsByProduct_Sorting(t *testing.T) {
	service := NewReviewService()

	productID := "PROD001"

	// Create reviews with different ratings and times
	for i := 1; i <= 3; i++ {
		req := CreateReviewRequest{
			ProductID: productID,
			UserID:    "user" + string(rune(i)),
			Username:  "User",
			Rating:    i,
			Title:     "Review",
			Content:   "Review content here.",
		}
		review, _ := service.CreateReview(req)

		// Add helpful votes to first review
		if i == 1 {
			service.AddReaction(review.ReviewID, "voter1", "helpful")
			service.AddReaction(review.ReviewID, "voter2", "helpful")
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Test rating descending sort
	reviews, _ := service.GetReviewsByProduct(productID, 0, "rating_desc")
	if reviews[0].Rating < reviews[len(reviews)-1].Rating {
		t.Error("Reviews not sorted by rating descending")
	}

	// Test rating ascending sort
	reviews, _ = service.GetReviewsByProduct(productID, 0, "rating_asc")
	if reviews[0].Rating > reviews[len(reviews)-1].Rating {
		t.Error("Reviews not sorted by rating ascending")
	}

	// Test helpful sort
	reviews, _ = service.GetReviewsByProduct(productID, 0, "helpful")
	if reviews[0].HelpfulCount < reviews[len(reviews)-1].HelpfulCount {
		t.Error("Reviews not sorted by helpful count")
	}
}
