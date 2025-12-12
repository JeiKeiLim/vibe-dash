.PHONY: build test test-all lint fmt check-fmt run clean install test-accuracy

# CGO_ENABLED=1 required for go-sqlite3
build:
	CGO_ENABLED=1 go build -o bin/vibe ./cmd/vibe

test:
	go test ./...

test-all:
	go test -tags=integration ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run

fmt:
	goimports -w .

check-fmt:
	@test -z "$$(goimports -l .)" || (echo "Run 'make fmt' to fix formatting" && exit 1)

run: build
	./bin/vibe

clean:
	rm -rf bin/

install:
	CGO_ENABLED=1 go install ./cmd/vibe

# Placeholder for detection accuracy testing (95% threshold)
test-accuracy:
	@echo "Detection accuracy tests not yet implemented"
