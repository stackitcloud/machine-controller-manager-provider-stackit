# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
# SPDX-License-Identifier: Apache-2.0

#############      builder                                  #############
FROM golang:1.25.3 AS builder

WORKDIR /workspace
COPY . .

# Build binary
RUN ./scripts/build.sh

#############      base                                     #############
FROM alpine:3.20 AS base

RUN apk add --no-cache bash curl tzdata ca-certificates
WORKDIR /

#############      machine-controller                       #############
FROM base AS machine-controller

COPY --from=builder /workspace/build/machine-controller /machine-controller

USER 65532:65532
ENTRYPOINT ["/machine-controller"]
