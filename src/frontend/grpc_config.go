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
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// getGRPCDialOptions returns appropriate gRPC dial options based on TLS configuration
// Supports three modes via ENABLE_GRPC_TLS environment variable:
// - "true" or "system": Use system CA certificates for TLS
// - "skip-verify": Use TLS but skip certificate verification (for testing)
// - empty or "false": Use insecure connection (backward compatible)
func getGRPCDialOptions(log logrus.FieldLogger) ([]grpc.DialOption, error) {
	tlsMode := os.Getenv("ENABLE_GRPC_TLS")

	var opts []grpc.DialOption

	switch tlsMode {
	case "true", "system":
		// Use system CA certificates for TLS verification
		log.Info("Using TLS with system CA certificates for gRPC connections")

		// Load system CA certificates
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load system CA certificates")
		}

		// Create TLS credentials with system CA pool
		tlsConfig := &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12, // Enforce minimum TLS 1.2
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))

	case "skip-verify":
		// Use TLS but skip certificate verification (for testing/development only)
		log.Warn("Using TLS with certificate verification DISABLED - DO NOT use in production")
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))

	case "custom":
		// Load custom CA certificate from file specified in GRPC_TLS_CA_CERT
		certFile := os.Getenv("GRPC_TLS_CA_CERT")
		if certFile == "" {
			return nil, errors.New("GRPC_TLS_CA_CERT must be set when ENABLE_GRPC_TLS=custom")
		}

		log.Infof("Using TLS with custom CA certificate from %s", certFile)

		// Read custom CA certificate
		caCert, err := os.ReadFile(certFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read CA certificate from %s", certFile)
		}

		// Create cert pool with custom CA
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate to pool")
		}

		tlsConfig := &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))

	default:
		// Default to insecure for backward compatibility
		if tlsMode != "" && tlsMode != "false" {
			log.Warnf("Unknown ENABLE_GRPC_TLS value '%s', using insecure connection", tlsMode)
		}
		log.Info("Using insecure gRPC connections (no TLS)")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add OpenTelemetry interceptors
	opts = append(opts,
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))

	return opts, nil
}

// mustConnGRPCWithTLS establishes a gRPC connection with TLS support
func mustConnGRPCWithTLS(ctx context.Context, log logrus.FieldLogger, conn **grpc.ClientConn, addr string) {
	var err error
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	opts, err := getGRPCDialOptions(log)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to get dial options"))
	}

	*conn, err = grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", addr))
	}

	log.Infof("Successfully connected to gRPC service at %s", addr)
}
