# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
# SPDX-License-Identifier: Apache-2.0

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
SOURCES := Makefile go.mod go.sum $(shell find $(DEST) -name '*.go' 2>/dev/null)
VERSION ?= $(shell git describe --dirty --tags --match='v*' 2>/dev/null || git rev-parse --short HEAD)
REGISTRY ?= ghcr.io
REPO ?= stackitcloud/machine-controller-manager-provider-stackit
PUSH ?= false
PLATFORMS ?= amd64 arm64
IS_DEV ?= true

include ./hack/tools.mk

.PHONY: image
image: $(KO) ## Builds a single binary specified by TARGET
	KO_DOCKER_REPO=$(REGISTRY)/$(REPO) \
	$(KO) build --push=$(PUSH) \
	--image-label org.opencontainers.image.source="https://github.com/stackitcloud/machine-controller-manager-provider-stackit" \
	--sbom none -t $(VERSION) \
	--bare \
	--platform linux/amd64,linux/arm64 \
	./cmd/machine-controller

.PHONY: clean-tools-bin
clean-tools-bin: ## Empty the tools binary directory.
	rm -rf $(TOOLS_BIN_DIR)/* $(TOOLS_BIN_DIR)/.version_*

.PHONY: fmt
fmt: $(GOIMPORTS_REVISER) ## Run go fmt against code.
	go fmt ./...
	$(GOIMPORTS_REVISER) .

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	go mod tidy

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run golangci-lint against code.
	$(GOLANGCI_LINT) run ./...

.PHONY: check
check: lint test ## Check everything (lint + test).

.PHONY: test
test: ## Run tests.
	./hack/test.sh ./cmd/... ./pkg/...