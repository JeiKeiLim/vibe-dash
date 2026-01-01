# SHELL must be bash for pipefail support
SHELL := /bin/bash

.PHONY: build test test-all test-behavioral lint fmt check-fmt run clean install test-accuracy

# Version information for ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Version=$(VERSION) \
                    -X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Commit=$(COMMIT) \
                    -X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.BuildDate=$(BUILD_DATE)"

# CGO_ENABLED=1 required for go-sqlite3
build:
	@set -o pipefail; \
	start=$$(date +%s); \
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/vibe ./cmd/vibe; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_build_summary $$exit_code $$((end-start)) bin/vibe $(VERSION); \
	exit $$exit_code

test:
	@set -o pipefail; \
	tmpfile="/tmp/vibe-test-output-$$$$.txt"; \
	start=$$(date +%s); \
	go test -v ./... 2>&1 | tee "$$tmpfile"; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_test_summary $$exit_code $$((end-start)) "$$tmpfile"; \
	exit $$exit_code

test-all:
	@set -o pipefail; \
	tmpfile="/tmp/vibe-test-output-$$$$.txt"; \
	start=$$(date +%s); \
	go test -v -tags=integration -timeout=10m ./... 2>&1 | tee "$$tmpfile"; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_test_summary $$exit_code $$((end-start)) "$$tmpfile"; \
	exit $$exit_code

# Behavioral tests only (for debugging TUI issues locally)
# Runs anchor, layout, and resource tests with verbose output
# Useful for isolating TUI-specific failures without running full suite
test-behavioral:
	go test -tags=integration -timeout=10m -v ./internal/adapters/tui/... -run 'TestAnchor_|TestLayout_|TestResource_'

lint:
	@start=$$(date +%s); \
	$(shell go env GOPATH)/bin/golangci-lint run; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_lint_summary $$exit_code $$((end-start)); \
	exit $$exit_code

fmt:
	$(shell go env GOPATH)/bin/goimports -w .

check-fmt:
	@test -z "$$($(shell go env GOPATH)/bin/goimports -l .)" || (echo "Run 'make fmt' to fix formatting" && exit 1)

run: build
	./bin/vibe

clean:
	rm -rf bin/

install:
	CGO_ENABLED=1 go install $(LDFLAGS) ./cmd/vibe

# Detection accuracy testing (95% threshold - launch blocker)
test-accuracy:
	@echo "Running detection accuracy tests (95% threshold)..."
	@go test -v -run TestDetectionAccuracy ./internal/adapters/detectors/... 2>&1 | tee /dev/stderr | grep -q "PASS" || (echo "FAILED: Detection accuracy below 95% threshold" && exit 1)
