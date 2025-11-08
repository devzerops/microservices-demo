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
	"context"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/checkoutservice/genproto"
	money "github.com/GoogleCloudPlatform/microservices-demo/src/checkoutservice/money"
)

// TestCheckoutServiceHealthCheck tests the health check endpoint
func TestCheckoutServiceHealthCheck(t *testing.T) {
	svc := &checkoutService{}

	res, err := svc.Check(context.Background(), nil)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	if res.Status.String() != "SERVING" {
		t.Errorf("Expected SERVING status, got %s", res.Status.String())
	}
}

// TestGenerateOrderID tests order ID generation
func TestGenerateOrderID(t *testing.T) {
	// Test that order IDs are valid UUIDs
	for i := 0; i < 10; i++ {
		orderID, err := uuid.NewRandom()
		if err != nil {
			t.Errorf("Failed to generate order ID: %v", err)
		}
		if orderID.String() == "" {
			t.Error("Generated order ID is empty")
		}
		// Verify it's a valid UUID format
		if _, err := uuid.Parse(orderID.String()); err != nil {
			t.Errorf("Generated order ID is not a valid UUID: %s", orderID.String())
		}
	}
}

// TestGenerateUniqueOrderIDs tests that order IDs are unique
func TestGenerateUniqueOrderIDs(t *testing.T) {
	orderIDs := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		orderID, err := uuid.NewRandom()
		if err != nil {
			t.Fatalf("Failed to generate order ID on iteration %d: %v", i, err)
		}
		id := orderID.String()
		if orderIDs[id] {
			t.Errorf("Duplicate order ID generated: %s", id)
		}
		orderIDs[id] = true
	}

	if len(orderIDs) != iterations {
		t.Errorf("Expected %d unique order IDs, got %d", iterations, len(orderIDs))
	}
}

// TestMoneyOperations tests money calculation functions
func TestMoneyAddition(t *testing.T) {
	tests := []struct {
		name     string
		money1   pb.Money
		money2   pb.Money
		expected pb.Money
		wantErr  bool
	}{
		{
			name:     "simple addition",
			money1:   pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 500000000},
			money2:   pb.Money{CurrencyCode: "USD", Units: 5, Nanos: 250000000},
			expected: pb.Money{CurrencyCode: "USD", Units: 15, Nanos: 750000000},
			wantErr:  false,
		},
		{
			name:     "addition with carry",
			money1:   pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 700000000},
			money2:   pb.Money{CurrencyCode: "USD", Units: 5, Nanos: 500000000},
			expected: pb.Money{CurrencyCode: "USD", Units: 16, Nanos: 200000000},
			wantErr:  false,
		},
		{
			name:     "zero addition",
			money1:   pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 0},
			money2:   pb.Money{CurrencyCode: "USD", Units: 0, Nanos: 0},
			expected: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 0},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := money.Sum(tt.money1, tt.money2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !money.AreEquals(result, tt.expected) {
				t.Errorf("Sum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMoneyMultiplication tests multiplying money by quantity
func TestMoneyMultiplication(t *testing.T) {
	tests := []struct {
		name     string
		money    pb.Money
		quantity int32
		wantErr  bool
	}{
		{
			name:     "multiply by 1",
			money:    pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 990000000},
			quantity: 1,
			wantErr:  false,
		},
		{
			name:     "multiply by 0",
			money:    pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 990000000},
			quantity: 0,
			wantErr:  false,
		},
		{
			name:     "multiply by large number",
			money:    pb.Money{CurrencyCode: "USD", Units: 5, Nanos: 500000000},
			quantity: 100,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money
			for i := int32(1); i < tt.quantity; i++ {
				var err error
				result, err = money.Sum(result, tt.money)
				if (err != nil) != tt.wantErr {
					t.Errorf("Multiplication iteration %d error = %v, wantErr %v", i, err, tt.wantErr)
					return
				}
			}
			// Verify result is valid
			if !tt.wantErr && !money.IsValid(result) {
				t.Errorf("Result is invalid after multiplication: %v", result)
			}
		})
	}
}

// TestMoneyValidation tests money validation
func TestMoneyValidation(t *testing.T) {
	tests := []struct {
		name  string
		money pb.Money
		valid bool
	}{
		{
			name:  "valid positive",
			money: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 500000000},
			valid: true,
		},
		{
			name:  "valid zero",
			money: pb.Money{CurrencyCode: "USD", Units: 0, Nanos: 0},
			valid: true,
		},
		{
			name:  "valid negative",
			money: pb.Money{CurrencyCode: "USD", Units: -10, Nanos: -500000000},
			valid: true,
		},
		{
			name:  "invalid mixed signs",
			money: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: -500000000},
			valid: false,
		},
		{
			name:  "invalid nanos overflow",
			money: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 1000000000},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.IsValid(tt.money)
			if result != tt.valid {
				t.Errorf("IsValid() = %v, want %v for money %v", result, tt.valid, tt.money)
			}
		})
	}
}

// TestUSDCurrency tests USD currency constant
func TestUSDCurrency(t *testing.T) {
	if usdCurrency != "USD" {
		t.Errorf("Expected usdCurrency to be 'USD', got '%s'", usdCurrency)
	}
}

// TestStatusCodes tests gRPC status code handling
func TestStatusCodes(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		msg  string
	}{
		{"not found", codes.NotFound, "item not found"},
		{"invalid argument", codes.InvalidArgument, "invalid input"},
		{"internal error", codes.Internal, "internal error occurred"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := status.Error(tt.code, tt.msg)
			st, ok := status.FromError(err)
			if !ok {
				t.Error("Failed to extract status from error")
			}
			if st.Code() != tt.code {
				t.Errorf("Expected code %v, got %v", tt.code, st.Code())
			}
			if st.Message() != tt.msg {
				t.Errorf("Expected message '%s', got '%s'", tt.msg, st.Message())
			}
		})
	}
}
