package main

import (
	"sync"
	"time"
)

type MetricsCollector struct {
	services      map[string]*ServiceMetrics
	allMetrics    []MetricValue
	requestCounts map[string]int64
	errorCounts   map[string]int64
	latencies     map[string][]float64
	mu            sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		services:      make(map[string]*ServiceMetrics),
		allMetrics:    make([]MetricValue, 0),
		requestCounts: make(map[string]int64),
		errorCounts:   make(map[string]int64),
		latencies:     make(map[string][]float64),
	}
}

func (mc *MetricsCollector) Record(metric MetricValue) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.allMetrics = append(mc.allMetrics, metric)

	// Keep only last 10000 metrics
	if len(mc.allMetrics) > 10000 {
		mc.allMetrics = mc.allMetrics[1:]
	}

	// Update specific counters based on metric name
	service := metric.Tags["service"]
	if service == "" {
		return
	}

	switch metric.Name {
	case "requests":
		mc.requestCounts[service] += int64(metric.Value)
	case "errors":
		mc.errorCounts[service] += int64(metric.Value)
	case "latency":
		mc.latencies[service] = append(mc.latencies[service], metric.Value)
		// Keep only last 1000 latencies
		if len(mc.latencies[service]) > 1000 {
			mc.latencies[service] = mc.latencies[service][1:]
		}
	}
}

func (mc *MetricsCollector) RecordHeartbeat(service, status string, metrics map[string]interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.services[service]; !exists {
		mc.services[service] = &ServiceMetrics{
			Name:   service,
			Status: status,
		}
	}

	svc := mc.services[service]
	svc.Status = status
	svc.LastHeartbeat = time.Now()

	// Update metrics from heartbeat
	if reqCount, ok := metrics["request_count"].(float64); ok {
		svc.RequestCount = int64(reqCount)
	}
	if errCount, ok := metrics["error_count"].(float64); ok {
		svc.ErrorCount = int64(errCount)
	}
	if avgLat, ok := metrics["avg_latency"].(float64); ok {
		svc.AvgLatency = avgLat
	}
	if uptime, ok := metrics["uptime"].(float64); ok {
		svc.Uptime = uptime
	}
}

func (mc *MetricsCollector) GetOverview() OverviewMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalRequests int64
	var totalErrors int64
	var totalLatency float64
	var latencyCount int
	activeUsers := 0

	for service, count := range mc.requestCounts {
		totalRequests += count

		if errCount, exists := mc.errorCounts[service]; exists {
			totalErrors += errCount
		}

		if latencies, exists := mc.latencies[service]; exists {
			for _, lat := range latencies {
				totalLatency += lat
				latencyCount++
			}
		}
	}

	avgLatency := 0.0
	if latencyCount > 0 {
		avgLatency = totalLatency / float64(latencyCount)
	}

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	// Calculate requests per second (last minute)
	requestsPerSecond := float64(totalRequests) / 60.0

	return OverviewMetrics{
		TotalRequests:     totalRequests,
		ActiveUsers:       activeUsers,
		AverageLatency:    avgLatency,
		ErrorRate:         errorRate,
		RequestsPerSecond: requestsPerSecond,
	}
}

func (mc *MetricsCollector) GetServiceMetrics() map[string]ServiceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]ServiceMetrics)

	for name, svc := range mc.services {
		// Calculate uptime
		if svc.LastHeartbeat.IsZero() {
			svc.Uptime = 0
		} else {
			svc.Uptime = time.Since(startTime).Hours()
		}

		// Get latest metrics
		if reqCount, exists := mc.requestCounts[name]; exists {
			svc.RequestCount = reqCount
		}
		if errCount, exists := mc.errorCounts[name]; exists {
			svc.ErrorCount = errCount
		}

		// Calculate average latency
		if latencies, exists := mc.latencies[name]; exists && len(latencies) > 0 {
			total := 0.0
			for _, lat := range latencies {
				total += lat
			}
			svc.AvgLatency = total / float64(len(latencies))
		}

		result[name] = *svc
	}

	// Add default services if not reporting
	defaultServices := []string{
		"visualsearch", "gamification", "inventory", "pwa",
		"search", "analytics", "productcatalog", "cart", "checkout",
	}

	for _, name := range defaultServices {
		if _, exists := result[name]; !exists {
			result[name] = ServiceMetrics{
				Name:          name,
				Status:        "unknown",
				Uptime:        0,
				RequestCount:  0,
				ErrorCount:    0,
				AvgLatency:    0,
				LastHeartbeat: time.Time{},
			}
		}
	}

	return result
}

func (mc *MetricsCollector) TotalMetrics() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.allMetrics)
}

func (mc *MetricsCollector) GetMetricsByName(name string, limit int) []MetricValue {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	results := []MetricValue{}

	// Get metrics in reverse order (newest first)
	for i := len(mc.allMetrics) - 1; i >= 0 && len(results) < limit; i-- {
		if mc.allMetrics[i].Name == name {
			results = append(results, mc.allMetrics[i])
		}
	}

	return results
}

func (mc *MetricsCollector) GetLatencyPercentile(service string, percentile float64) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	latencies, exists := mc.latencies[service]
	if !exists || len(latencies) == 0 {
		return 0
	}

	// Simple percentile calculation
	sorted := make([]float64, len(latencies))
	copy(sorted, latencies)

	// Bubble sort (simple for small arrays)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
