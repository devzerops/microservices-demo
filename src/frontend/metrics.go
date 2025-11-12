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
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics for business-level monitoring
// Note: Istio provides network-level metrics (request count, latency, errors)
// These metrics track application-specific business logic and security events
var (
	// HTTP request metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_http_requests_total",
			Help: "Total number of HTTP requests by method, endpoint, and status",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "frontend_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Rate limiting metrics
	rateLimitExceeded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_rate_limit_exceeded_total",
			Help: "Total number of rate limit violations by endpoint type",
		},
		[]string{"endpoint_type"}, // "ai", "post", "get"
	)

	rateLimitAllowed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_rate_limit_allowed_total",
			Help: "Total number of allowed requests by endpoint type",
		},
		[]string{"endpoint_type"},
	)

	// CSRF protection metrics
	csrfValidationFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_csrf_validation_failures_total",
			Help: "Total number of CSRF validation failures",
		},
	)

	csrfValidationSuccesses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_csrf_validation_successes_total",
			Help: "Total number of successful CSRF validations",
		},
	)

	// Session metrics
	activeSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "frontend_active_sessions",
			Help: "Current number of active sessions (estimated)",
		},
	)

	// Cart operations
	cartOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_cart_operations_total",
			Help: "Total number of cart operations",
		},
		[]string{"operation", "status"}, // operation: "add", "empty", "checkout"
	)

	// Product views
	productViews = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_product_views_total",
			Help: "Total number of product page views",
		},
		[]string{"product_id"},
	)

	// Checkout metrics
	checkoutAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_checkout_attempts_total",
			Help: "Total number of checkout attempts",
		},
	)

	checkoutSuccesses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_checkout_successes_total",
			Help: "Total number of successful checkouts",
		},
	)

	checkoutFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_checkout_failures_total",
			Help: "Total number of failed checkouts",
		},
		[]string{"reason"},
	)

	// gRPC client metrics
	grpcClientRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "frontend_grpc_client_requests_total",
			Help: "Total number of gRPC client requests",
		},
		[]string{"service", "method", "status"},
	)

	grpcClientDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "frontend_grpc_client_duration_seconds",
			Help:    "gRPC client request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
)

// initMetrics registers all Prometheus metrics
func initMetrics() {
	// Register HTTP metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)

	// Register rate limiting metrics
	prometheus.MustRegister(rateLimitExceeded)
	prometheus.MustRegister(rateLimitAllowed)

	// Register CSRF metrics
	prometheus.MustRegister(csrfValidationFailures)
	prometheus.MustRegister(csrfValidationSuccesses)

	// Register session metrics
	prometheus.MustRegister(activeSessions)

	// Register cart metrics
	prometheus.MustRegister(cartOperations)

	// Register product metrics
	prometheus.MustRegister(productViews)

	// Register checkout metrics
	prometheus.MustRegister(checkoutAttempts)
	prometheus.MustRegister(checkoutSuccesses)
	prometheus.MustRegister(checkoutFailures)

	// Register gRPC metrics
	prometheus.MustRegister(grpcClientRequests)
	prometheus.MustRegister(grpcClientDuration)
}

// metricsHandler returns the Prometheus metrics HTTP handler
func metricsHandler() http.Handler {
	return promhttp.Handler()
}
