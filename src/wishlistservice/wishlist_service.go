package main

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WishlistService handles all wishlist-related operations
type WishlistService struct {
	wishlists      map[string]*Wishlist    // userID -> Wishlist
	items          map[string]*WishlistItem // itemID -> WishlistItem
	alerts         map[string][]*PriceAlert // userID -> alerts
	priceHistory   map[string][]PriceUpdate // productID -> price history
	mu             sync.RWMutex
}

// NewWishlistService creates a new wishlist service instance
func NewWishlistService() *WishlistService {
	return &WishlistService{
		wishlists:    make(map[string]*Wishlist),
		items:        make(map[string]*WishlistItem),
		alerts:       make(map[string][]*PriceAlert),
		priceHistory: make(map[string][]PriceUpdate),
	}
}

// AddItem adds an item to a user's wishlist
func (ws *WishlistService) AddItem(userID string, req AddItemRequest) (*WishlistItem, error) {
	if err := validateAddItemRequest(req); err != nil {
		return nil, err
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check if product already in wishlist
	if wishlist, exists := ws.wishlists[userID]; exists {
		for _, item := range wishlist.Items {
			if item.ProductID == req.ProductID {
				return nil, errors.New("product already in wishlist")
			}
		}
	}

	// Create wishlist item
	itemID := uuid.New().String()
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	item := &WishlistItem{
		ItemID:            itemID,
		UserID:            userID,
		ProductID:         req.ProductID,
		ProductName:       req.ProductName,
		ProductImageURL:   req.ProductImageURL,
		CurrentPrice:      req.CurrentPrice,
		OriginalPrice:     req.CurrentPrice,
		TargetPrice:       req.TargetPrice,
		Priority:          priority,
		Notes:             req.Notes,
		NotifyOnPriceDrop: req.NotifyOnPriceDrop,
		NotifyOnRestock:   req.NotifyOnRestock,
		InStock:           req.InStock,
		AddedAt:           time.Now(),
		UpdatedAt:         time.Now(),
	}

	ws.items[itemID] = item

	// Initialize or update wishlist
	if _, exists := ws.wishlists[userID]; !exists {
		ws.wishlists[userID] = &Wishlist{
			UserID:     userID,
			Items:      []*WishlistItem{},
			TotalItems: 0,
			SharedWith: []string{},
			IsPublic:   false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}

	ws.wishlists[userID].Items = append(ws.wishlists[userID].Items, item)
	ws.wishlists[userID].TotalItems++
	ws.wishlists[userID].UpdatedAt = time.Now()

	return item, nil
}

// RemoveItem removes an item from a user's wishlist
func (ws *WishlistService) RemoveItem(userID, itemID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	item, exists := ws.items[itemID]
	if !exists {
		return errors.New("item not found")
	}

	if item.UserID != userID {
		return errors.New("unauthorized: item does not belong to user")
	}

	// Remove from items map
	delete(ws.items, itemID)

	// Remove from wishlist
	if wishlist, exists := ws.wishlists[userID]; exists {
		for i, wItem := range wishlist.Items {
			if wItem.ItemID == itemID {
				wishlist.Items = append(wishlist.Items[:i], wishlist.Items[i+1:]...)
				wishlist.TotalItems--
				wishlist.UpdatedAt = time.Now()
				break
			}
		}
	}

	return nil
}

// GetWishlist retrieves a user's complete wishlist
func (ws *WishlistService) GetWishlist(userID string, sortBy string) (*Wishlist, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	wishlist, exists := ws.wishlists[userID]
	if !exists {
		// Return empty wishlist
		return &Wishlist{
			UserID:     userID,
			Items:      []*WishlistItem{},
			TotalItems: 0,
			SharedWith: []string{},
			IsPublic:   false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}, nil
	}

	// Sort items if requested
	if sortBy != "" {
		ws.sortItems(wishlist.Items, sortBy)
	}

	return wishlist, nil
}

// GetItem retrieves a specific wishlist item
func (ws *WishlistService) GetItem(userID, itemID string) (*WishlistItem, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	item, exists := ws.items[itemID]
	if !exists {
		return nil, errors.New("item not found")
	}

	if item.UserID != userID {
		return nil, errors.New("unauthorized: item does not belong to user")
	}

	return item, nil
}

// UpdateItem updates a wishlist item
func (ws *WishlistService) UpdateItem(userID, itemID string, req UpdateItemRequest) (*WishlistItem, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	item, exists := ws.items[itemID]
	if !exists {
		return nil, errors.New("item not found")
	}

	if item.UserID != userID {
		return nil, errors.New("unauthorized: item does not belong to user")
	}

	// Update fields if provided
	if req.TargetPrice != nil {
		item.TargetPrice = *req.TargetPrice
	}

	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, errors.New("invalid priority: must be high, medium, or low")
		}
		item.Priority = *req.Priority
	}

	if req.Notes != nil {
		item.Notes = *req.Notes
	}

	if req.NotifyOnPriceDrop != nil {
		item.NotifyOnPriceDrop = *req.NotifyOnPriceDrop
	}

	if req.NotifyOnRestock != nil {
		item.NotifyOnRestock = *req.NotifyOnRestock
	}

	item.UpdatedAt = time.Now()

	// Update wishlist timestamp
	if wishlist, exists := ws.wishlists[userID]; exists {
		wishlist.UpdatedAt = time.Now()
	}

	return item, nil
}

// UpdateProductPrice updates the price of a product and generates alerts
func (ws *WishlistService) UpdateProductPrice(productID string, newPrice float64, inStock bool) ([]*PriceAlert, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	var alerts []*PriceAlert

	// Find all items with this product
	for _, item := range ws.items {
		if item.ProductID != productID {
			continue
		}

		oldPrice := item.CurrentPrice
		priceChange := newPrice - oldPrice
		percentChange := 0.0
		if oldPrice > 0 {
			percentChange = (priceChange / oldPrice) * 100
		}

		// Update price
		item.CurrentPrice = newPrice
		item.InStock = inStock
		item.UpdatedAt = time.Now()

		// Record price history
		priceUpdate := PriceUpdate{
			ProductID:     productID,
			OldPrice:      oldPrice,
			NewPrice:      newPrice,
			PriceChange:   priceChange,
			PercentChange: percentChange,
			UpdatedAt:     time.Now(),
		}

		if ws.priceHistory[productID] == nil {
			ws.priceHistory[productID] = []PriceUpdate{}
		}
		ws.priceHistory[productID] = append(ws.priceHistory[productID], priceUpdate)

		// Generate alerts
		shouldAlert := false
		alertType := ""

		// Check price drop alert
		if item.NotifyOnPriceDrop && priceChange < 0 {
			shouldAlert = true
			alertType = "price_drop"
		}

		// Check target price reached
		if item.TargetPrice > 0 && newPrice <= item.TargetPrice {
			shouldAlert = true
			alertType = "target_reached"
		}

		// Check restock alert
		if item.NotifyOnRestock && inStock && !item.InStock {
			shouldAlert = true
			alertType = "restock"
		}

		if shouldAlert {
			alert := &PriceAlert{
				AlertID:     uuid.New().String(),
				UserID:      item.UserID,
				ItemID:      item.ItemID,
				ProductID:   productID,
				ProductName: item.ProductName,
				PriceUpdate: priceUpdate,
				TargetPrice: item.TargetPrice,
				AlertType:   alertType,
				IsRead:      false,
				CreatedAt:   time.Now(),
			}

			alerts = append(alerts, alert)

			// Store alert
			if ws.alerts[item.UserID] == nil {
				ws.alerts[item.UserID] = []*PriceAlert{}
			}
			ws.alerts[item.UserID] = append(ws.alerts[item.UserID], alert)
		}
	}

	return alerts, nil
}

// GetAlerts retrieves alerts for a user
func (ws *WishlistService) GetAlerts(userID string, unreadOnly bool) ([]*PriceAlert, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	userAlerts, exists := ws.alerts[userID]
	if !exists {
		return []*PriceAlert{}, nil
	}

	if !unreadOnly {
		return userAlerts, nil
	}

	// Filter unread only
	var unread []*PriceAlert
	for _, alert := range userAlerts {
		if !alert.IsRead {
			unread = append(unread, alert)
		}
	}

	return unread, nil
}

// MarkAlertAsRead marks an alert as read
func (ws *WishlistService) MarkAlertAsRead(userID, alertID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	userAlerts, exists := ws.alerts[userID]
	if !exists {
		return errors.New("no alerts found for user")
	}

	for _, alert := range userAlerts {
		if alert.AlertID == alertID {
			alert.IsRead = true
			return nil
		}
	}

	return errors.New("alert not found")
}

// ShareWishlist shares a wishlist with another user
func (ws *WishlistService) ShareWishlist(userID, shareWithUserID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wishlist, exists := ws.wishlists[userID]
	if !exists {
		return errors.New("wishlist not found")
	}

	// Check if already shared
	for _, sharedUserID := range wishlist.SharedWith {
		if sharedUserID == shareWithUserID {
			return errors.New("wishlist already shared with this user")
		}
	}

	wishlist.SharedWith = append(wishlist.SharedWith, shareWithUserID)
	wishlist.UpdatedAt = time.Now()

	return nil
}

// UnshareWishlist removes sharing access
func (ws *WishlistService) UnshareWishlist(userID, unshareUserID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wishlist, exists := ws.wishlists[userID]
	if !exists {
		return errors.New("wishlist not found")
	}

	for i, sharedUserID := range wishlist.SharedWith {
		if sharedUserID == unshareUserID {
			wishlist.SharedWith = append(wishlist.SharedWith[:i], wishlist.SharedWith[i+1:]...)
			wishlist.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("wishlist not shared with this user")
}

// SetPublic sets the wishlist public/private status
func (ws *WishlistService) SetPublic(userID string, isPublic bool) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wishlist, exists := ws.wishlists[userID]
	if !exists {
		return errors.New("wishlist not found")
	}

	wishlist.IsPublic = isPublic
	wishlist.UpdatedAt = time.Now()

	return nil
}

// GetStats retrieves statistics about a user's wishlist
func (ws *WishlistService) GetStats(userID string) (*WishlistStats, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	wishlist, exists := ws.wishlists[userID]
	if !exists {
		return &WishlistStats{UserID: userID}, nil
	}

	stats := &WishlistStats{
		UserID: userID,
	}

	var totalValue float64
	var totalPriceDrop float64
	var itemsWithPriceDrop int

	for _, item := range wishlist.Items {
		stats.TotalItems++
		totalValue += item.CurrentPrice

		// Count items on sale
		if item.CurrentPrice < item.OriginalPrice {
			stats.ItemsOnSale++
			priceDrop := item.OriginalPrice - item.CurrentPrice
			totalPriceDrop += priceDrop
			itemsWithPriceDrop++
		}

		// Count out of stock
		if !item.InStock {
			stats.OutOfStockItems++
		}

		// Count by priority
		switch item.Priority {
		case "high":
			stats.HighPriorityItems++
		case "medium":
			stats.MediumPriorityItems++
		case "low":
			stats.LowPriorityItems++
		}
	}

	stats.TotalValue = totalValue
	if itemsWithPriceDrop > 0 {
		stats.AveragePriceDrop = totalPriceDrop / float64(itemsWithPriceDrop)
	}

	return stats, nil
}

// Helper functions

func (ws *WishlistService) sortItems(items []*WishlistItem, sortBy string) {
	switch sortBy {
	case "price_asc":
		sort.Slice(items, func(i, j int) bool {
			return items[i].CurrentPrice < items[j].CurrentPrice
		})
	case "price_desc":
		sort.Slice(items, func(i, j int) bool {
			return items[i].CurrentPrice > items[j].CurrentPrice
		})
	case "priority":
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		sort.Slice(items, func(i, j int) bool {
			return priorityOrder[items[i].Priority] < priorityOrder[items[j].Priority]
		})
	case "discount":
		sort.Slice(items, func(i, j int) bool {
			discountI := items[i].OriginalPrice - items[i].CurrentPrice
			discountJ := items[j].OriginalPrice - items[j].CurrentPrice
			return discountI > discountJ
		})
	case "recent":
		sort.Slice(items, func(i, j int) bool {
			return items[i].AddedAt.After(items[j].AddedAt)
		})
	case "oldest":
		sort.Slice(items, func(i, j int) bool {
			return items[i].AddedAt.Before(items[j].AddedAt)
		})
	}
}

func validateAddItemRequest(req AddItemRequest) error {
	if req.ProductID == "" {
		return errors.New("product_id is required")
	}
	if req.CurrentPrice < 0 {
		return errors.New("current_price must be non-negative")
	}
	if req.TargetPrice < 0 {
		return errors.New("target_price must be non-negative")
	}
	if req.Priority != "" && !isValidPriority(req.Priority) {
		return errors.New("priority must be high, medium, or low")
	}
	return nil
}

func isValidPriority(priority string) bool {
	return priority == "high" || priority == "medium" || priority == "low"
}
