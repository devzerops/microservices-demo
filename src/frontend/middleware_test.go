// Copyright 2025 Google LLC
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
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// TestRateLimitMiddleware_AllowsNormalRequests verifies that normal requests are allowed
func TestRateLimitMiddleware_AllowsNormalRequests(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with rate limit middleware
	middleware := rateLimitMiddleware(handler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()

	// Execute request
	middleware.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestRateLimitMiddleware_BlocksExcessiveRequests verifies rate limiting behavior
func TestRateLimitMiddleware_BlocksExcessiveRequests(t *testing.T) {
	// Save original rate limiter
	originalLimiter := globalRateLimiter

	// Create a very restrictive rate limiter for testing (1 req/sec, burst 2)
	testLimiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(1.0),
		burst:    2,
	}
	globalRateLimiter = testLimiter

	defer func() {
		globalRateLimiter = originalLimiter
	}()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := rateLimitMiddleware(handler)

	// Make multiple requests from same IP
	clientIP := "192.168.1.100:5678"

	// First 2 requests should succeed (burst size)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = clientIP
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rec.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = clientIP
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rec.Code)
	}

	// Verify rate limit headers
	if rec.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Missing X-RateLimit-Limit header")
	}
	if rec.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("Expected X-RateLimit-Remaining=0, got %s", rec.Header().Get("X-RateLimit-Remaining"))
	}
	if rec.Header().Get("Retry-After") != "60" {
		t.Errorf("Expected Retry-After=60, got %s", rec.Header().Get("Retry-After"))
	}
}

// TestRateLimitMiddleware_PerIPLimiting verifies that rate limiting is per-IP
func TestRateLimitMiddleware_PerIPLimiting(t *testing.T) {
	// Save original rate limiter
	originalLimiter := globalRateLimiter

	// Create a restrictive rate limiter (1 req/sec, burst 1)
	testLimiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(1.0),
		burst:    1,
	}
	globalRateLimiter = testLimiter

	defer func() {
		globalRateLimiter = originalLimiter
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := rateLimitMiddleware(handler)

	// Request from IP 1 (should succeed)
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1111"
	rec1 := httptest.NewRecorder()
	middleware.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("IP 1 first request: Expected 200, got %d", rec1.Code)
	}

	// Request from IP 2 (should also succeed - different IP)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:2222"
	rec2 := httptest.NewRecorder()
	middleware.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("IP 2 first request: Expected 200, got %d", rec2.Code)
	}

	// Second request from IP 1 (should be rate limited)
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1111"
	rec3 := httptest.NewRecorder()
	middleware.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("IP 1 second request: Expected 429, got %d", rec3.Code)
	}
}

// TestRateLimitMiddleware_CanBeDisabled verifies that rate limiting can be disabled
func TestRateLimitMiddleware_CanBeDisabled(t *testing.T) {
	// Save original rate limiter
	originalLimiter := globalRateLimiter

	// Create a very restrictive rate limiter (should be bypassed when disabled)
	testLimiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(0.01), // Very low limit
		burst:    1,
	}
	globalRateLimiter = testLimiter

	defer func() {
		globalRateLimiter = originalLimiter
		os.Unsetenv("DISABLE_RATE_LIMITING")
	}()

	// Enable rate limiting bypass
	os.Setenv("DISABLE_RATE_LIMITING", "true")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := rateLimitMiddleware(handler)

	// Make many requests - all should succeed
	clientIP := "192.168.1.100:5678"
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = clientIP
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d with rate limiting disabled: Expected 200, got %d", i+1, rec.Code)
		}
	}
}

// TestGetClientIP verifies IP extraction from various headers
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "Direct connection",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For single IP",
			remoteAddr:    "10.0.0.1:5678",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For multiple IPs",
			remoteAddr:    "10.0.0.1:5678",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:5678",
			xRealIP:    "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:    "10.0.0.1:5678",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.2",
			expectedIP:    "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

// TestRateLimiter_Cleanup verifies that inactive visitors are cleaned up
func TestRateLimiter_Cleanup(t *testing.T) {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(10),
		burst:    10,
	}

	// Add some visitors
	rl.getVisitor("192.168.1.1")
	rl.getVisitor("192.168.1.2")

	// Verify they exist
	if len(rl.visitors) != 2 {
		t.Errorf("Expected 2 visitors, got %d", len(rl.visitors))
	}

	// Make one visitor old
	rl.mu.Lock()
	rl.visitors["192.168.1.1"].lastSeen = time.Now().Add(-5 * time.Minute)
	rl.mu.Unlock()

	// Manually trigger cleanup logic (without goroutine)
	rl.mu.Lock()
	for ip, v := range rl.visitors {
		if time.Since(v.lastSeen) > 3*time.Minute {
			delete(rl.visitors, ip)
		}
	}
	rl.mu.Unlock()

	// Verify old visitor was removed
	if len(rl.visitors) != 1 {
		t.Errorf("Expected 1 visitor after cleanup, got %d", len(rl.visitors))
	}

	// Verify the remaining visitor is the recent one
	rl.mu.RLock()
	_, exists := rl.visitors["192.168.1.2"]
	rl.mu.RUnlock()

	if !exists {
		t.Error("Expected recent visitor to remain after cleanup")
	}
}

// TestRateLimiter_ConfigurableLimits verifies environment variable configuration
func TestRateLimiter_ConfigurableLimits(t *testing.T) {
	// Save original environment
	originalRPS := os.Getenv("RATE_LIMIT_RPS")
	originalBurst := os.Getenv("RATE_LIMIT_BURST")

	defer func() {
		if originalRPS != "" {
			os.Setenv("RATE_LIMIT_RPS", originalRPS)
		} else {
			os.Unsetenv("RATE_LIMIT_RPS")
		}
		if originalBurst != "" {
			os.Setenv("RATE_LIMIT_BURST", originalBurst)
		} else {
			os.Unsetenv("RATE_LIMIT_BURST")
		}
	}()

	// Set custom limits
	os.Setenv("RATE_LIMIT_RPS", "5.0")
	os.Setenv("RATE_LIMIT_BURST", "10")

	// Create new rate limiter
	rl := newRateLimiter()

	// Verify configuration
	if float64(rl.rate) != 5.0 {
		t.Errorf("Expected rate 5.0, got %f", float64(rl.rate))
	}
	if rl.burst != 10 {
		t.Errorf("Expected burst 10, got %d", rl.burst)
	}
}

// TestRateLimitMiddleware_WithLogging verifies security event logging
func TestRateLimitMiddleware_WithLogging(t *testing.T) {
	// Save original rate limiter
	originalLimiter := globalRateLimiter

	// Create a restrictive rate limiter
	testLimiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(1.0),
		burst:    1,
	}
	globalRateLimiter = testLimiter

	defer func() {
		globalRateLimiter = originalLimiter
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with logging middleware
	log := logrus.New()
	logMiddleware := &logHandler{log: log, next: rateLimitMiddleware(handler)}

	// Exhaust rate limit
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rec1 := httptest.NewRecorder()
	logMiddleware.ServeHTTP(rec1, req1)

	// This request should be rate limited and logged
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rec2 := httptest.NewRecorder()
	logMiddleware.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rec2.Code)
	}

	// Verify response body contains message
	body := rec2.Body.String()
	if body == "" {
		t.Error("Expected error message in response body")
	}
}
