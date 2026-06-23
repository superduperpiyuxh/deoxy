# deoxy — Multi-language documentation comment generator
# Build, test, lint, and clean targets.
#
# Cross-compilation note:
#   tree-sitter requires CGO. For native builds (CGO_ENABLED=1), the host C
#   toolchain is used. For cross-compilation, set CC to the appropriate
#   cross-compiler for the target platform:
#
#     make cross-linux-arm64   CROSS_CC=aarch64-linux-gnu-gcc
#     make cross-windows-amd64 CROSS_CC=x86_64-w64-mingw32-gcc
#     make cross-windows-arm64 CROSS_CC=aarch64-w64-mingw32-gcc
#
#   macOS→macOS cross from Linux is not supported by Go's CGO toolchain.
#   Build natively on macOS for darwin targets.

BINARY    ?= deoxy
VERSION   ?= 0.1.0
GO        ?= go
CROSS_DIR ?= build
CROSS_CC  ?=

.PHONY: all build test lint clean
.PHONY: cross cross-linux-amd64 cross-linux-arm64
.PHONY: cross-darwin-amd64 cross-darwin-arm64
.PHONY: cross-windows-amd64 cross-windows-arm64

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
	rm -rf $(CROSS_DIR)

# --- Cross-compilation targets ---
# Each target respects CROSS_CC for the C cross-compiler.
# Unset CROSS_CC → builds with CGO_ENABLED=0 (tree-sitter won't be available).
# Example: make cross-linux-arm64 CROSS_CC=aarch64-linux-gnu-gcc

cross-linux-amd64:
	GOOS=linux GOARCH=amd64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-linux-amd64 ./cmd/deoxy

cross-linux-arm64:
	GOOS=linux GOARCH=arm64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-linux-arm64 ./cmd/deoxy

cross-darwin-amd64:
	GOOS=darwin GOARCH=amd64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-darwin-amd64 ./cmd/deoxy

cross-darwin-arm64:
	GOOS=darwin GOARCH=arm64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-darwin-arm64 ./cmd/deoxy

cross-windows-amd64:
	GOOS=windows GOARCH=amd64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-windows-amd64.exe ./cmd/deoxy

cross-windows-arm64:
	GOOS=windows GOARCH=arm64 \
	CGO_ENABLED=$(if $(CROSS_CC),1,0) CC=$(CROSS_CC) \
	$(GO) build -ldflags "-X main.version=$(VERSION)" \
		-o $(CROSS_DIR)/deoxy-windows-arm64.exe ./cmd/deoxy

cross: cross-linux-amd64 cross-linux-arm64 cross-darwin-amd64 cross-darwin-arm64 \
	cross-windows-amd64 cross-windows-arm64
