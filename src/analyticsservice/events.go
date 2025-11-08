package main

import (
	"sync"
	"time"
)

type EventTracker struct {
	events          []Event
	eventsByType    map[string][]Event
	eventsByService map[string][]Event
	recentMinute    []Event
	mu              sync.RWMutex
}

func NewEventTracker() *EventTracker {
	et := &EventTracker{
		events:          make([]Event, 0),
		eventsByType:    make(map[string][]Event),
		eventsByService: make(map[string][]Event),
		recentMinute:    make([]Event, 0),
	}

	// Start cleanup worker
	go et.cleanupLoop()

	return et
}

func (et *EventTracker) Track(event Event) {
	et.mu.Lock()
	defer et.mu.Unlock()

	// Add to main list
	et.events = append(et.events, event)

	// Add to type index
	if et.eventsByType[event.Type] == nil {
		et.eventsByType[event.Type] = make([]Event, 0)
	}
	et.eventsByType[event.Type] = append(et.eventsByType[event.Type], event)

	// Add to service index
	if et.eventsByService[event.Service] == nil {
		et.eventsByService[event.Service] = make([]Event, 0)
	}
	et.eventsByService[event.Service] = append(et.eventsByService[event.Service], event)

	// Add to recent events
	et.recentMinute = append(et.recentMinute, event)

	// Limit main events list
	if len(et.events) > 10000 {
		et.events = et.events[1000:]
	}
}

func (et *EventTracker) GetEvents(limit int, eventType, service string) []Event {
	et.mu.RLock()
	defer et.mu.RUnlock()

	var source []Event

	if eventType != "" {
		source = et.eventsByType[eventType]
	} else if service != "" {
		source = et.eventsByService[service]
	} else {
		source = et.events
	}

	// Get latest events
	start := len(source) - limit
	if start < 0 {
		start = 0
	}

	result := make([]Event, len(source)-start)
	copy(result, source[start:])

	// Reverse to get newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (et *EventTracker) GetRealtimeStats() RealtimeStats {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Count recent events
	requestsLast1Min := 0
	errorsLast1Min := 0
	cutoff := time.Now().Add(-1 * time.Minute)

	for _, event := range et.recentMinute {
		if event.Timestamp.After(cutoff) {
			if event.Type == "request" {
				requestsLast1Min++
			} else if event.Type == "error" {
				errorsLast1Min++
			}
		}
	}

	// Get top endpoints
	endpointCounts := make(map[string]int)
	endpointTimes := make(map[string][]float64)

	for _, event := range et.recentMinute {
		if event.Timestamp.After(cutoff) {
			if endpoint, ok := event.Data["endpoint"].(string); ok {
				endpointCounts[endpoint]++

				if latency, ok := event.Data["latency"].(float64); ok {
					endpointTimes[endpoint] = append(endpointTimes[endpoint], latency)
				}
			}
		}
	}

	// Convert to sorted list
	topEndpoints := make([]EndpointStat, 0)
	for endpoint, count := range endpointCounts {
		avgTime := 0.0
		if times, exists := endpointTimes[endpoint]; exists && len(times) > 0 {
			total := 0.0
			for _, t := range times {
				total += t
			}
			avgTime = total / float64(len(times))
		}

		topEndpoints = append(topEndpoints, EndpointStat{
			Endpoint: endpoint,
			Count:    count,
			AvgTime:  avgTime,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(topEndpoints); i++ {
		for j := i + 1; j < len(topEndpoints); j++ {
			if topEndpoints[i].Count < topEndpoints[j].Count {
				topEndpoints[i], topEndpoints[j] = topEndpoints[j], topEndpoints[i]
			}
		}
	}

	// Limit to top 10
	if len(topEndpoints) > 10 {
		topEndpoints = topEndpoints[:10]
	}

	// Get recent events (last 20)
	recentEvents := et.GetEvents(20, "", "")

	return RealtimeStats{
		RequestsLast1Min:  requestsLast1Min,
		ErrorsLast1Min:    errorsLast1Min,
		ActiveConnections: 0, // Will be set from external source
		TopEndpoints:      topEndpoints,
		RecentEvents:      recentEvents,
	}
}

func (et *EventTracker) TotalEvents() int {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return len(et.events)
}

func (et *EventTracker) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		et.cleanup()
	}
}

func (et *EventTracker) cleanup() {
	et.mu.Lock()
	defer et.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Minute)

	// Clean up recent minute events
	newRecent := make([]Event, 0)
	for _, event := range et.recentMinute {
		if event.Timestamp.After(cutoff) {
			newRecent = append(newRecent, event)
		}
	}
	et.recentMinute = newRecent

	// Clean up old events (keep last hour)
	hourCutoff := time.Now().Add(-1 * time.Hour)
	newEvents := make([]Event, 0)
	for _, event := range et.events {
		if event.Timestamp.After(hourCutoff) {
			newEvents = append(newEvents, event)
		}
	}
	et.events = newEvents

	// Rebuild indices
	et.eventsByType = make(map[string][]Event)
	et.eventsByService = make(map[string][]Event)

	for _, event := range et.events {
		et.eventsByType[event.Type] = append(et.eventsByType[event.Type], event)
		et.eventsByService[event.Service] = append(et.eventsByService[event.Service], event)
	}
}

func (et *EventTracker) GetEventCountByType() map[string]int {
	et.mu.RLock()
	defer et.mu.RUnlock()

	result := make(map[string]int)
	for eventType, events := range et.eventsByType {
		result[eventType] = len(events)
	}
	return result
}

func (et *EventTracker) GetEventCountByService() map[string]int {
	et.mu.RLock()
	defer et.mu.RUnlock()

	result := make(map[string]int)
	for service, events := range et.eventsByService {
		result[service] = len(events)
	}
	return result
}
