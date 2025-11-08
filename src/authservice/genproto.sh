#!/bin/bash -eu

cd "$(dirname "$0")"

# Generate Go protobuf code
protoc -I ../../protos \
  --go_out=genproto \
  --go_opt=paths=source_relative \
  --go-grpc_out=genproto \
  --go-grpc_opt=paths=source_relative \
  ../../protos/demo.proto
