# deoxy

**Multi-language documentation comment generator**

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](#)

deoxy is a deterministic, AI-free CLI tool that parses source code using [tree-sitter](https://tree-sitter.github.io/tree-sitter/) AST analysis and auto-generates properly-formatted doc comments. No cloud, no hallucinations — just correct, idiomatic documentation.

## Features

- **Multi-language parsing** — Supports Go, Python, C, C++, and Rust with more on the roadmap
- **AST-level analysis** — Extracts function signatures, parameters, return types, struct fields, and type parameters via tree-sitter queries
- **Idiomatic doc styles** — GoDoc prose, Python docstrings (Google style), Doxygen/JSDoc for C/C++, Rustdoc
- **Deterministic output** — Same input always produces the same output; no external API calls or AI
- **In-place injection** — Insert doc comments directly into source files with existing-comment detection
- **Customizable templates** — Configurable tag order, custom tags, and per-language docstyle via `.deoxy.yaml`
- **VS Code extension** — (Coming in Phase 4) Generate doc comments with a single command
- **Git-aware mode** — (Coming in Phase 5) Only process files changed since the last commit

## Installation

### Via `go install`

```bash
go install github.com/superduperpiyuxh/deoxy/cmd/deoxy@latest
```

### Prebuilt binaries

Download the latest release for your platform from the [Releases page](https://github.com/superduperpiyuxh/deoxy/releases). Binaries are available for:

| OS      | amd64 | arm64 |
| ------- | ----- | ----- |
| Linux   | ✓     | ✓     |
| macOS   | ✓     | ✓     |
| Windows | ✓     | ✓     |

### Build from source

```bash
git clone https://github.com/superduperpiyuxh/deoxy.git
cd deoxy
make build
./deoxy
```

## Quick Start

> **Note:** CLI commands (`generate`, `init`, `watch`) will be implemented in Phase 3. The current binary only prints the version.

```bash
# Verify installation
deoxy
# Output: deoxy v0.1.0
```

Usage examples will be added once the CLI is operational.

## Supported Languages

| Language | Parsing | Doc Generation | Status         |
| -------- | ------- | -------------- | -------------- |
| Go       | Planned | Planned        | Phase 1        |
| Python   | Planned | Planned        | Phase 1        |
| C        | Planned | Planned        | Phase 1        |
| C++      | Planned | Planned        | Phase 1        |
| Rust     | Planned | Planned        | Phase 1–2      |

## Development

### Prerequisites

- Go 1.26.3 or later
- Make (optional, raw `go` commands also work)

### Building

```bash
make build       # builds the deoxy binary
```

Or without Make:

```bash
go build -o deoxy ./cmd/deoxy
```

### Testing

```bash
make test        # runs all tests with race detection
```

Or:

```bash
go test ./... -v -count=1
```

### Linting

```bash
make lint        # runs golangci-lint or falls back to go vet
```

Or:

```bash
go vet ./...
```

### Project Structure

```
deoxy/
├── cmd/
│   └── deoxy/
│       └── main.go          # Entry point
├── internal/
│   ├── config/              # Configuration loading
│   ├── generator/           # Top-level orchestration
│   ├── lang/                # Language registry
│   └── parser/              # Tree-sitter AST parser
├── queries/
│   ├── go/                  # Tree-sitter queries for Go
│   ├── python/              # Tree-sitter queries for Python
│   ├── c/                   # Tree-sitter queries for C
│   ├── cpp/                 # Tree-sitter queries for C++
│   └── rust/                # Tree-sitter queries for Rust
├── .goreleaser.yaml         # Cross-compilation config
├── Makefile                 # Build, test, lint, clean
└── go.mod                   # Module definition
```

## Configuration

deoxy uses a `.deoxy.yaml` configuration file to customize behavior per project. A default config can be generated with `deoxy init` (available in Phase 3).

> **Note:** Configuration loading is planned for Phase 2.

### Example structure

```yaml
# .deoxy.yaml (placeholder — full schema coming in Phase 2)
languages:
  go:
    docstyle: godoc
  python:
    docstyle: google
```

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or pull request on [GitHub](https://github.com/superduperpiyuxh/deoxy).
