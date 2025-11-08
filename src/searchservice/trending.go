package main

import (
	"sort"
	"sync"
	"time"
)

type TrendingTracker struct {
	queries map[string]*QueryStats
	mu      sync.RWMutex
}

type QueryStats struct {
	Query        string
	Count        int
	LastSearched time.Time
	Timestamps   []time.Time
	Rank         int
	PrevRank     int
}

func NewTrendingTracker() *TrendingTracker {
	return &TrendingTracker{
		queries: make(map[string]*QueryStats),
	}
}

func (tt *TrendingTracker) Track(query string) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	query = normalizeQuery(query)
	if query == "" {
		return
	}

	now := time.Now()

	if stats, exists := tt.queries[query]; exists {
		stats.Count++
		stats.LastSearched = now
		stats.Timestamps = append(stats.Timestamps, now)
	} else {
		tt.queries[query] = &QueryStats{
			Query:        query,
			Count:        1,
			LastSearched: now,
			Timestamps:   []time.Time{now},
			Rank:         0,
			PrevRank:     0,
		}
	}
}

func (tt *TrendingTracker) GetTop(n int, period string) []TrendingQuery {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	// Parse period
	duration := parsePeriod(period)
	cutoff := time.Now().Add(-duration)

	// Calculate trending scores
	trending := []TrendingQuery{}
	total := 0

	for _, stats := range tt.queries {
		// Count searches in period
		count := 0
		for _, ts := range stats.Timestamps {
			if ts.After(cutoff) {
				count++
			}
		}

		if count > 0 {
			velocity := float64(count) / duration.Minutes()

			trending = append(trending, TrendingQuery{
				Query:      stats.Query,
				Count:      count,
				Velocity:   velocity,
				Rank:       0,
				Change:     0,
				Percentage: 0,
			})

			total += count
		}
	}

	// Sort by count (descending)
	sort.Slice(trending, func(i, j int) bool {
		if trending[i].Count == trending[j].Count {
			return trending[i].Velocity > trending[j].Velocity
		}
		return trending[i].Count > trending[j].Count
	})

	// Assign ranks and calculate percentages
	for i := range trending {
		trending[i].Rank = i + 1
		trending[i].Percentage = float64(trending[i].Count) / float64(total) * 100

		// Calculate rank change (mock for now)
		if stats, exists := tt.queries[trending[i].Query]; exists {
			trending[i].Change = stats.PrevRank - (i + 1)
			stats.PrevRank = stats.Rank
			stats.Rank = i + 1
		}
	}

	// Limit results
	if len(trending) > n {
		trending = trending[:n]
	}

	return trending
}

func (tt *TrendingTracker) TotalSearches() int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	total := 0
	for _, stats := range tt.queries {
		total += stats.Count
	}
	return total
}

func (tt *TrendingTracker) UpdateLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tt.cleanup()
	}
}

func (tt *TrendingTracker) cleanup() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)

	for query, stats := range tt.queries {
		// Remove old timestamps
		newTimestamps := []time.Time{}
		for _, ts := range stats.Timestamps {
			if ts.After(cutoff) {
				newTimestamps = append(newTimestamps, ts)
			}
		}

		if len(newTimestamps) == 0 {
			delete(tt.queries, query)
		} else {
			stats.Timestamps = newTimestamps
			stats.Count = len(newTimestamps)
		}
	}
}

func parsePeriod(period string) time.Duration {
	switch period {
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return 1 * time.Hour
	}
}

func normalizeQuery(query string) string {
	// Simple normalization (lowercase, trim)
	// In production, you might want more sophisticated normalization
	return strings.ToLower(strings.TrimSpace(query))
}
