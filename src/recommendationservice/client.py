#!/usr/bin/python
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import sys
import grpc
import demo_pb2
import demo_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('recommendationservice-server')

if __name__ == "__main__":
    # get port
    if len(sys.argv) > 1:
        port = sys.argv[1]
    else:
        port = "8080"

    # Configure TLS based on ENABLE_GRPC_TLS environment variable
    # For test client: supports "true"/"system" (TLS), "custom" (custom CA), or empty/"false" (insecure)
    tls_mode = os.environ.get('ENABLE_GRPC_TLS', '')
    addr = 'localhost:' + port

    if tls_mode in ['true', 'system']:
        logger.info("Using TLS with system CA certificates")
        credentials = grpc.ssl_channel_credentials()
        channel = grpc.secure_channel(addr, credentials)
    elif tls_mode == 'custom':
        ca_cert_file = os.environ.get('GRPC_TLS_CA_CERT', '')
        if ca_cert_file:
            logger.info(f"Using TLS with custom CA certificate from {ca_cert_file}")
            with open(ca_cert_file, 'rb') as f:
                ca_cert = f.read()
            credentials = grpc.ssl_channel_credentials(root_certificates=ca_cert)
            channel = grpc.secure_channel(addr, credentials)
        else:
            logger.warning("GRPC_TLS_CA_CERT not set, falling back to insecure channel")
            channel = grpc.insecure_channel(addr)
    else:
        logger.info("Using insecure gRPC connection (no TLS)")
        channel = grpc.insecure_channel(addr)

    stub = demo_pb2_grpc.RecommendationServiceStub(channel)
    # form request
    request = demo_pb2.ListRecommendationsRequest(user_id="test", product_ids=["test"])
    # make call to server
    response = stub.ListRecommendations(request)
    logger.info(response)
