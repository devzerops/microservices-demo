// Copyright 2018 Google LLC
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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/genproto"
	"github.com/sirupsen/logrus"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "healthy" or "unhealthy"
	Error  string `json:"error,omitempty"`
}

// HealthResponse represents the overall health status
type HealthResponse struct {
	Status string        `json:"status"` // "healthy" or "unhealthy"
	Checks []HealthCheck `json:"checks"`
}

// livenessHandler handles liveness probe - returns 200 if the service is running
// This is a simple check that the process is alive and can serve requests.
// Kubernetes will restart the pod if this check fails.
func (fe *frontendServer) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

// readinessHandler handles readiness probe - returns 200 only if all dependencies are available
// This checks if the service is ready to accept traffic by verifying critical dependencies.
// Kubernetes will not route traffic to the pod if this check fails.
func (fe *frontendServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)

	// Define critical dependencies to check
	checks := []struct {
		name  string
		check func(context.Context) error
	}{
		{"product_catalog", fe.checkProductCatalog},
		{"currency", fe.checkCurrency},
		{"cart", fe.checkCart},
	}

	response := HealthResponse{
		Status: "healthy",
		Checks: make([]HealthCheck, 0, len(checks)),
	}

	// Run all health checks
	for _, c := range checks {
		healthCheck := HealthCheck{
			Name:   c.name,
			Status: "healthy",
		}

		if err := c.check(ctx); err != nil {
			healthCheck.Status = "unhealthy"
			healthCheck.Error = err.Error()
			response.Status = "unhealthy"

			log.WithFields(logrus.Fields{
				"service": c.name,
				"error":   err,
			}).Warn("health check failed")
		}

		response.Checks = append(response.Checks, healthCheck)
	}

	// Set response status code
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkProductCatalog verifies the product catalog service is accessible
func (fe *frontendServer) checkProductCatalog(ctx context.Context) error {
	if fe.productCatalogSvcConn == nil {
		return fmt.Errorf("product catalog connection not initialized")
	}

	client := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn)

	// Try to list products with a short timeout
	_, err := client.ListProducts(ctx, &pb.Empty{})
	if err != nil {
		return fmt.Errorf("product catalog unavailable: %w", err)
	}

	return nil
}

// checkCurrency verifies the currency service is accessible
func (fe *frontendServer) checkCurrency(ctx context.Context) error {
	if fe.currencySvcConn == nil {
		return fmt.Errorf("currency connection not initialized")
	}

	client := pb.NewCurrencyServiceClient(fe.currencySvcConn)

	// Try to get supported currencies
	_, err := client.GetSupportedCurrencies(ctx, &pb.Empty{})
	if err != nil {
		return fmt.Errorf("currency service unavailable: %w", err)
	}

	return nil
}

// checkCart verifies the cart service is accessible
func (fe *frontendServer) checkCart(ctx context.Context) error {
	if fe.cartSvcConn == nil {
		return fmt.Errorf("cart connection not initialized")
	}

	client := pb.NewCartServiceClient(fe.cartSvcConn)

	// Try to get cart for a test user (this should not fail even if cart is empty)
	_, err := client.GetCart(ctx, &pb.GetCartRequest{UserId: "health-check"})
	if err != nil {
		return fmt.Errorf("cart service unavailable: %w", err)
	}

	return nil
}
