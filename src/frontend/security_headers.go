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
	"net/http"
	"os"
)

// securityHeadersMiddleware adds standard HTTP security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Frame-Options: Prevents clickjacking attacks
		// DENY: Page cannot be displayed in a frame
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevents MIME type sniffing
		// nosniff: Browser will not MIME-sniff away from declared content-type
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Enables browser's XSS filter
		// 1; mode=block: Enable XSS filter and block rendering if attack detected
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Controls how much referrer information is sent
		// strict-origin-when-cross-origin: Send origin only when protocol security level stays same
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: Helps prevent XSS, clickjacking, and other code injection attacks
		// This is a relaxed policy suitable for demos - tighten for production
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://stackpath.bootstrapcdn.com https://fonts.googleapis.com; " +
			"style-src 'self' 'unsafe-inline' https://stackpath.bootstrapcdn.com https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https://storage.googleapis.com https://googleusercontent.com; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		w.Header().Set("Content-Security-Policy", csp)

		// Permissions-Policy: Controls which browser features can be used
		// Disable potentially dangerous features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Strict-Transport-Security (HSTS): Enforces HTTPS connections
		// Only set if HTTPS is enabled (not set for local development by default)
		if os.Getenv("ENABLE_HTTPS") == "true" || os.Getenv("ENABLE_HTTPS") == "1" {
			// max-age=31536000: 1 year
			// includeSubDomains: Apply to all subdomains
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
