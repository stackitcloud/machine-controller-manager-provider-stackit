#!/usr/bin/env bash

set -euo pipefail

# Build configuration
BINARY_NAME="${BINARY_NAME:-machine-controller}"
OUTPUT_DIR="${OUTPUT_DIR:-build}"
OUTPUT_PATH="${OUTPUT_DIR}/${BINARY_NAME}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building ${BINARY_NAME}...${NC}"
echo "  Output: ${OUTPUT_PATH}"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Build with CGO disabled for static binary (required for Alpine)
CGO_ENABLED=0 GO111MODULE=on go build \
    -mod=vendor \
    -o "${OUTPUT_PATH}" \
    cmd/machine-controller/main.go

echo -e "${GREEN}✓ Build complete: ${OUTPUT_PATH}${NC}"
