#!/usr/bin/env bash
# Build binary inside Docker container (no local Go required)

set -euo pipefail

OUTPUT_DIR="${OUTPUT_DIR:-build}"

echo "Building binary in Docker container..."

docker build \
    --file scripts/Dockerfile_build \
    --output "${OUTPUT_DIR}" \
    .

echo "✓ Binary built via Docker: ${OUTPUT_DIR}/machine-controller"
