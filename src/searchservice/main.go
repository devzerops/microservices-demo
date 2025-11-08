package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var (
	searchEngine *SearchEngine
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}
)

type SearchEngine struct {
	trie          *Trie
	trending      *TrendingTracker
	searchHistory *SearchHistory
	fuzzyMatcher  *FuzzyMatcher
	mu            sync.RWMutex
}

type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	UserID string `json:"user_id"`
}

type AutocompleteResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Took        int64        `json:"took_ms"`
}

type Suggestion struct {
	Text       string  `json:"text"`
	Score      int     `json:"score"`
	Category   string  `json:"category,omitempty"`
	Popularity int     `json:"popularity"`
	IsExact    bool    `json:"is_exact"`
}

type TrendingResponse struct {
	Trending  []TrendingQuery `json:"trending"`
	Period    string          `json:"period"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type TrendingQuery struct {
	Query      string  `json:"query"`
	Count      int     `json:"count"`
	Velocity   float64 `json:"velocity"` // Searches per minute
	Rank       int     `json:"rank"`
	Change     int     `json:"change"` // Rank change
	Percentage float64 `json:"percentage"`
}

type SearchHistoryResponse struct {
	UserID   string        `json:"user_id"`
	History  []HistoryItem `json:"history"`
	Total    int           `json:"total"`
	PageSize int           `json:"page_size"`
}

type HistoryItem struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	ResultsFound int    `json:"results_found"`
}

func NewSearchEngine() *SearchEngine {
	se := &SearchEngine{
		trie:          NewTrie(),
		trending:      NewTrendingTracker(),
		searchHistory: NewSearchHistory(),
		fuzzyMatcher:  NewFuzzyMatcher(),
	}

	// Initialize with demo data
	se.initializeDemoData()

	// Start background workers
	go se.trending.UpdateLoop()

	return se
}

func (se *SearchEngine) initializeDemoData() {
	// Product categories
	products := []struct {
		name     string
		category string
		score    int
	}{
		{"sunglasses", "accessories", 100},
		{"tank top", "clothing", 95},
		{"t-shirt", "clothing", 90},
		{"sneakers", "shoes", 85},
		{"backpack", "accessories", 80},
		{"jeans", "clothing", 75},
		{"jacket", "clothing", 70},
		{"watch", "accessories", 65},
		{"hat", "accessories", 60},
		{"shorts", "clothing", 55},
		{"socks", "clothing", 50},
		{"belt", "accessories", 45},
		{"shoes", "shoes", 40},
		{"sandals", "shoes", 35},
		{"dress", "clothing", 30},
		{"skirt", "clothing", 25},
		{"sweater", "clothing", 20},
		{"hoodie", "clothing", 15},
		{"boots", "shoes", 10},
		{"scarf", "accessories", 5},
	}

	for _, p := range products {
		se.trie.Insert(p.name, p.score, p.category)
	}

	// Add some trending searches
	trendingQueries := []string{
		"summer collection", "sale", "new arrivals",
		"sunglasses", "t-shirt", "sneakers",
	}

	for _, q := range trendingQueries {
		se.trending.Track(q)
		time.Sleep(10 * time.Millisecond)
	}
}

// HTTP Handlers

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	query := r.URL.Query().Get("q")
	limit := 10
	userID := r.URL.Query().Get("user_id")

	if query == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "query parameter 'q' is required",
		})
		return
	}

	// Track the search
	searchEngine.trending.Track(query)
	if userID != "" {
		searchEngine.searchHistory.Add(userID, query, 0) // Results count updated later
	}

	// Get suggestions from Trie
	suggestions := searchEngine.trie.Autocomplete(query, limit)

	// Add fuzzy matches if not enough exact matches
	if len(suggestions) < limit {
		fuzzySuggestions := searchEngine.fuzzyMatcher.FindSimilar(query, limit-len(suggestions))
		for _, fs := range fuzzySuggestions {
			suggestions = append(suggestions, Suggestion{
				Text:       fs.Text,
				Score:      fs.Score,
				Category:   fs.Category,
				Popularity: fs.Popularity,
				IsExact:    false,
			})
		}
	}

	took := time.Since(start).Milliseconds()

	respondJSON(w, http.StatusOK, AutocompleteResponse{
		Suggestions: suggestions,
		Took:        took,
	})
}

func trendingHandler(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1h"
	}

	trending := searchEngine.trending.GetTop(10, period)

	respondJSON(w, http.StatusOK, TrendingResponse{
		Trending:  trending,
		Period:    period,
		UpdatedAt: time.Now(),
	})
}

func searchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
		return
	}

	history := searchEngine.searchHistory.Get(userID, 20)

	respondJSON(w, http.StatusOK, SearchHistoryResponse{
		UserID:   userID,
		History:  history,
		Total:    len(history),
		PageSize: 20,
	})
}

func indexProductHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Category string `json:"category"`
		Score    int    `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	searchEngine.trie.Insert(req.Name, req.Score, req.Category)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "product indexed successfully",
		"name":    req.Name,
	})
}

func clearHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	searchEngine.searchHistory.Clear(userID)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "history cleared",
	})
}

// WebSocket handler for real-time trending updates
func trendingWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[WebSocket] Client connected for trending updates")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send initial trending data
	trending := searchEngine.trending.GetTop(10, "1h")
	if err := conn.WriteJSON(map[string]interface{}{
		"type": "snapshot",
		"data": trending,
	}); err != nil {
		log.Printf("WebSocket write error: %v", err)
		return
	}

	// Stream updates
	for range ticker.C {
		trending := searchEngine.trending.GetTop(10, "1h")

		message := map[string]interface{}{
			"type":       "update",
			"data":       trending,
			"updated_at": time.Now(),
		}

		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "search-service",
		"timestamp": time.Now(),
		"stats": map[string]interface{}{
			"indexed_terms":  searchEngine.trie.Size(),
			"trending_count": len(searchEngine.trending.GetTop(10, "1h")),
		},
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"indexed_terms":      searchEngine.trie.Size(),
		"total_searches":     searchEngine.trending.TotalSearches(),
		"unique_searchers":   searchEngine.searchHistory.UniqueUsers(),
		"trending_queries":   len(searchEngine.trending.GetTop(10, "1h")),
		"avg_response_time":  "< 5ms",
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8097"
	}

	// Initialize search engine
	searchEngine = NewSearchEngine()

	// Setup router
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/autocomplete", autocompleteHandler).Methods("GET")
	router.HandleFunc("/trending", trendingHandler).Methods("GET")
	router.HandleFunc("/history/{user_id}", searchHistoryHandler).Methods("GET")
	router.HandleFunc("/history/{user_id}", clearHistoryHandler).Methods("DELETE")
	router.HandleFunc("/index", indexProductHandler).Methods("POST")
	router.HandleFunc("/ws/trending", trendingWebSocketHandler)
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/stats", statsHandler).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	log.Printf("[Search Service] Starting on port %s", port)
	log.Printf("[Search Service] Autocomplete: GET /autocomplete?q=<query>")
	log.Printf("[Search Service] Trending: GET /trending")
	log.Printf("[Search Service] WebSocket: ws://localhost:%s/ws/trending", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
