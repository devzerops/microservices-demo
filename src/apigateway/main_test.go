package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns environment variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default")
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}
	})

	t.Run("returns default value when env not set", func(t *testing.T) {
		result := getEnv("NON_EXISTENT_VAR", "default_value")
		if result != "default_value" {
			t.Errorf("Expected 'default_value', got '%s'", result)
		}
	})

	t.Run("returns default value for empty env", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		result := getEnv("EMPTY_VAR", "default")
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})
}

func TestGetClientIP(t *testing.T) {
	t.Run("extracts IP from X-Forwarded-For", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")

		ip := getClientIP(req)
		if ip != "192.168.1.1" {
			t.Errorf("Expected '192.168.1.1', got '%s'", ip)
		}
	})

	t.Run("extracts IP from X-Real-IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "192.168.1.2")

		ip := getClientIP(req)
		if ip != "192.168.1.2" {
			t.Errorf("Expected '192.168.1.2', got '%s'", ip)
		}
	})

	t.Run("prefers X-Forwarded-For over X-Real-IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("X-Real-IP", "192.168.1.2")

		ip := getClientIP(req)
		if ip != "192.168.1.1" {
			t.Errorf("Expected '192.168.1.1' from X-Forwarded-For, got '%s'", ip)
		}
	})

	t.Run("falls back to RemoteAddr", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.3:12345"

		ip := getClientIP(req)
		if ip != "192.168.1.3" {
			t.Errorf("Expected '192.168.1.3' from RemoteAddr, got '%s'", ip)
		}
	})
}

func TestRespondJSON(t *testing.T) {
	t.Run("writes JSON response with correct content type", func(t *testing.T) {
		rr := httptest.NewRecorder()

		data := map[string]string{
			"message": "test",
			"status":  "ok",
		}

		respondJSON(rr, http.StatusOK, data)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if response["message"] != "test" {
			t.Errorf("Expected message 'test', got '%s'", response["message"])
		}
	})

	t.Run("handles different status codes", func(t *testing.T) {
		testCases := []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		}

		for _, statusCode := range testCases {
			rr := httptest.NewRecorder()
			respondJSON(rr, statusCode, map[string]string{"status": "test"})

			if rr.Code != statusCode {
				t.Errorf("Expected status code %d, got %d", statusCode, rr.Code)
			}
		}
	})
}

func TestGatewayErrorHandler(t *testing.T) {
	gateway := NewGateway()

	req := httptest.NewRequest("GET", "/test/path", nil)
	rr := httptest.NewRecorder()

	// Simulate an error
	err := http.ErrServerClosed

	gateway.errorHandler(rr, req, err)

	if status := rr.Code; status != http.StatusBadGateway {
		t.Errorf("Expected status code %d, got %d", http.StatusBadGateway, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["error"] != "service_unavailable" {
		t.Errorf("Expected error 'service_unavailable', got '%v'", response["error"])
	}

	if response["path"] != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%v'", response["path"])
	}
}

func TestNewGateway(t *testing.T) {
	gateway := NewGateway()

	if gateway == nil {
		t.Fatal("Expected gateway to be created")
	}

	if gateway.proxies == nil {
		t.Fatal("Expected proxies map to be initialized")
	}

	// Check that proxies were created for services
	expectedServices := []string{"visualsearch", "gamification", "inventory", "pwa", "search", "analytics"}

	for _, service := range expectedServices {
		if gateway.proxies[service] == nil {
			t.Errorf("Expected proxy for service '%s' to be created", service)
		}
	}
}

func TestServicePathTrimming(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		service  string
		expected string
	}{
		{
			name:     "trims service prefix",
			input:    "/visualsearch/search",
			service:  "visualsearch",
			expected: "/search",
		},
		{
			name:     "handles root path",
			input:    "/visualsearch",
			service:  "visualsearch",
			expected: "/",
		},
		{
			name:     "handles nested paths",
			input:    "/gamification/api/points",
			service:  "gamification",
			expected: "/api/points",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := strings.TrimPrefix(tt.input, "/"+tt.service)
			if path == "" {
				path = "/"
			}

			if path != tt.expected {
				t.Errorf("Expected path '%s', got '%s'", tt.expected, path)
			}
		})
	}
}
