package main

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Product represents a searchable product
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Rating      float64   `json:"rating"`
	ReviewCount int       `json:"review_count"`
	InStock     bool      `json:"in_stock"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Popularity  int       `json:"popularity"` // View/purchase count
}

// SearchFilters represents search filter options
type SearchFilters struct {
	Categories  []string  `json:"categories,omitempty"`
	MinPrice    *float64  `json:"min_price,omitempty"`
	MaxPrice    *float64  `json:"max_price,omitempty"`
	MinRating   *float64  `json:"min_rating,omitempty"`
	InStockOnly bool      `json:"in_stock_only"`
	SortBy      string    `json:"sort_by"` // relevance, price_asc, price_desc, rating, popularity, newest
	Page        int       `json:"page"`
	PageSize    int       `json:"page_size"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Product   Product `json:"product"`
	Score     float64 `json:"score"`
	Relevance float64 `json:"relevance"`
	MatchType string  `json:"match_type"` // exact, partial, fuzzy
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query       string         `json:"query"`
	Results     []SearchResult `json:"results"`
	Total       int            `json:"total"`
	Page        int            `json:"page"`
	PageSize    int            `json:"page_size"`
	TotalPages  int            `json:"total_pages"`
	Filters     SearchFilters  `json:"filters"`
	Took        int64          `json:"took_ms"`
	Suggestions []string       `json:"suggestions,omitempty"`
}

// ProductIndex manages the searchable product database
type ProductIndex struct {
	products map[string]*Product // ID -> Product
	mu       sync.RWMutex
}

// NewProductIndex creates a new product index
func NewProductIndex() *ProductIndex {
	pi := &ProductIndex{
		products: make(map[string]*Product),
	}

	// Initialize with demo products
	pi.initializeDemoProducts()

	return pi
}

func (pi *ProductIndex) initializeDemoProducts() {
	demoProducts := []Product{
		{
			ID:          "OLJCESPC7Z",
			Name:        "Sunglasses",
			Description: "Vintage sunglasses with UV protection",
			Category:    "accessories",
			Price:       19.99,
			Rating:      4.5,
			ReviewCount: 156,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -3, 0),
			Popularity:  250,
		},
		{
			ID:          "66VCHSJNUP",
			Name:        "Tank Top",
			Description: "Comfortable cotton tank top",
			Category:    "clothing",
			Price:       18.99,
			Rating:      4.2,
			ReviewCount: 89,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
			Popularity:  180,
		},
		{
			ID:          "1YMWWN1N4O",
			Name:        "Watch",
			Description: "Stainless steel watch",
			Category:    "accessories",
			Price:       50.00,
			Rating:      4.8,
			ReviewCount: 342,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -5, 0),
			Popularity:  420,
		},
		{
			ID:          "L9ECAV7KIM",
			Name:        "Loafers",
			Description: "Leather loafers",
			Category:    "shoes",
			Price:       89.00,
			Rating:      4.6,
			ReviewCount: 213,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -4, 0),
			Popularity:  310,
		},
		{
			ID:          "2ZYFJ3GM2N",
			Name:        "Hairdryer",
			Description: "Professional hairdryer",
			Category:    "home",
			Price:       28.99,
			Rating:      4.3,
			ReviewCount: 127,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -1, 0),
			Popularity:  95,
		},
		{
			ID:          "0PUK6V6EV0",
			Name:        "Candle Holder",
			Description: "Vintage candle holder",
			Category:    "home",
			Price:       18.99,
			Rating:      4.1,
			ReviewCount: 64,
			InStock:     false,
			CreatedAt:   time.Now().AddDate(0, -6, 0),
			Popularity:  78,
		},
		{
			ID:          "LS4PSXUNUM",
			Name:        "Salt & Pepper Shakers",
			Description: "Ceramic shakers set",
			Category:    "home",
			Price:       18.49,
			Rating:      4.4,
			ReviewCount: 142,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -2, -15),
			Popularity:  156,
		},
		{
			ID:          "9SIQT8TOJO",
			Name:        "Bamboo Glass Jar",
			Description: "Eco-friendly storage jar",
			Category:    "home",
			Price:       5.00,
			Rating:      4.0,
			ReviewCount: 98,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, -1, -10),
			Popularity:  134,
		},
		{
			ID:          "6E92ZMYYFZ",
			Name:        "Mug",
			Description: "Ceramic coffee mug",
			Category:    "home",
			Price:       8.99,
			Rating:      4.7,
			ReviewCount: 231,
			InStock:     true,
			CreatedAt:   time.Now().AddDate(0, 0, -20),
			Popularity:  289,
		},
	}

	for i := range demoProducts {
		pi.products[demoProducts[i].ID] = &demoProducts[i]
	}
}

// AddProduct adds or updates a product in the index
func (pi *ProductIndex) AddProduct(product Product) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.products[product.ID] = &product
}

// Search performs a comprehensive search with filters and sorting
func (pi *ProductIndex) Search(query string, filters SearchFilters) []SearchResult {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	var results []SearchResult

	// Search through all products
	for _, product := range pi.products {
		// Apply filters first
		if !pi.matchesFilters(product, filters) {
			continue
		}

		// Calculate relevance score
		score, matchType := pi.calculateRelevance(product, query)
		if score > 0 {
			results = append(results, SearchResult{
				Product:   *product,
				Score:     score,
				Relevance: score,
				MatchType: matchType,
			})
		}
	}

	// Sort results
	pi.sortResults(results, filters.SortBy)

	return results
}

// matchesFilters checks if a product matches the given filters
func (pi *ProductIndex) matchesFilters(product *Product, filters SearchFilters) bool {
	// Category filter
	if len(filters.Categories) > 0 {
		matched := false
		for _, cat := range filters.Categories {
			if strings.EqualFold(product.Category, cat) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Price range filter
	if filters.MinPrice != nil && product.Price < *filters.MinPrice {
		return false
	}
	if filters.MaxPrice != nil && product.Price > *filters.MaxPrice {
		return false
	}

	// Rating filter
	if filters.MinRating != nil && product.Rating < *filters.MinRating {
		return false
	}

	// Stock filter
	if filters.InStockOnly && !product.InStock {
		return false
	}

	return true
}

// calculateRelevance calculates how relevant a product is to the query
func (pi *ProductIndex) calculateRelevance(product *Product, query string) (float64, string) {
	if query == "" {
		return 50.0, "all" // Return all products with base score
	}

	productName := strings.ToLower(product.Name)
	productDesc := strings.ToLower(product.Description)
	productCategory := strings.ToLower(product.Category)

	// Exact name match
	if productName == query {
		return 100.0, "exact"
	}

	// Starts with query
	if strings.HasPrefix(productName, query) {
		return 90.0, "prefix"
	}

	// Contains query in name
	if strings.Contains(productName, query) {
		return 80.0, "partial"
	}

	// Category match
	if productCategory == query {
		return 70.0, "category"
	}

	// Contains in description
	if strings.Contains(productDesc, query) {
		return 60.0, "description"
	}

	// Word match (any word in name matches query)
	nameWords := strings.Fields(productName)
	for _, word := range nameWords {
		if strings.HasPrefix(word, query) {
			return 50.0, "word"
		}
	}

	// Fuzzy match using Levenshtein distance
	distance := levenshteinDistance(query, productName)
	maxLen := len(query)
	if len(productName) > maxLen {
		maxLen = len(productName)
	}

	similarity := 1.0 - float64(distance)/float64(maxLen)
	if similarity > 0.6 {
		return similarity * 40.0, "fuzzy"
	}

	return 0, ""
}

// sortResults sorts search results based on the specified criteria
func (pi *ProductIndex) sortResults(results []SearchResult, sortBy string) {
	switch sortBy {
	case "price_asc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Product.Price < results[j].Product.Price
		})
	case "price_desc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Product.Price > results[j].Product.Price
		})
	case "rating":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Product.Rating > results[j].Product.Rating
		})
	case "popularity":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Product.Popularity > results[j].Product.Popularity
		})
	case "newest":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Product.CreatedAt.After(results[j].Product.CreatedAt)
		})
	case "relevance", "":
		// Default: sort by relevance score (already calculated)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
}

func min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}
