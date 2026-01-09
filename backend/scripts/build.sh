#!/bin/bash
# Build script for Omnipoll

set -e

echo "Building Omnipoll backend..."

# Build frontend first
cd ../frontend
npm run build
cp -r dist ../backend/web/

# Build Go binary
cd ../backend
go build -o bin/omnipoll ./cmd/omnipoll

echo "Build complete!"
