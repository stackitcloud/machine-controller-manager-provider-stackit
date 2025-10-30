# Go module just recipes
#
# This file provides common Go (golang) development tasks as just recipes.
#
# Usage examples:
#   mod? go 'vendir/just-pantry/modules/golang/module.just'
#   just go::format
#   just go::test
#
# For more information, see the project documentation, the just manual (https://just.systems/man/en/modules1190.html), or run `just --help`.

import '../../lib/module.just'
import 'vars.just'

[group('golang')]
[no-cd]
info:
    #!/usr/bin/env bash
    echo "Golang PATH: {{ go-path }}"
    echo "Golang GOBIN: {{ go-bin }}"
    echo "Golang Module: {{ go-mod }}"
    echo "Golang Module Path: {{ go-mod-path }}"
    echo "Golang Module Short: {{ go-mod-short }}"

# Formats golang code recursively
[group('golang')]
[no-cd]
format:
    go fmt ./...

# Runs go vet for static analysis
[group('golang')]
[no-cd]
vet:
    go vet ./...

# Runs all tests
[group('golang')]
[no-cd]
test:
    go test ./...

# Runs tests with verbose output
[group('golang')]
[no-cd]
test-verbose:
    go test -v ./...

# Runs tests with race detector
[group('golang')]
[no-cd]
test-race:
    go test -race ./...

# Runs tests and generates coverage report
[group('golang')]
[no-cd]
test-cover:
    go test -cover ./...

# Updates dependencies
[group('golang')]
[no-cd]
mod-tidy:
    go mod tidy

# Downloads dependencies
[group('golang')]
[no-cd]
mod-download:
    go mod download

# Generates vendor directory with dependencies
[group('golang')]
[no-cd]
mod-vendor:
    go mod vendor

# Runs golangci-lint with default settings
[group('golang')]
[no-cd]
lint: (requires 'golangci-lint')
    golangci-lint run ./...

# Runs golangci-lint with default settings
[group('golang')]
[no-cd]
lint-fix: (requires 'golangci-lint')
    golangci-lint run --fix ./...

# Verify golangci-lint linter configuration
[group('golang')]
[no-cd]
lint-config-verify: (requires 'golangci-lint')
    golangci-lint config verify

# Runs golangci-lint with all issues shown
[group('golang')]
[no-cd]
lint-issues: (requires 'golangci-lint')
    golangci-lint run --issues-exit-code=0 --out-format=tab ./...
