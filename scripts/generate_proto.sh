#!/bin/bash

set -e

echo "Generating protobuf files..."

if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install it first."
    echo "Visit: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

mkdir -p internal/api

echo "Generating auth.proto..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/api/auth.proto

echo "Generating metadata.proto..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/api/metadata.proto

echo "Generating file.proto..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/api/file.proto

echo "Protobuf files generated successfully!"