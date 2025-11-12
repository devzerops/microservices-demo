// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// rateLimiter implements token bucket rate limiting per session
type rateLimiter struct {
	buckets map[string]*bucket
	mu      sync.RWMutex

	// Configuration
	maxTokens     int           // Maximum tokens in bucket
	refillRate    time.Duration // Time to add one token
	cleanupPeriod time.Duration // How often to cleanup old buckets
}

// bucket represents a token bucket for rate limiting
type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(maxTokens int, refillRate time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets:       make(map[string]*bucket),
		maxTokens:     maxTokens,
		refillRate:    refillRate,
		cleanupPeriod: 5 * time.Minute,
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// allow checks if a request should be allowed and consumes a token if so
func (rl *rateLimiter) allow(sessionID string) (allowed bool, remaining int, resetTime time.Time) {
	rl.mu.Lock()
	b, exists := rl.buckets[sessionID]
	if !exists {
		b = &bucket{
			tokens:     rl.maxTokens,
			lastRefill: time.Now(),
		}
		rl.buckets[sessionID] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > rl.maxTokens {
			b.tokens = rl.maxTokens
		}
		b.lastRefill = now
	}

	// Check if we can consume a token
	if b.tokens > 0 {
		b.tokens--
		// Calculate reset time (when bucket will be full again)
		tokensNeeded := rl.maxTokens - b.tokens
		resetTime = now.Add(time.Duration(tokensNeeded) * rl.refillRate)
		return true, b.tokens, resetTime
	}

	// Calculate when next token will be available
	resetTime = b.lastRefill.Add(rl.refillRate)
	return false, 0, resetTime
}

// cleanup removes old buckets to prevent memory leaks
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupPeriod)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for sessionID, b := range rl.buckets {
			b.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(b.lastRefill) > 10*time.Minute {
				delete(rl.buckets, sessionID)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Rate limiter instances for different endpoint types
var (
	// AI assistant endpoint - strict limit (10 requests per minute)
	aiRateLimiter *rateLimiter

	// POST endpoints - moderate limit (60 requests per minute)
	postRateLimiter *rateLimiter

	// GET endpoints - relaxed limit (120 requests per minute)
	getRateLimiter *rateLimiter

	// Rate limiting enabled flag
	rateLimitingEnabled bool
)

// initRateLimiting initializes rate limiters based on environment configuration
func initRateLimiting(log logrus.FieldLogger) {
	enabled := os.Getenv("ENABLE_RATE_LIMITING")
	rateLimitingEnabled = enabled == "true" || enabled == "1"

	if !rateLimitingEnabled {
		log.Info("Rate limiting disabled")
		return
	}

	// AI assistant endpoint: 10 requests per minute (1 token every 6 seconds)
	aiLimit := getEnvInt("RATE_LIMIT_AI", 10)
	aiRateLimiter = newRateLimiter(aiLimit, 60*time.Second/time.Duration(aiLimit))

	// POST endpoints: 60 requests per minute (1 token every second)
	postLimit := getEnvInt("RATE_LIMIT_POST", 60)
	postRateLimiter = newRateLimiter(postLimit, 60*time.Second/time.Duration(postLimit))

	// GET endpoints: 120 requests per minute (1 token every 0.5 seconds)
	getLimit := getEnvInt("RATE_LIMIT_GET", 120)
	getRateLimiter = newRateLimiter(getLimit, 60*time.Second/time.Duration(getLimit))

	log.Infof("Rate limiting enabled: AI=%d/min, POST=%d/min, GET=%d/min",
		aiLimit, postLimit, getLimit)
}

// getEnvInt gets an integer from environment variable or returns default
func getEnvInt(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil && val > 0 {
			return val
		}
	}
	return defaultValue
}

// rateLimitMiddleware applies rate limiting to HTTP requests
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if rate limiting is disabled
		if !rateLimitingEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get session ID from context
		sessionID, ok := r.Context().Value(ctxKeySessionID{}).(string)
		if !ok || sessionID == "" {
			// If no session ID, allow but log warning
			if log, ok := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger); ok {
				log.Warn("Rate limiting skipped: no session ID")
			}
			next.ServeHTTP(w, r)
			return
		}

		// Select appropriate rate limiter based on endpoint
		var limiter *rateLimiter
		var limitType string

		if r.URL.Path == baseUrl+"/bot" {
			limiter = aiRateLimiter
			limitType = "AI"
		} else if r.Method == http.MethodPost {
			limiter = postRateLimiter
			limitType = "POST"
		} else {
			limiter = getRateLimiter
			limitType = "GET"
		}

		// Check rate limit
		allowed, remaining, resetTime := limiter.allow(sessionID)

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.maxTokens))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// Log rate limit exceeded
			if log, ok := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger); ok {
				log.WithFields(logrus.Fields{
					"session":      sessionID,
					"endpoint":     r.URL.Path,
					"limit_type":   limitType,
					"reset_time":   resetTime.Format(time.RFC3339),
				}).Warn("Rate limit exceeded")
			}

			// Return 429 Too Many Requests
			w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds())+1, 10))
			http.Error(w, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.",
				int(time.Until(resetTime).Seconds())+1), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
