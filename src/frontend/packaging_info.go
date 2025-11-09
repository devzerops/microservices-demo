// Copyright 2023 Google LLC
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

/*
As part of an optional Google Cloud demo, you can run an additional "packaging" microservice (HTTP server).
This file contains code related to the frontend and the "packaging" microservice.
*/

var (
	packagingServiceUrl string
	// HTTP client with timeout to prevent resource exhaustion
	packagingHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type PackagingInfo struct {
	Weight float32 `json:"weight"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
	Depth  float32 `json:"depth"`
}

// init() is a special function in Golang that will run when this package is imported.
func init() {
	packagingServiceUrl = os.Getenv("PACKAGING_SERVICE_URL")
}

func isPackagingServiceConfigured() bool {
	return packagingServiceUrl != ""
}

// validateProductId validates that a product ID contains only safe characters
// to prevent SSRF and path traversal attacks
func validateProductId(productId string) error {
	if productId == "" {
		return fmt.Errorf("product ID cannot be empty")
	}

	// Product IDs should only contain alphanumeric characters and hyphens
	// This prevents path traversal (../) and URL manipulation attacks
	validProductId := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validProductId.MatchString(productId) {
		return fmt.Errorf("invalid product ID: must contain only alphanumeric characters and hyphens")
	}

	// Additional length check to prevent excessively long IDs
	if len(productId) > 64 {
		return fmt.Errorf("product ID too long: maximum 64 characters")
	}

	return nil
}

func httpGetPackagingInfo(productId string) (*PackagingInfo, error) {
	// Validate product ID to prevent SSRF attacks
	if err := validateProductId(productId); err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Construct URL safely using url.JoinPath to prevent path traversal
	baseURL, err := url.Parse(packagingServiceUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid packaging service URL: %w", err)
	}

	// Use url.JoinPath to safely combine base URL with product ID
	fullURL := baseURL.JoinPath(productId).String()

	// Validate that the final URL still points to the expected host
	finalURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("constructed URL is invalid: %w", err)
	}
	if finalURL.Host != baseURL.Host {
		return nil, fmt.Errorf("URL host mismatch: potential SSRF attack")
	}

	logrus.WithField("url", fullURL).Debug("Requesting packaging info from URL")

	// Use HTTP client with timeout
	resp, err := packagingHTTPClient.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	// Read the JSON response body (using io.ReadAll instead of deprecated ioutil.ReadAll)
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response into a PackagingInfo struct
	var packagingInfo PackagingInfo
	err = json.Unmarshal(responseBody, &packagingInfo)
	if err != nil {
		return nil, err
	}

	return &packagingInfo, nil
}
