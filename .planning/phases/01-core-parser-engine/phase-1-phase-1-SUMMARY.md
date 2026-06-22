---
phase: 1
plan: phase-1
subsystem: Core Parser Engine
tags: [tree-sitter, parser, symbol-extraction, go, python, c, cpp, rust]
requires: [phase-0]
provides: [symbol-types, parser-pool, language-registry, query-runner, scanner]
affects: [internal/symbol, internal/parser, internal/scanner, queries]
tech-stack:
  added: [github.com/tree-sitter/go-tree-sitter v0.25.0, tree-sitter grammars for go/python/c/cpp/rust]
  patterns: [CGo with explicit Close(), parser pool per language, go:embed for query files]
key-files:
  created:
    - internal/symbol/symbol.go
    - internal/parser/parser.go
    - internal/parser/registry.go
    - internal/parser/queryrunner.go
    - internal/parser/parser_test.go
    - internal/parser/queries/*.scm
    - internal/scanner/scanner.go
    - internal/scanner/scanner_test.go
    - queries/*/docs.scm
    - testdata/fixtures/*/sample.*
  modified:
    - Makefile
    - go.mod
    - go.sum
decisions:
  - "Use official github.com/tree-sitter/go-tree-sitter v0.25+ (not smacker fork)"
  - "Grammar bindings imported from github.com/tree-sitter/tree-sitter-<lang>/bindings/go"
  - "Query files duplicated to internal/parser/queries/ for go:embed compatibility (embed cannot use .. paths)"
  - "Each language gets one parser in pool, reused across all Parse() calls"
  - "All tree-sitter CGo objects closed via defer — never runtime.SetFinalizer"
metrics:
  duration: ~8 minutes
  completed_date: 2026-06-22
  tasks_total: 14
  tasks_completed: 14
  files_created: 18
  test_count: 21
  test_pass_rate: 100%
---

# Phase 1 Plan 1: Core Parser Engine Summary

Tree-sitter integration for Go, Python, C, C++, and Rust. Source files are parsed and function/method/class/struct/interface signatures are extracted as typed `SymbolInfo` structs. Builds the parser pool, language registry, per-language tree-sitter query files (.scm), query runner, file scanner, and end-to-end integration tests.

## Tasks Completed

| # | Task | Type | Commit |
|---|------|------|--------|
| P1-T1 | Add tree-sitter Go deps + CGO_ENABLED=1 Makefile | auto | `e4f3cbf` |
| P1-T2 | Create internal/symbol/symbol.go — SymbolInfo, Kind, Param | auto | `23701a7` |
| P1-T3 | Create internal/parser/parser.go — parser pool with Close() | auto | `cfb2098` |
| P1-T4 | Create internal/parser/registry.go — language registry | auto | `15040f4` |
| P1-T5 | Create queries/go/docs.scm — Go tree-sitter query | auto | `11ae578` |
| P1-T6 | Create queries/python/docs.scm — Python query | auto | `11ae578` |
| P1-T7 | Create queries/c/docs.scm — C query | auto | `11ae578` |
| P1-T8 | Create queries/cpp/docs.scm — C++ query | auto | `11ae578` |
| P1-T9 | Create queries/rust/docs.scm — Rust query | auto | `11ae578` |
| P1-T10 | Create internal/parser/queryrunner.go — query→SymbolInfo | auto | `ea36bec` |
| P1-T11 | Create internal/scanner/scanner.go — file discovery | auto | `d001747` |
| P1-T12 | Create internal/scanner/scanner_test.go — scanner tests | auto | `7305583` |
| P1-T13 | Create test fixtures for all 5 languages | auto | `192cc1c` |
| P1-T14 | Create internal/parser/parser_test.go — integration tests | auto | `290cea0` |

## Verification Results

| Step | Result |
|------|--------|
| `go build ./...` | ✅ exit 0 |
| `go vet ./...` | ✅ exit 0 |
| `go test ./... -count=1` | ✅ all 21 tests pass |
| `go test ./... -count=1 -race` | ✅ all tests pass with race detector |
| `make build` | ✅ produces `deoxy` binary |
| `make test` | ✅ all targets pass with -race |
| Code audit for defer Close() | ✅ Every Parser, Query, QueryCursor, Tree has defer Close() |

## Language Support Verification

| Language | Functions | Methods | Structs | Classes | Interfaces/Traits |
|----------|-----------|---------|---------|---------|-------------------|
| Go | ✅ TestParseGoFunctions | ✅ TestParseGoMethods | ✅ TestParseGoStructs | N/A | ✅ TestParseGoInterfaces |
| Python | ✅ TestParsePythonFunctions | ✅ (self/cls detection) | N/A | ✅ TestParsePythonClasses | N/A |
| C | ✅ TestParseCFunctions | N/A | N/A | N/A | N/A |
| C++ | ✅ TestParseCppFunctionsAndClasses | ✅ (inside class) | ✅ (struct_specifier) | ✅ TestParseCppFunctionsAndClasses | N/A |
| Rust | ✅ TestParseRustFunctionsAndStructs | ✅ (impl_item) | ✅ TestParseRustFunctionsAndStructs | N/A | ✅ (trait_item) |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] go:embed does not support `..` paths**
- **Found during:** Task P1-T3/P1-T4
- **Issue:** `//go:embed ../../queries/go/docs.scm` fails because Go embed patterns cannot contain `..`
- **Fix:** Created `internal/parser/queries/` directory with copies of all .scm files. Registry embeds from `queries/go.scm` (relative path). Root `queries/*/docs.scm` files kept for reference.
- **Files modified:** `internal/parser/registry.go`, created `internal/parser/queries/*.scm`

**2. [Rule 1 - Bug] Incorrect tree-sitter query node names**
- **Found during:** Task P1-T14 test execution
- **Issue:** Go query used `type_specification` instead of `type_spec`; C query had wrong field order; Rust query missed `declaration_list` intermediate node in impl_item
- **Fix:** Updated all 3 query files based on AST dump analysis
- **Files modified:** `queries/go/docs.scm`, `queries/c/docs.scm`, `queries/rust/docs.scm`, `internal/parser/queries/*.scm`

**3. [Rule 1 - Bug] Method receiver not parsed for @method capture**
- **Found during:** Task P1-T14 `TestParseGoMethods` test
- **Issue:** Go methods use `@method` capture (not `@func`), but receiver parsing only happened in the `@func` path
- **Fix:** Added explicit `@receiver` capture check in the `@method` path of `extractSymbolInfo`
- **Files modified:** `internal/parser/queryrunner.go`

## Known Stubs

None — all SymbolInfo fields are populated from real tree-sitter captures.

## Threat Flags

None — threat model covers all created surface (no new network endpoints, no auth paths).

## Self-Check: PASSED

- All 18 created files verified on disk
- All 12 phase-1 commits exist in git log
- All 21 tests pass (14 in parser, 7 in scanner)
- All 5 languages parse test fixtures correctly
- `go build ./...` and `go vet ./...` pass with zero errors
- `make build` and `make test` succeed
