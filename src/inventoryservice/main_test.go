package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCalculateTotalStock(t *testing.T) {
	tests := []struct {
		name       string
		warehouses map[string]int
		expected   int
	}{
		{
			name:       "single warehouse",
			warehouses: map[string]int{"US-WEST": 100},
			expected:   100,
		},
		{
			name:       "multiple warehouses",
			warehouses: map[string]int{"US-WEST": 100, "US-EAST": 200, "EU": 50},
			expected:   350,
		},
		{
			name:       "empty warehouses",
			warehouses: map[string]int{},
			expected:   0,
		},
		{
			name:       "zero quantities",
			warehouses: map[string]int{"US-WEST": 0, "US-EAST": 0},
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalStock(tt.warehouses)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if response["service"] != "inventory-service" {
		t.Errorf("Expected service 'inventory-service', got '%v'", response["service"])
	}
}

func TestRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rootHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["service"] != "Real-time Inventory Service" {
		t.Errorf("Expected service name, got '%v'", response["service"])
	}

	if response["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", response["version"])
	}
}

func TestGetAllInventoryHandler(t *testing.T) {
	// Initialize sample inventory
	inventoryMu.Lock()
	inventory = make(map[string]*Product)
	inventory["TEST001"] = &Product{
		ProductID:      "TEST001",
		Name:           "Test Product",
		TotalStock:     100,
		ReservedStock:  10,
		AvailableStock: 90,
		Warehouses:     map[string]int{"US-WEST": 100},
	}
	inventoryMu.Unlock()

	req, err := http.NewRequest("GET", "/inventory", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getAllInventoryHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["total"] != float64(1) {
		t.Errorf("Expected 1 product, got %v", response["total"])
	}
}

func TestGetProductInventoryHandler(t *testing.T) {
	// Initialize sample inventory
	inventoryMu.Lock()
	inventory = make(map[string]*Product)
	inventory["TEST001"] = &Product{
		ProductID:      "TEST001",
		Name:           "Test Product",
		TotalStock:     100,
		ReservedStock:  10,
		AvailableStock: 90,
		Warehouses:     map[string]int{"US-WEST": 100},
	}
	inventoryMu.Unlock()

	t.Run("existing product", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/inventory/TEST001", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}", getProductInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var product Product
		if err := json.Unmarshal(rr.Body.Bytes(), &product); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if product.ProductID != "TEST001" {
			t.Errorf("Expected product ID TEST001, got %s", product.ProductID)
		}
	})

	t.Run("non-existing product", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/inventory/NONEXISTENT", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}", getProductInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
		}
	})
}

func TestUpdateInventoryHandler(t *testing.T) {
	// Initialize sample inventory
	inventoryMu.Lock()
	inventory = make(map[string]*Product)
	inventory["TEST001"] = &Product{
		ProductID:      "TEST001",
		Name:           "Test Product",
		TotalStock:     100,
		ReservedStock:  10,
		AvailableStock: 90,
		Warehouses:     map[string]int{"US-WEST": 100},
	}
	inventoryMu.Unlock()

	t.Run("valid update", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"warehouse":   "US-WEST",
			"change":      50,
			"update_type": "restock",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/inventory/TEST001/update", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}/update", updateInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var product Product
		if err := json.Unmarshal(rr.Body.Bytes(), &product); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if product.TotalStock != 150 {
			t.Errorf("Expected total stock 150, got %d", product.TotalStock)
		}
	})

	t.Run("non-existing product", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"warehouse":   "US-WEST",
			"change":      50,
			"update_type": "restock",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/inventory/NONEXISTENT/update", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}/update", updateInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
		}
	})
}

func TestReserveInventoryHandler(t *testing.T) {
	// Initialize sample inventory
	inventoryMu.Lock()
	inventory = make(map[string]*Product)
	inventory["TEST001"] = &Product{
		ProductID:      "TEST001",
		Name:           "Test Product",
		TotalStock:     100,
		ReservedStock:  0,
		AvailableStock: 100,
		Warehouses:     map[string]int{"US-WEST": 100},
	}
	inventoryMu.Unlock()

	t.Run("successful reservation", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"quantity": 10,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/inventory/TEST001/reserve", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}/reserve", reserveInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response["success"] != true {
			t.Error("Expected successful reservation")
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"quantity": 200,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/inventory/TEST001/reserve", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/inventory/{product_id}/reserve", reserveInventoryHandler)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("Expected status code %d, got %d", http.StatusConflict, status)
		}
	})
}

func TestRespondJSON(t *testing.T) {
	rr := httptest.NewRecorder()

	payload := map[string]string{
		"message": "test",
	}

	respondJSON(rr, http.StatusOK, payload)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestRespondError(t *testing.T) {
	rr := httptest.NewRecorder()

	respondError(rr, http.StatusBadRequest, "test error")

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", response["error"])
	}
}
