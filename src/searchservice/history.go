package main

import (
	"sync"
	"time"
)

type SearchHistory struct {
	userHistory map[string][]HistoryItem
	mu          sync.RWMutex
}

func NewSearchHistory() *SearchHistory {
	return &SearchHistory{
		userHistory: make(map[string][]HistoryItem),
	}
}

func (sh *SearchHistory) Add(userID, query string, resultsFound int) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	item := HistoryItem{
		Query:        query,
		Timestamp:    time.Now(),
		ResultsFound: resultsFound,
	}

	if history, exists := sh.userHistory[userID]; exists {
		// Add to beginning
		sh.userHistory[userID] = append([]HistoryItem{item}, history...)

		// Keep only last 100 items
		if len(sh.userHistory[userID]) > 100 {
			sh.userHistory[userID] = sh.userHistory[userID][:100]
		}
	} else {
		sh.userHistory[userID] = []HistoryItem{item}
	}
}

func (sh *SearchHistory) Get(userID string, limit int) []HistoryItem {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	history, exists := sh.userHistory[userID]
	if !exists {
		return []HistoryItem{}
	}

	if len(history) > limit {
		return history[:limit]
	}

	return history
}

func (sh *SearchHistory) Clear(userID string) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	delete(sh.userHistory, userID)
}

func (sh *SearchHistory) UniqueUsers() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	return len(sh.userHistory)
}

func (sh *SearchHistory) GetRecentSearches(userID string, count int) []string {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	history, exists := sh.userHistory[userID]
	if !exists {
		return []string{}
	}

	queries := []string{}
	seen := make(map[string]bool)

	for _, item := range history {
		if !seen[item.Query] {
			queries = append(queries, item.Query)
			seen[item.Query] = true

			if len(queries) >= count {
				break
			}
		}
	}

	return queries
}
