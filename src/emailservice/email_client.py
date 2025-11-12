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
import grpc

import demo_pb2
import demo_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('emailservice-client')

def send_confirmation_email(email, order):
  # Configure TLS based on ENABLE_GRPC_TLS environment variable
  # For test client: supports "true"/"system" (TLS), "custom" (custom CA), or empty/"false" (insecure)
  tls_mode = os.environ.get('ENABLE_GRPC_TLS', '')
  addr = '[::]:8080'

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

  stub = demo_pb2_grpc.EmailServiceStub(channel)
  try:
    response = stub.SendOrderConfirmation(demo_pb2.SendOrderConfirmationRequest(
      email = email,
      order = order
    ))
    logger.info('Request sent.')
  except grpc.RpcError as err:
    logger.error(err.details())
    logger.error('{}, {}'.format(err.code().name, err.code().value))

if __name__ == '__main__':
  logger.info('Client for email service.')
