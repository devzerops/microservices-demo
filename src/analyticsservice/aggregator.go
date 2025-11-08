package main

import (
	"sync"
	"time"
)

type Aggregator struct {
	hourly  map[string]*HourlyStats
	daily   map[string]*DailyStats
	mu      sync.RWMutex
}

type HourlyStats struct {
	Hour         time.Time
	Requests     int64
	Errors       int64
	TotalLatency float64
	UniqueUsers  map[string]bool
}

type DailyStats struct {
	Date         time.Time
	Requests     int64
	Errors       int64
	TotalLatency float64
	UniqueUsers  map[string]bool
	PeakHour     time.Time
	PeakRequests int64
}

func NewAggregator() *Aggregator {
	a := &Aggregator{
		hourly: make(map[string]*HourlyStats),
		daily:  make(map[string]*DailyStats),
	}

	// Start aggregation workers
	go a.aggregateLoop()

	return a
}

func (a *Aggregator) aggregateLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.cleanup()
	}
}

func (a *Aggregator) RecordRequest(timestamp time.Time, userID string, latency float64, isError bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Hourly aggregation
	hourKey := timestamp.Truncate(time.Hour).Format(time.RFC3339)
	if a.hourly[hourKey] == nil {
		a.hourly[hourKey] = &HourlyStats{
			Hour:        timestamp.Truncate(time.Hour),
			UniqueUsers: make(map[string]bool),
		}
	}

	hourStats := a.hourly[hourKey]
	hourStats.Requests++
	hourStats.TotalLatency += latency
	if isError {
		hourStats.Errors++
	}
	if userID != "" {
		hourStats.UniqueUsers[userID] = true
	}

	// Daily aggregation
	dayKey := timestamp.Truncate(24 * time.Hour).Format("2006-01-02")
	if a.daily[dayKey] == nil {
		a.daily[dayKey] = &DailyStats{
			Date:        timestamp.Truncate(24 * time.Hour),
			UniqueUsers: make(map[string]bool),
		}
	}

	dayStats := a.daily[dayKey]
	dayStats.Requests++
	dayStats.TotalLatency += latency
	if isError {
		dayStats.Errors++
	}
	if userID != "" {
		dayStats.UniqueUsers[userID] = true
	}

	// Update peak hour
	if hourStats.Requests > dayStats.PeakRequests {
		dayStats.PeakHour = hourStats.Hour
		dayStats.PeakRequests = hourStats.Requests
	}
}

func (a *Aggregator) GetHourlyStats(hours int) []HourlyStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make([]HourlyStats, 0)
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	for _, stats := range a.hourly {
		if stats.Hour.After(cutoff) {
			results = append(results, *stats)
		}
	}

	// Sort by hour (descending)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Hour.Before(results[j].Hour) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func (a *Aggregator) GetDailyStats(days int) []DailyStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make([]DailyStats, 0)
	cutoff := time.Now().AddDate(0, 0, -days)

	for _, stats := range a.daily {
		if stats.Date.After(cutoff) {
			results = append(results, *stats)
		}
	}

	// Sort by date (descending)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Date.Before(results[j].Date) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func (a *Aggregator) cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Remove hourly stats older than 7 days
	hourCutoff := time.Now().Add(-7 * 24 * time.Hour)
	for key, stats := range a.hourly {
		if stats.Hour.Before(hourCutoff) {
			delete(a.hourly, key)
		}
	}

	// Remove daily stats older than 90 days
	dayCutoff := time.Now().AddDate(0, 0, -90)
	for key, stats := range a.daily {
		if stats.Date.Before(dayCutoff) {
			delete(a.daily, key)
		}
	}
}

func (a *Aggregator) GetSummary() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Last 24 hours summary
	hourlyStats := a.GetHourlyStats(24)
	totalRequests := int64(0)
	totalErrors := int64(0)
	totalLatency := 0.0
	allUsers := make(map[string]bool)

	for _, stats := range hourlyStats {
		totalRequests += stats.Requests
		totalErrors += stats.Errors
		totalLatency += stats.TotalLatency
		for user := range stats.UniqueUsers {
			allUsers[user] = true
		}
	}

	avgLatency := 0.0
	if totalRequests > 0 {
		avgLatency = totalLatency / float64(totalRequests)
	}

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"period":           "24h",
		"total_requests":   totalRequests,
		"total_errors":     totalErrors,
		"error_rate":       errorRate,
		"avg_latency_ms":   avgLatency,
		"unique_users":     len(allUsers),
		"requests_per_hour": float64(totalRequests) / 24.0,
	}
}
