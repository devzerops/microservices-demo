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

// Package profiling provides common profiling utilities for Go microservices
package profiling

import (
	"time"

	"cloud.google.com/go/profiler"
	"github.com/sirupsen/logrus"
)

// InitProfiling initializes Google Cloud Profiler with retry logic.
// This function attempts to start the profiler up to 3 times with exponential backoff.
//
// Parameters:
//   - log: A logrus field logger for logging profiling status
//   - service: The name of the service (e.g., "shippingservice")
//   - version: The version of the service (e.g., "1.0.0")
//   - projectID: Optional GCP project ID. If empty, auto-detected from environment
func InitProfiling(log logrus.FieldLogger, service, version, projectID string) {
	for i := 1; i <= 3; i++ {
		config := profiler.Config{
			Service:        service,
			ServiceVersion: version,
		}

		// Set ProjectID if provided
		if projectID != "" {
			config.ProjectID = projectID
		}

		if err := profiler.Start(config); err != nil {
			log.Warnf("failed to start profiler (attempt %d/3): %+v", i, err)
		} else {
			log.Info("started Stackdriver profiler")
			return
		}

		// Exponential backoff between retries
		d := time.Second * 10 * time.Duration(i)
		log.Infof("sleeping %v to retry initializing Stackdriver profiler", d)
		time.Sleep(d)
	}
	log.Warn("could not initialize Stackdriver profiler after retrying, giving up")
}

// InitProfilingSimple initializes profiler with service name and version only.
// Uses auto-detection for GCP project ID.
func InitProfilingSimple(log logrus.FieldLogger, service, version string) {
	InitProfiling(log, service, version, "")
}
