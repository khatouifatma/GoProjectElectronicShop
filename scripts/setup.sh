#!/bin/bash
# Quick setup script for Electronic Shop API
set -e
echo "Downloading Go dependencies..."
go mod download && go mod tidy
echo "Done! Run: go run cmd/main.go"
