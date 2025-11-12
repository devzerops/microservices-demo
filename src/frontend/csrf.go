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
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	csrfCookieName = cookiePrefix + "csrf-token"
	csrfTokenLength = 32
)

type ctxKeyCSRFToken struct{}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// csrfProtection is a middleware that implements CSRF protection
func csrfProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get existing token from cookie, or generate a new one
		var token string
		if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
			token = cookie.Value
		} else {
			newToken, err := generateCSRFToken()
			if err != nil {
				http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
				return
			}
			token = newToken

			// Set the CSRF token cookie
			// Use SameSite=Strict for maximum protection
			// HttpOnly=false because we need JavaScript access for AJAX requests
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				MaxAge:   cookieMaxAge,
				HttpOnly: false, // Allow JavaScript access for AJAX
				Secure:   false,  // Set to true in production with HTTPS
				SameSite: http.SameSiteStrictMode,
			})
		}

		// Add token to context for template access
		ctx := context.WithValue(r.Context(), ctxKeyCSRFToken{}, token)
		r = r.WithContext(ctx)

		// For POST requests, validate the CSRF token
		if r.Method == http.MethodPost {
			var submittedToken string

			// Check for token in form field first
			if err := r.ParseForm(); err == nil {
				submittedToken = r.FormValue("csrf_token")
			}

			// If not in form, check for token in custom header (for AJAX requests)
			if submittedToken == "" {
				submittedToken = r.Header.Get("X-CSRF-Token")
			}

			// Validate token
			if submittedToken == "" || !strings.EqualFold(submittedToken, token) {
				// Record CSRF validation failure
				csrfValidationFailures.Inc()

				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}

			// Record successful CSRF validation
			csrfValidationSuccesses.Inc()
		}

		next.ServeHTTP(w, r)
	})
}
