package main

import (
	"testing"
	"time"
)

func TestAddItem(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	req := AddItemRequest{
		ProductID:         "PROD001",
		ProductName:       "Cool Sunglasses",
		CurrentPrice:      99.99,
		Priority:          "high",
		NotifyOnPriceDrop: true,
		InStock:           true,
	}

	item, err := service.AddItem(userID, req)

	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	if item.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, item.UserID)
	}

	if item.ProductID != req.ProductID {
		t.Errorf("Expected product ID %s, got %s", req.ProductID, item.ProductID)
	}

	if item.CurrentPrice != req.CurrentPrice {
		t.Errorf("Expected price %.2f, got %.2f", req.CurrentPrice, item.CurrentPrice)
	}

	if item.OriginalPrice != req.CurrentPrice {
		t.Errorf("Original price should equal current price on add")
	}
}

func TestAddItem_Duplicate(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		InStock:      true,
	}

	// Add first time
	_, err := service.AddItem(userID, req)
	if err != nil {
		t.Fatalf("Failed to add item first time: %v", err)
	}

	// Try to add same product again
	_, err = service.AddItem(userID, req)
	if err == nil {
		t.Error("Expected error when adding duplicate product")
	}
}

func TestRemoveItem(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		InStock:      true,
	}

	item, _ := service.AddItem(userID, req)

	// Remove the item
	err := service.RemoveItem(userID, item.ItemID)
	if err != nil {
		t.Fatalf("Failed to remove item: %v", err)
	}

	// Try to get removed item
	_, err = service.GetItem(userID, item.ItemID)
	if err == nil {
		t.Error("Expected error when getting removed item")
	}
}

func TestGetWishlist(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	// Add multiple items
	for i := 1; i <= 3; i++ {
		req := AddItemRequest{
			ProductID:    "PROD00" + string(rune(i+'0')),
			CurrentPrice: float64(i * 10),
			InStock:      true,
		}
		service.AddItem(userID, req)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	wishlist, err := service.GetWishlist(userID, "")
	if err != nil {
		t.Fatalf("Failed to get wishlist: %v", err)
	}

	if wishlist.TotalItems != 3 {
		t.Errorf("Expected 3 items, got %d", wishlist.TotalItems)
	}

	if len(wishlist.Items) != 3 {
		t.Errorf("Expected 3 items in array, got %d", len(wishlist.Items))
	}
}

func TestUpdateItem(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		Priority:     "low",
		InStock:      true,
	}

	item, _ := service.AddItem(userID, req)

	// Update priority and target price
	newPriority := "high"
	targetPrice := 79.99
	notes := "Wait for sale"
	updateReq := UpdateItemRequest{
		Priority:    &newPriority,
		TargetPrice: &targetPrice,
		Notes:       &notes,
	}

	updated, err := service.UpdateItem(userID, item.ItemID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}

	if updated.Priority != newPriority {
		t.Errorf("Expected priority %s, got %s", newPriority, updated.Priority)
	}

	if updated.TargetPrice != targetPrice {
		t.Errorf("Expected target price %.2f, got %.2f", targetPrice, updated.TargetPrice)
	}

	if updated.Notes != notes {
		t.Errorf("Expected notes %s, got %s", notes, updated.Notes)
	}
}

func TestUpdateProductPrice_PriceDropAlert(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	productID := "PROD001"
	req := AddItemRequest{
		ProductID:         productID,
		CurrentPrice:      100.00,
		NotifyOnPriceDrop: true,
		InStock:           true,
	}

	service.AddItem(userID, req)

	// Update to lower price
	alerts, err := service.UpdateProductPrice(productID, 80.00, true)
	if err != nil {
		t.Fatalf("Failed to update price: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.AlertType != "price_drop" {
		t.Errorf("Expected alert type price_drop, got %s", alert.AlertType)
	}

	if alert.PriceUpdate.NewPrice != 80.00 {
		t.Errorf("Expected new price 80.00, got %.2f", alert.PriceUpdate.NewPrice)
	}

	if alert.PriceUpdate.PriceChange >= 0 {
		t.Error("Expected negative price change for price drop")
	}
}

func TestUpdateProductPrice_TargetPriceReached(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	productID := "PROD001"
	req := AddItemRequest{
		ProductID:         productID,
		CurrentPrice:      100.00,
		TargetPrice:       75.00,
		NotifyOnPriceDrop: true,
		InStock:           true,
	}

	service.AddItem(userID, req)

	// Update to target price
	alerts, err := service.UpdateProductPrice(productID, 70.00, true)
	if err != nil {
		t.Fatalf("Failed to update price: %v", err)
	}

	if len(alerts) == 0 {
		t.Fatal("Expected at least one alert")
	}

	// Should have target_reached alert
	hasTargetAlert := false
	for _, alert := range alerts {
		if alert.AlertType == "target_reached" {
			hasTargetAlert = true
			break
		}
	}

	if !hasTargetAlert {
		t.Error("Expected target_reached alert")
	}
}

func TestUpdateProductPrice_RestockAlert(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	productID := "PROD001"
	req := AddItemRequest{
		ProductID:       productID,
		CurrentPrice:    100.00,
		NotifyOnRestock: true,
		InStock:         false, // Out of stock initially
	}

	service.AddItem(userID, req)

	// Update to in stock
	alerts, err := service.UpdateProductPrice(productID, 100.00, true)
	if err != nil {
		t.Fatalf("Failed to update price: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].AlertType != "restock" {
		t.Errorf("Expected restock alert, got %s", alerts[0].AlertType)
	}
}

func TestGetAlerts(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"
	productID := "PROD001"

	req := AddItemRequest{
		ProductID:         productID,
		CurrentPrice:      100.00,
		NotifyOnPriceDrop: true,
		InStock:           true,
	}

	service.AddItem(userID, req)

	// Generate price drop alert
	service.UpdateProductPrice(productID, 80.00, true)

	// Get all alerts
	alerts, err := service.GetAlerts(userID, false)
	if err != nil {
		t.Fatalf("Failed to get alerts: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected at least one alert")
	}

	// All should be unread initially
	unreadAlerts, _ := service.GetAlerts(userID, true)
	if len(unreadAlerts) != len(alerts) {
		t.Error("All alerts should be unread initially")
	}
}

func TestMarkAlertAsRead(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"
	productID := "PROD001"

	req := AddItemRequest{
		ProductID:         productID,
		CurrentPrice:      100.00,
		NotifyOnPriceDrop: true,
		InStock:           true,
	}

	service.AddItem(userID, req)
	service.UpdateProductPrice(productID, 80.00, true)

	alerts, _ := service.GetAlerts(userID, false)
	if len(alerts) == 0 {
		t.Fatal("No alerts to test")
	}

	alertID := alerts[0].AlertID

	// Mark as read
	err := service.MarkAlertAsRead(userID, alertID)
	if err != nil {
		t.Fatalf("Failed to mark alert as read: %v", err)
	}

	// Check unread count
	unreadAlerts, _ := service.GetAlerts(userID, true)
	if len(unreadAlerts) != 0 {
		t.Error("Expected no unread alerts after marking as read")
	}
}

func TestShareWishlist(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"
	shareWithUserID := "user456"

	// Add an item first
	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		InStock:      true,
	}
	service.AddItem(userID, req)

	// Share wishlist
	err := service.ShareWishlist(userID, shareWithUserID)
	if err != nil {
		t.Fatalf("Failed to share wishlist: %v", err)
	}

	// Check shared status
	wishlist, _ := service.GetWishlist(userID, "")
	if len(wishlist.SharedWith) != 1 {
		t.Errorf("Expected 1 shared user, got %d", len(wishlist.SharedWith))
	}

	if wishlist.SharedWith[0] != shareWithUserID {
		t.Errorf("Expected shared with %s, got %s", shareWithUserID, wishlist.SharedWith[0])
	}
}

func TestUnshareWishlist(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"
	shareWithUserID := "user456"

	// Add item and share
	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		InStock:      true,
	}
	service.AddItem(userID, req)
	service.ShareWishlist(userID, shareWithUserID)

	// Unshare
	err := service.UnshareWishlist(userID, shareWithUserID)
	if err != nil {
		t.Fatalf("Failed to unshare wishlist: %v", err)
	}

	// Check unshared
	wishlist, _ := service.GetWishlist(userID, "")
	if len(wishlist.SharedWith) != 0 {
		t.Error("Expected no shared users after unsharing")
	}
}

func TestSetPublic(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	// Add item first
	req := AddItemRequest{
		ProductID:    "PROD001",
		CurrentPrice: 99.99,
		InStock:      true,
	}
	service.AddItem(userID, req)

	// Set public
	err := service.SetPublic(userID, true)
	if err != nil {
		t.Fatalf("Failed to set public: %v", err)
	}

	wishlist, _ := service.GetWishlist(userID, "")
	if !wishlist.IsPublic {
		t.Error("Expected wishlist to be public")
	}

	// Set private
	service.SetPublic(userID, false)
	wishlist, _ = service.GetWishlist(userID, "")
	if wishlist.IsPublic {
		t.Error("Expected wishlist to be private")
	}
}

func TestGetStats(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	// Add items with different priorities and prices
	items := []AddItemRequest{
		{ProductID: "P1", CurrentPrice: 100.00, Priority: "high", InStock: true},
		{ProductID: "P2", CurrentPrice: 50.00, Priority: "medium", InStock: true},
		{ProductID: "P3", CurrentPrice: 75.00, Priority: "low", InStock: false},
	}

	for _, item := range items {
		service.AddItem(userID, item)
	}

	// Update one price to create a sale
	service.UpdateProductPrice("P1", 80.00, true)

	stats, err := service.GetStats(userID)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", stats.TotalItems)
	}

	if stats.HighPriorityItems != 1 {
		t.Errorf("Expected 1 high priority item, got %d", stats.HighPriorityItems)
	}

	if stats.MediumPriorityItems != 1 {
		t.Errorf("Expected 1 medium priority item, got %d", stats.MediumPriorityItems)
	}

	if stats.LowPriorityItems != 1 {
		t.Errorf("Expected 1 low priority item, got %d", stats.LowPriorityItems)
	}

	if stats.OutOfStockItems != 1 {
		t.Errorf("Expected 1 out of stock item, got %d", stats.OutOfStockItems)
	}

	if stats.ItemsOnSale != 1 {
		t.Errorf("Expected 1 item on sale, got %d", stats.ItemsOnSale)
	}
}

func TestSorting(t *testing.T) {
	service := NewWishlistService()
	userID := "user123"

	// Add items with different prices and priorities
	items := []AddItemRequest{
		{ProductID: "P1", CurrentPrice: 100.00, Priority: "low", InStock: true},
		{ProductID: "P2", CurrentPrice: 50.00, Priority: "high", InStock: true},
		{ProductID: "P3", CurrentPrice: 75.00, Priority: "medium", InStock: true},
	}

	for _, item := range items {
		service.AddItem(userID, item)
		time.Sleep(10 * time.Millisecond)
	}

	// Test price ascending sort
	wishlist, _ := service.GetWishlist(userID, "price_asc")
	if wishlist.Items[0].CurrentPrice > wishlist.Items[1].CurrentPrice {
		t.Error("Items not sorted by price ascending")
	}

	// Test priority sort
	wishlist, _ = service.GetWishlist(userID, "priority")
	if wishlist.Items[0].Priority != "high" {
		t.Error("Items not sorted by priority")
	}
}
