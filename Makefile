# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
# SPDX-License-Identifier: Apache-2.0

# Minimal Makefile wrapper - all logic is in justfile
# This provides backward compatibility for make users

.PHONY: build
build:
	just build

.PHONY: build-local
build-local:
	just build

.PHONY: start
start:
	just start

.PHONY: clean
clean:
	just clean

.PHONY: revendor
revendor:
	just revendor

.PHONY: update-dependencies
update-dependencies:
	just update-deps

.PHONY: test-unit
test-unit:
	just golang::test

.PHONY: docker-image
docker-image:
	just docker-build

.PHONY: docker-push
docker-push:
	just docker-push

.PHONY: lint
lint:
	just lint

.PHONY: fmt
fmt:
	just fmt
