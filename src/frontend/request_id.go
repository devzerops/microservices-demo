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
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

// Context key for request ID
type ctxKeyRequestID struct{}

// generateRequestID creates a new random request ID (16 bytes = 32 hex chars)
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a simple counter-based ID if random generation fails
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}

// requestIDMiddleware adds or extracts request ID for distributed tracing.
// Compatible with Istio's x-request-id header for seamless integration.
//
// When Istio is enabled:
//   - Istio automatically generates and propagates x-request-id across services
//   - This middleware extracts and logs the Istio-generated ID
//
// When Istio is disabled:
//   - This middleware generates a new request ID if none exists
//   - The ID is propagated to downstream services via HTTP and gRPC headers
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get request ID from incoming request
		// Istio uses x-request-id, which is the de facto standard
		requestID := r.Header.Get("x-request-id")

		// If no request ID exists, generate a new one
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Store request ID in context for use in handlers and logging
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, requestID)

		// Add request ID to response headers for client-side tracing
		w.Header().Set("X-Request-ID", requestID)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestID extracts the request ID from the context
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return requestID
	}
	return "unknown"
}

// addRequestIDToGRPCContext adds the request ID to gRPC metadata for propagation
// to downstream services. This ensures distributed tracing works across HTTP and gRPC boundaries.
func addRequestIDToGRPCContext(ctx context.Context) context.Context {
	requestID := getRequestID(ctx)
	if requestID != "unknown" {
		// Add to outgoing gRPC metadata
		md := metadata.Pairs("x-request-id", requestID)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// enrichLogWithRequestID adds request ID to log entries for correlation
func enrichLogWithRequestID(ctx context.Context, log logrus.FieldLogger) logrus.FieldLogger {
	requestID := getRequestID(ctx)
	return log.WithField("request_id", requestID)
}
