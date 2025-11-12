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
import random
import time
import traceback
from concurrent import futures

import googlecloudprofiler
from google.auth.exceptions import DefaultCredentialsError
import grpc

import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

from opentelemetry import trace
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient, GrpcInstrumentorServer
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

from logger import getJSONLogger
logger = getJSONLogger('recommendationservice-server')

def initStackdriverProfiling():
  project_id = None
  try:
    project_id = os.environ["GCP_PROJECT_ID"]
  except KeyError:
    # Environment variable not set
    pass

  for retry in range(1,4):
    try:
      if project_id:
        googlecloudprofiler.start(service='recommendation_server', service_version='1.0.0', verbose=0, project_id=project_id)
      else:
        googlecloudprofiler.start(service='recommendation_server', service_version='1.0.0', verbose=0)
      logger.info("Successfully started Stackdriver Profiler.")
      return
    except (BaseException) as exc:
      logger.info("Unable to start Stackdriver Profiler Python agent. " + str(exc))
      if (retry < 4):
        logger.info("Sleeping %d seconds to retry Stackdriver Profiler agent initialization"%(retry*10))
        time.sleep (1)
      else:
        logger.warning("Could not initialize Stackdriver Profiler after retrying, giving up")
  return

class RecommendationService(demo_pb2_grpc.RecommendationServiceServicer):
    def ListRecommendations(self, request, context):
        max_responses = 5
        # fetch list of products from product catalog stub
        cat_response = product_catalog_stub.ListProducts(demo_pb2.Empty())
        product_ids = [x.id for x in cat_response.products]
        filtered_products = list(set(product_ids)-set(request.product_ids))
        num_products = len(filtered_products)
        num_return = min(max_responses, num_products)
        # sample list of indicies to return
        indices = random.sample(range(num_products), num_return)
        # fetch product ids from indices
        prod_list = [filtered_products[i] for i in indices]
        logger.info("[Recv ListRecommendations] product_ids={}".format(prod_list))
        # build and return response
        response = demo_pb2.ListRecommendationsResponse()
        response.product_ids.extend(prod_list)
        return response

    def Check(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.SERVING)

    def Watch(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)


if __name__ == "__main__":
    logger.info("initializing recommendationservice")

    try:
      if "DISABLE_PROFILER" in os.environ:
        raise KeyError()
      else:
        logger.info("Profiler enabled.")
        initStackdriverProfiling()
    except KeyError:
        logger.info("Profiler disabled.")

    try:
      grpc_client_instrumentor = GrpcInstrumentorClient()
      grpc_client_instrumentor.instrument()
      grpc_server_instrumentor = GrpcInstrumentorServer()
      grpc_server_instrumentor.instrument()
      if os.environ["ENABLE_TRACING"] == "1":
        trace.set_tracer_provider(TracerProvider())
        otel_endpoint = os.getenv("COLLECTOR_SERVICE_ADDR", "localhost:4317")
        trace.get_tracer_provider().add_span_processor(
          BatchSpanProcessor(
              OTLPSpanExporter(
              endpoint = otel_endpoint
            )
          )
        )
    except (KeyError, DefaultCredentialsError):
        logger.info("Tracing disabled.")
    except Exception as e:
        logger.warn(f"Exception on Cloud Trace setup: {traceback.format_exc()}, tracing disabled.")

    port_str = os.environ.get('PORT', "8080")
    try:
        port = int(port_str)
        if port < 1 or port > 65535:
            logger.error(f"Invalid PORT value: {port}. Must be between 1 and 65535.")
            sys.exit(1)
    except ValueError:
        logger.error(f"Invalid PORT value: {port_str}. Must be a number.")
        sys.exit(1)

    catalog_addr = os.environ.get('PRODUCT_CATALOG_SERVICE_ADDR', '')
    if catalog_addr == "":
        raise Exception('PRODUCT_CATALOG_SERVICE_ADDR environment variable not set')
    logger.info("product catalog address: " + catalog_addr)

    # Configure TLS based on ENABLE_GRPC_TLS environment variable
    # Supports: "true"/"system" (TLS with system CA), "skip-verify" (TLS without verification),
    # "custom" (TLS with custom CA from GRPC_TLS_CA_CERT), or empty/"false" (insecure)
    tls_mode = os.environ.get('ENABLE_GRPC_TLS', '')

    if tls_mode in ['true', 'system']:
        logger.info("Using TLS with system CA certificates for gRPC connections")
        credentials = grpc.ssl_channel_credentials()
        channel = grpc.secure_channel(catalog_addr, credentials)
    elif tls_mode == 'skip-verify':
        logger.warning("Using TLS with certificate verification DISABLED - DO NOT use in production")
        # Create credentials that skip verification (for testing only)
        credentials = grpc.ssl_channel_credentials()
        # Note: Python gRPC doesn't have a direct "skip verify" option
        # Use system certs but be aware this is less secure
        channel = grpc.secure_channel(catalog_addr, credentials)
    elif tls_mode == 'custom':
        ca_cert_file = os.environ.get('GRPC_TLS_CA_CERT', '')
        if not ca_cert_file:
            raise Exception('GRPC_TLS_CA_CERT must be set when ENABLE_GRPC_TLS=custom')
        logger.info(f"Using TLS with custom CA certificate from {ca_cert_file}")
        with open(ca_cert_file, 'rb') as f:
            ca_cert = f.read()
        credentials = grpc.ssl_channel_credentials(root_certificates=ca_cert)
        channel = grpc.secure_channel(catalog_addr, credentials)
    else:
        if tls_mode and tls_mode != 'false':
            logger.warning(f"Unknown ENABLE_GRPC_TLS value '{tls_mode}', using insecure connection")
        logger.info("Using insecure gRPC connections (no TLS)")
        channel = grpc.insecure_channel(catalog_addr)

    product_catalog_stub = demo_pb2_grpc.ProductCatalogServiceStub(channel)

    # create gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    # add class to gRPC server
    service = RecommendationService()
    demo_pb2_grpc.add_RecommendationServiceServicer_to_server(service, server)
    health_pb2_grpc.add_HealthServicer_to_server(service, server)

    # start server
    logger.info(f"listening on port: {port}")
    server.add_insecure_port(f'[::]:{port}')
    server.start()

    # keep alive
    try:
         while True:
            time.sleep(10000)
    except KeyboardInterrupt:
            server.stop(0)
