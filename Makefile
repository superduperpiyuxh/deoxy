# deoxy — Multi-language documentation comment generator
# Build, test, lint, and clean targets.

BINARY  ?= deoxy
VERSION ?= 0.1.0
GO      ?= go

.PHONY: all build test lint clean

all: build test lint

build:
	CGO_ENABLED=1 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/deoxy

build-race:
	CGO_ENABLED=1 $(GO) build -race -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/deoxy

test:
	$(GO) test ./... -v -count=1 -race

lint:
	golangci-lint run ./... || $(GO) vet ./...

clean:
	rm -f $(BINARY) $(BINARY)*.out coverage.out
