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
	"regexp"
	"testing"

	"golang.org/x/net/context"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/shippingservice/genproto"
)

// TestGetQuoteEmptyCart tests getting quote for empty cart
func TestGetQuoteEmptyCart(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: &pb.Address{
			StreetAddress: "123 Main St",
			City:          "Seattle",
			State:         "WA",
			Country:       "USA",
		},
		Items: []*pb.CartItem{},
	}

	res, err := s.GetQuote(context.Background(), req)
	if err != nil {
		t.Errorf("TestGetQuoteEmptyCart failed: %v", err)
	}
	if res.CostUsd.GetUnits() != 0 || res.CostUsd.GetNanos() != 0 {
		t.Errorf("Expected zero cost for empty cart, got %d.%d", res.CostUsd.GetUnits(), res.CostUsd.GetNanos())
	}
}

// TestGetQuoteSingleItem tests getting quote for single item
func TestGetQuoteSingleItem(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: &pb.Address{
			StreetAddress: "456 Oak Ave",
			City:          "Portland",
			State:         "OR",
			Country:       "USA",
		},
		Items: []*pb.CartItem{
			{
				ProductId: "test-product-1",
				Quantity:  1,
			},
		},
	}

	res, err := s.GetQuote(context.Background(), req)
	if err != nil {
		t.Errorf("TestGetQuoteSingleItem failed: %v", err)
	}
	// Quote should be calculated as: items * $8.99 = $8.99
	if res.CostUsd.GetUnits() != 8 || res.CostUsd.GetNanos() != 990000000 {
		t.Errorf("Expected cost 8.99, got %d.%d", res.CostUsd.GetUnits(), res.CostUsd.GetNanos())
	}
}

// TestGetQuoteLargeQuantity tests getting quote for large quantity
func TestGetQuoteLargeQuantity(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: &pb.Address{
			StreetAddress: "789 Elm St",
			City:          "Boston",
			State:         "MA",
			Country:       "USA",
		},
		Items: []*pb.CartItem{
			{
				ProductId: "large-order",
				Quantity:  100,
			},
		},
	}

	res, err := s.GetQuote(context.Background(), req)
	if err != nil {
		t.Errorf("TestGetQuoteLargeQuantity failed: %v", err)
	}
	// Should handle large quantities without error
	if res.CostUsd == nil {
		t.Error("Expected valid cost for large quantity")
	}
}

// TestGetQuoteMultipleAddresses tests quote consistency across different addresses
func TestGetQuoteMultipleAddresses(t *testing.T) {
	s := server{}

	items := []*pb.CartItem{
		{
			ProductId: "prod-1",
			Quantity:  2,
		},
	}

	addresses := []*pb.Address{
		{StreetAddress: "123 A St", City: "New York", State: "NY", Country: "USA"},
		{StreetAddress: "456 B Ave", City: "Los Angeles", State: "CA", Country: "USA"},
		{StreetAddress: "789 C Blvd", City: "Chicago", State: "IL", Country: "USA"},
	}

	for _, addr := range addresses {
		req := &pb.GetQuoteRequest{
			Address: addr,
			Items:   items,
		}

		res, err := s.GetQuote(context.Background(), req)
		if err != nil {
			t.Errorf("GetQuote failed for address %v: %v", addr, err)
		}
		// All should return same cost since cost doesn't depend on address
		if res.CostUsd.GetUnits() != 17 || res.CostUsd.GetNanos() != 980000000 {
			t.Errorf("Expected consistent cost across addresses, got %d.%d for %s",
				res.CostUsd.GetUnits(), res.CostUsd.GetNanos(), addr.City)
		}
	}
}

// TestShipOrderTrackingIDFormat tests tracking ID format validation
func TestShipOrderTrackingIDFormat(t *testing.T) {
	s := server{}

	req := &pb.ShipOrderRequest{
		Address: &pb.Address{
			StreetAddress: "Test St",
			City:          "Test City",
			State:         "TS",
			Country:       "Test",
		},
		Items: []*pb.CartItem{
			{
				ProductId: "test-product",
				Quantity:  1,
			},
		},
	}

	res, err := s.ShipOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("TestShipOrderTrackingIDFormat failed: %v", err)
	}

	// Verify tracking ID format: should be like "XX-12345-67890"
	trackingIDPattern := regexp.MustCompile(`^[A-Z]{2}-\d{5}-\d{5}$`)
	if !trackingIDPattern.MatchString(res.TrackingId) {
		t.Errorf("Tracking ID format is invalid: %s (expected format: XX-#####-#####)", res.TrackingId)
	}
}

// TestShipOrderMultipleItems tests shipping multiple items
func TestShipOrderMultipleItems(t *testing.T) {
	s := server{}

	req := &pb.ShipOrderRequest{
		Address: &pb.Address{
			StreetAddress: "Multi Item St",
			City:          "Test",
			State:         "TS",
			Country:       "USA",
		},
		Items: []*pb.CartItem{
			{ProductId: "item1", Quantity: 2},
			{ProductId: "item2", Quantity: 5},
			{ProductId: "item3", Quantity: 1},
		},
	}

	res, err := s.ShipOrder(context.Background(), req)
	if err != nil {
		t.Errorf("TestShipOrderMultipleItems failed: %v", err)
	}
	if len(res.TrackingId) == 0 {
		t.Error("Tracking ID should not be empty")
	}
}

// TestShipOrderUniqueness tests that tracking IDs are unique
func TestShipOrderUniqueness(t *testing.T) {
	s := server{}

	req := &pb.ShipOrderRequest{
		Address: &pb.Address{
			StreetAddress: "Unique Test St",
			City:          "Uniqueville",
			State:         "UN",
			Country:       "USA",
		},
		Items: []*pb.CartItem{
			{
				ProductId: "unique-product",
				Quantity:  1,
			},
		},
	}

	trackingIDs := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		res, err := s.ShipOrder(context.Background(), req)
		if err != nil {
			t.Fatalf("ShipOrder failed on iteration %d: %v", i, err)
		}
		if trackingIDs[res.TrackingId] {
			t.Errorf("Duplicate tracking ID generated: %s", res.TrackingId)
		}
		trackingIDs[res.TrackingId] = true
	}

	if len(trackingIDs) != iterations {
		t.Errorf("Expected %d unique tracking IDs, got %d", iterations, len(trackingIDs))
	}
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	s := server{}

	res, err := s.Check(context.Background(), nil)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	if res.Status.String() != "SERVING" {
		t.Errorf("Expected SERVING status, got %s", res.Status.String())
	}
}

// TestGetQuoteNilAddress tests error handling for nil address
func TestGetQuoteNilAddress(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: nil,
		Items: []*pb.CartItem{
			{
				ProductId: "test",
				Quantity:  1,
			},
		},
	}

	// This should not panic, even with nil address
	_, err := s.GetQuote(context.Background(), req)
	if err != nil {
		// It's ok if it returns an error for nil address
		t.Logf("GetQuote with nil address returned error (expected): %v", err)
	}
}
