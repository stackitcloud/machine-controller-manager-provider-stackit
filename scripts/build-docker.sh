#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
# SPDX-License-Identifier: Apache-2.0
#
# Build binary inside Docker container (no local Go required)

set -euo pipefail

OUTPUT_DIR="${OUTPUT_DIR:-build}"

echo "Building binary in Docker container..."

docker build \
    --file scripts/Dockerfile_build \
    --output "${OUTPUT_DIR}" \
    .

echo "✓ Binary built via Docker: ${OUTPUT_DIR}/machine-controller"
