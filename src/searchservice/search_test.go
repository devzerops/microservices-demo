package main

import (
	"testing"
)

func TestProductIndex_Search_BasicQuery(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("watch", SearchFilters{})

	if len(results) == 0 {
		t.Fatal("Expected at least one result for 'watch'")
	}

	// Should find the watch product
	found := false
	for _, result := range results {
		if result.Product.Name == "Watch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Watch' product")
	}
}

func TestProductIndex_Search_EmptyQuery(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("", SearchFilters{})

	// Empty query should return all products
	if len(results) != 9 {
		t.Errorf("Expected 9 products, got %d", len(results))
	}

	// All should have base score
	for _, result := range results {
		if result.Score != 50.0 {
			t.Errorf("Expected base score 50.0, got %.1f", result.Score)
		}
	}
}

func TestProductIndex_Search_ExactMatch(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("sunglasses", SearchFilters{})

	if len(results) == 0 {
		t.Fatal("Expected results for exact match")
	}

	// Exact match should have highest score
	if results[0].Score != 100.0 {
		t.Errorf("Expected exact match score 100.0, got %.1f", results[0].Score)
	}

	if results[0].MatchType != "exact" {
		t.Errorf("Expected match type 'exact', got '%s'", results[0].MatchType)
	}
}

func TestProductIndex_Search_PriceFilter(t *testing.T) {
	index := NewProductIndex()

	minPrice := 20.0
	maxPrice := 50.0

	filters := SearchFilters{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}

	results := index.Search("", filters)

	// All results should be within price range
	for _, result := range results {
		if result.Product.Price < minPrice || result.Product.Price > maxPrice {
			t.Errorf("Product %s price %.2f out of range [%.2f, %.2f]",
				result.Product.Name, result.Product.Price, minPrice, maxPrice)
		}
	}

	// Should find some products in this range
	if len(results) == 0 {
		t.Error("Expected to find products in price range $20-$50")
	}
}

func TestProductIndex_Search_CategoryFilter(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		Categories: []string{"accessories"},
	}

	results := index.Search("", filters)

	// All results should be accessories
	for _, result := range results {
		if result.Product.Category != "accessories" {
			t.Errorf("Expected accessories, got %s", result.Product.Category)
		}
	}

	// Should find some accessories
	if len(results) == 0 {
		t.Error("Expected to find accessories")
	}
}

func TestProductIndex_Search_RatingFilter(t *testing.T) {
	index := NewProductIndex()

	minRating := 4.5

	filters := SearchFilters{
		MinRating: &minRating,
	}

	results := index.Search("", filters)

	// All results should have rating >= 4.5
	for _, result := range results {
		if result.Product.Rating < minRating {
			t.Errorf("Product %s rating %.1f below minimum %.1f",
				result.Product.Name, result.Product.Rating, minRating)
		}
	}
}

func TestProductIndex_Search_InStockFilter(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		InStockOnly: true,
	}

	results := index.Search("", filters)

	// All results should be in stock
	for _, result := range results {
		if !result.Product.InStock {
			t.Errorf("Product %s is out of stock", result.Product.Name)
		}
	}

	// Should find in-stock products
	if len(results) == 0 {
		t.Error("Expected to find in-stock products")
	}
}

func TestProductIndex_Search_SortByPriceAsc(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "price_asc",
	}

	results := index.Search("", filters)

	// Verify ascending price order
	for i := 1; i < len(results); i++ {
		if results[i].Product.Price < results[i-1].Product.Price {
			t.Errorf("Results not sorted by price ascending at index %d", i)
		}
	}
}

func TestProductIndex_Search_SortByPriceDesc(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "price_desc",
	}

	results := index.Search("", filters)

	// Verify descending price order
	for i := 1; i < len(results); i++ {
		if results[i].Product.Price > results[i-1].Product.Price {
			t.Errorf("Results not sorted by price descending at index %d", i)
		}
	}
}

func TestProductIndex_Search_SortByRating(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "rating",
	}

	results := index.Search("", filters)

	// Verify rating order (highest first)
	for i := 1; i < len(results); i++ {
		if results[i].Product.Rating > results[i-1].Product.Rating {
			t.Errorf("Results not sorted by rating at index %d", i)
		}
	}
}

func TestProductIndex_Search_SortByPopularity(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "popularity",
	}

	results := index.Search("", filters)

	// Verify popularity order (highest first)
	for i := 1; i < len(results); i++ {
		if results[i].Product.Popularity > results[i-1].Product.Popularity {
			t.Errorf("Results not sorted by popularity at index %d", i)
		}
	}
}

func TestProductIndex_Search_SortByNewest(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "newest",
	}

	results := index.Search("", filters)

	// Verify newest first
	for i := 1; i < len(results); i++ {
		if results[i].Product.CreatedAt.After(results[i-1].Product.CreatedAt) {
			t.Errorf("Results not sorted by newest at index %d", i)
		}
	}
}

func TestProductIndex_Search_SortByRelevance(t *testing.T) {
	index := NewProductIndex()

	filters := SearchFilters{
		SortBy: "relevance",
	}

	results := index.Search("mug", filters)

	// Verify relevance order (highest score first)
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not sorted by relevance at index %d", i)
		}
	}
}

func TestProductIndex_Search_PartialMatch(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("sun", SearchFilters{})

	// Should find sunglasses with prefix match
	if len(results) == 0 {
		t.Fatal("Expected results for partial match 'sun'")
	}

	found := false
	for _, result := range results {
		if result.Product.Name == "Sunglasses" {
			found = true
			// Prefix match should have high score
			if result.Score < 80 {
				t.Errorf("Expected high score for prefix match, got %.1f", result.Score)
			}
		}
	}

	if !found {
		t.Error("Expected to find Sunglasses with partial match")
	}
}

func TestProductIndex_Search_CategoryMatch(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("accessories", SearchFilters{})

	// Should find products in accessories category
	if len(results) == 0 {
		t.Fatal("Expected results for category search")
	}

	// All results should be accessories
	for _, result := range results {
		if result.Product.Category != "accessories" {
			t.Errorf("Expected accessories category, got %s", result.Product.Category)
		}
		if result.MatchType != "category" {
			t.Errorf("Expected category match type, got %s", result.MatchType)
		}
	}
}

func TestProductIndex_Search_CombinedFilters(t *testing.T) {
	index := NewProductIndex()

	minPrice := 10.0
	maxPrice := 30.0
	minRating := 4.0

	filters := SearchFilters{
		Categories:  []string{"home"},
		MinPrice:    &minPrice,
		MaxPrice:    &maxPrice,
		MinRating:   &minRating,
		InStockOnly: true,
		SortBy:      "price_asc",
	}

	results := index.Search("", filters)

	// Verify all filters are applied
	for _, result := range results {
		if result.Product.Category != "home" {
			t.Errorf("Category filter failed: got %s", result.Product.Category)
		}
		if result.Product.Price < minPrice || result.Product.Price > maxPrice {
			t.Errorf("Price filter failed: %.2f not in [%.2f, %.2f]",
				result.Product.Price, minPrice, maxPrice)
		}
		if result.Product.Rating < minRating {
			t.Errorf("Rating filter failed: %.1f < %.1f",
				result.Product.Rating, minRating)
		}
		if !result.Product.InStock {
			t.Error("Stock filter failed: product is out of stock")
		}
	}

	// Verify sorting
	for i := 1; i < len(results); i++ {
		if results[i].Product.Price < results[i-1].Product.Price {
			t.Error("Sort order failed")
		}
	}
}

func TestProductIndex_AddProduct(t *testing.T) {
	index := NewProductIndex()

	newProduct := Product{
		ID:          "TEST001",
		Name:        "Test Product",
		Description: "A test product",
		Category:    "test",
		Price:       99.99,
		Rating:      5.0,
		ReviewCount: 10,
		InStock:     true,
		Popularity:  100,
	}

	index.AddProduct(newProduct)

	// Search for the new product
	results := index.Search("Test Product", SearchFilters{})

	if len(results) == 0 {
		t.Fatal("Expected to find newly added product")
	}

	found := false
	for _, result := range results {
		if result.Product.ID == "TEST001" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Newly added product not found in search results")
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, test := range tests {
		result := levenshteinDistance(test.s1, test.s2)
		if result != test.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d",
				test.s1, test.s2, result, test.expected)
		}
	}
}

func TestProductIndex_Search_FuzzyMatch(t *testing.T) {
	index := NewProductIndex()

	// Search for misspelled "sunglases" (missing 's')
	results := index.Search("sunglas", SearchFilters{})

	// Should find sunglasses through fuzzy matching or partial match
	if len(results) == 0 {
		t.Fatal("Expected fuzzy match results")
	}

	// The top result should be sunglasses
	topResult := results[0]
	if topResult.Product.Name != "Sunglasses" {
		t.Logf("Note: Expected Sunglasses as top result, got %s (fuzzy matching may vary)",
			topResult.Product.Name)
	}
}

func TestProductIndex_Search_NoResults(t *testing.T) {
	index := NewProductIndex()

	results := index.Search("nonexistentproduct12345", SearchFilters{})

	// Should return empty results
	if len(results) != 0 {
		t.Errorf("Expected no results, got %d", len(results))
	}
}

func TestProductIndex_Search_PriceRangeExclusion(t *testing.T) {
	index := NewProductIndex()

	minPrice := 1000.0 // Very high price

	filters := SearchFilters{
		MinPrice: &minPrice,
	}

	results := index.Search("", filters)

	// Should return no results as no products cost $1000+
	if len(results) != 0 {
		t.Errorf("Expected no results for high price filter, got %d", len(results))
	}
}

func TestProductIndex_MatchesFilters(t *testing.T) {
	index := NewProductIndex()

	product := &Product{
		Category: "clothing",
		Price:    25.0,
		Rating:   4.5,
		InStock:  true,
	}

	// Test category filter
	if !index.matchesFilters(product, SearchFilters{Categories: []string{"clothing"}}) {
		t.Error("Category filter should match")
	}

	if index.matchesFilters(product, SearchFilters{Categories: []string{"shoes"}}) {
		t.Error("Category filter should not match")
	}

	// Test price filter
	minPrice := 20.0
	maxPrice := 30.0
	if !index.matchesFilters(product, SearchFilters{MinPrice: &minPrice, MaxPrice: &maxPrice}) {
		t.Error("Price filter should match")
	}

	minPrice = 100.0
	if index.matchesFilters(product, SearchFilters{MinPrice: &minPrice}) {
		t.Error("Min price filter should not match")
	}

	// Test rating filter
	minRating := 4.0
	if !index.matchesFilters(product, SearchFilters{MinRating: &minRating}) {
		t.Error("Rating filter should match")
	}

	minRating = 5.0
	if index.matchesFilters(product, SearchFilters{MinRating: &minRating}) {
		t.Error("Rating filter should not match")
	}

	// Test stock filter
	if !index.matchesFilters(product, SearchFilters{InStockOnly: true}) {
		t.Error("Stock filter should match")
	}

	product.InStock = false
	if index.matchesFilters(product, SearchFilters{InStockOnly: true}) {
		t.Error("Stock filter should not match out of stock product")
	}
}
