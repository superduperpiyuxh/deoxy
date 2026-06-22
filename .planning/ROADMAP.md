# Roadmap: deoxy

**Module:** `github.com/superduperpiyuxh/deoxy`
**Go version:** 1.26.3
**Description:** Multi-language documentation comment generator (CLI + VS Code extension) — parses source code with tree-sitter, generates properly-formatted doc comments (GoDoc, JSDoc, Javadoc, Python docstrings, Rustdoc, Doxygen), and injects them into source files.

**Milestone:** v1.0 MVP — CLI that generates doc comments for Go, Python, C, C++, and Rust in batch mode.

## Phases

- [x] **Phase 0: Foundation** — Project scaffolding, module init, CI, README, directory structure
- [x] **Phase 1: Core Parser Engine** — Tree-sitter integration, language registry, query files, AST extraction
- [x] **Phase 2: Template Engine** — Go text/template-based doc generation, per-language comment templates, config system
- [ ] **Phase 3: CLI** — Cobra CLI with generate/init/watch commands, source injection, stdin/stdout mode
- [ ] **Phase 4: VS Code Extension** — Thin TypeScript extension + Go sidecar via JSON-RPC over stdio
- [ ] **Phase 5: Advanced Features** — Smart text (getter/setter/ctor detection), alignment, custom tags, git-aware mode

## Phase Details

### Phase 0: Foundation
**Goal**: Project is scaffolded, builds, and has CI verifying it compiles
**Depends on**: Nothing
**Key files**: `main.go`, `go.mod`, `Makefile`, `.goreleaser.yaml`, `.github/workflows/ci.yml`, `README.md`
**Estimated complexity**: S
**Success Criteria** (what must be TRUE):
  1. `go build ./...` succeeds with zero errors
  2. `go test ./...` passes (trivial placeholder test)
  3. CI pipeline (GitHub Actions) runs on push/PR and reports green
  4. Project directory layout is established (`cmd/`, `internal/`, `queries/`)
**Anti-goals**:
  - Do NOT add any tree-sitter dependencies yet
  - Do NOT design the CLI API surface (that's Phase 3)
  - Do NOT implement any application logic
  - Do NOT set up VS Code extension infrastructure
**Plans**: 1 (phase-0) ✅ Completed

---

### Phase 1: Core Parser Engine
**Goal**: deoxy can parse Go, Python, C, C++, and Rust source files and extract function/method/class/struct signatures as typed `SymbolInfo` structs
**Depends on**: Phase 0
**Key files**:
  - `internal/symbol/symbol.go` — `SymbolInfo` type (`Name`, `Kind`, `Params`, `Returns`, `TypeParams`, `Doc`)
  - `internal/parser/parser.go` — Tree-sitter parser wrapper with `defer Close()` discipline
  - `internal/parser/registry.go` — Language registry mapping extensions → grammars → query files
  - `internal/scanner/scanner.go` — File discovery by glob/language extension
  - `queries/go/docs.scm` — Tree-sitter query for Go functions, methods, structs, interfaces
  - `queries/python/docs.scm` — Python function/class detection
  - `queries/c/docs.scm` — C function detection
  - `queries/cpp/docs.scm` — C++ function/method/class detection
  - `queries/rust/docs.scm` — Rust function/method/struct/trait detection
  - `internal/parser/parser_test.go` — Test suite per language with real code samples
**Estimated complexity**: L
**Success Criteria** (what must be TRUE):
  1. `deoxy` can parse a Go source file and extract all exported function signatures (name, parameters with types, return types)
  2. `deoxy` can parse Go struct definitions and extract field names and types
  3. `deoxy` can parse Python function definitions with type annotations
  4. `deoxy` can parse C/C++ function declarations (including pointers, arrays in params)
  5. `deoxy` can parse Rust function items and struct definitions
  6. All tree-sitter Parser/Query/QueryCursor objects are explicitly `Close()`d (verified by test with leak detection)
  7. Parsers are reused per language (not recreated per file) via a parser pool
**Anti-goals**:
  - Do NOT generate any doc comment text (that's Phase 2)
  - Do NOT modify source files (that's Phase 3)
  - Do NOT implement CLI commands beyond a debug dump flag
  - Do NOT attempt to extract doc comments from function bodies (e.g., `throw`/`raise` detection)
  - Do NOT handle all possible AST node types — start with functions, methods, structs, classes
  - Do NOT use `runtime.SetFinalizer` for CGo cleanup — always use explicit `Close()`
**Plans**: 1 (phase-1) ✅ Completed

---

### Phase 2: Template Engine
**Goal**: ✅ Complete — Given `SymbolInfo` from Phase 1, deoxy renders proper doc comments in the target language's format (GoDoc prose, Python docstrings, Doxygen/JSDoc-style for C/C++, Rustdoc)
**Depends on**: Phase 1
**Key files**:
  - `internal/template/engine.go` — Template engine core, applies templates to SymbolInfo
  - `internal/template/godoc.go` — GoDoc comment templates (prose-style `// FuncName ...`)
  - `internal/template/pydoc.go` — Python docstring templates (Google style by default, configurable)
  - `internal/template/doxygen.go` — Doxygen/JSDoc-style comment templates for C/C++
  - `internal/template/rustdoc.go` — Rust `///` doc comment templates
  - `internal/config/config.go` — Per-language config loading (`deoxy.yaml`), docstring style selection
  - `internal/config/config_test.go`
  - `internal/template/engine_test.go` — Golden file tests for each language's output
**Estimated complexity**: M
**Success Criteria** (what must be TRUE):
  1. For a Go function `func Add(a int, b int) int`, deoxy generates `// Add adds a and b and returns the sum.\n//\n// a is the first operand.\n// b is the second operand.`
  2. For a Python function `def greet(name: str, age: int) -> str`, deoxy generates Google-style docstring with `Args:` and `Returns:` sections
  3. For a C function `int divide(int dividend, int divisor)`, deoxy generates `/**\n * @brief ...\n * @param dividend ...\n * @param divisor ...\n * @return ...\n */`
  4. Template tag order is configurable (brief first vs. param first)
  5. Go output uses GoDoc convention (prose, no `@param` tags) by default, with configurable `--doxygen-style` flag
  6. Config file (`.deoxy.yaml`) can set per-language docstyle preferences
**Anti-goals**:
  - Do NOT insert comments into source files (that's Phase 3)
  - Do NOT implement CLI commands yet (Phase 3)
  - Do NOT build the VS Code extension yet (Phase 4)
  - Do NOT implement smart text / name inference (Phase 5)
  - Do NOT handle AI-generated descriptions
  - Do NOT build watch mode
**Plans**: 8 tasks across 4 waves
**Plan file**: `.planning/plans/phase-2.json`

Plans:
- [x] P2-T1 — `internal/template/engine.go` — Core template engine with Render()
- [x] P2-T2 — `internal/template/engine_test.go` — Engine core tests
- [x] P2-T3 — `internal/template/templates.go` — Per-language template definitions (GoDoc, Doxygen, PyDoc, Rustdoc)
- [x] P2-T4 — `internal/template/templates_test.go` — Golden file/snapshot tests per language
- [x] P2-T5 — `internal/config/config.go` — Config file loader for .deoxy.yaml
- [x] P2-T6 — `internal/config/config_test.go` — Config loading and validation tests
- [x] P2-T7 — `internal/generator/generator.go` — Pipeline orchestrator (config → scan → parse → render)
- [x] P2-T8 — `internal/generator/generator_test.go` — End-to-end pipeline tests

---

### Phase 3: CLI
**Goal**: `deoxy generate`, `deoxy init`, `deoxy watch`, and `deoxy serve` work as a polished CLI tool with source file injection
**Depends on**: Phase 2
**Key files**:
  - `cmd/root.go` — Root cobra command, global flags
  - `cmd/generate.go` — `deoxy generate [path...]` command
  - `cmd/init.go` — `deoxy init` (scaffold `.deoxy.yaml` in project)
  - `cmd/watch.go` — `deoxy watch` (fsnotify-based file watching)
  - `internal/writer/writer.go` — SourceWriter: read → insert/replace → write
  - `internal/writer/injector.go` — Comment injection logic (byte-offset correct, handles build tags, BOM, line endings)
  - `internal/writer/injector_test.go` — Roundtrip tests (generate → format check compiled)
  - `main.go` — Entry point calling `cmd.Execute()`
  - `.deoxy.yaml` — Default config generated by `deoxy init`
**Estimated complexity**: M
**Success Criteria** (what must be TRUE):
  1. `deoxy generate ./src` finds all supported source files, generates doc comments, and writes them back in-place
  2. `deoxy generate --diff ./src` shows proposed changes without modifying files
  3. `deoxy generate --dry-run ./src` processes files but does not write any output
  4. `deoxy init` creates a `.deoxy.yaml` config file with sensible defaults for the project
  5. `deoxy watch ./src` re-generates comments when source files change (fsnotify)
  6. Generated Go files pass `gofmt` validation after comment insertion
  7. Existing comments are detected and respected (skip mode by default; `--force` overwrites)
  8. Build directives (`//go:build`, `//go:generate`) are not broken by comment insertion
  9. Original line endings (LF/CRLF) are preserved in output
**Anti-goals**:
  - Do NOT build the VS Code extension (Phase 4)
  - Do NOT add AI integration
  - Do NOT implement smart text inference (Phase 5)
  - Do NOT implement git-aware mode (Phase 5)
  - Do NOT build a documentation site generator (out of scope permanently)
**Plans**: TBD
**UI hint**: yes

---

### Phase 4: VS Code Extension
**Goal**: Users can generate doc comments from VS Code with a single command/keystroke, using the deoxy binary as a sidecar
**Depends on**: Phase 3
**Key files**:
  - `extension/package.json` — Extension manifest, commands, activation events
  - `extension/src/extension.ts` — Main extension entry point, register commands
  - `extension/src/sidecar.ts` — Go binary lifecycle management (spawn, health check, restart)
  - `extension/src/protocol.ts` — JSON-RPC 2.0 types and Content-Length framing helpers
  - `extension/tsconfig.json` — TypeScript config
  - `extension/esbuild.mjs` — Build bundler
  - `cmd/serve.go` — `deoxy serve` JSON-RPC sidecar subcommand
  - `internal/rpc/handler.go` — JSON-RPC method dispatch (`generate`, `ping`, `shutdown`)
  - `internal/rpc/types.go` — JSON-RPC 2.0 request/response structs
  - `internal/rpc/transport.go` — Content-Length header framing I/O
  - `Makefile` — Cross-compilation targets for `linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, `windows/amd64`
  - `.github/workflows/release.yml` — Build matrix + VSIX packaging
**Estimated complexity**: L
**Success Criteria** (what must be TRUE):
  1. VS Code extension activates when a supported language file is opened
  2. "Generate Doc Comment" command inserts a doc comment above the current function/method
  3. Go sidecar binary is distributed with the VSIX and spawned automatically on activation
  4. Sidecar process lifecycle is managed (start on activation, graceful shutdown on deactivation, restart on crash)
  5. JSON-RPC communication uses Content-Length header framing (LSP standard)
  6. Extension works on all three platforms (Linux, macOS, Windows) on both amd64 and arm64
  7. Error handling: if Go binary is missing or crashes, extension shows an actionable error message
**Anti-goals**:
  - Do NOT implement LSP (too heavy for doc generation)
  - Do NOT build a custom UI beyond VS Code native commands
  - Do NOT add AI-powered completions
  - Do NOT implement streaming response for large files (single-shot requests are fine)
  - Do NOT rely on `runtime.SetFinalizer` in the sidecar — process-per-request model or explicit `Close()` only
**Plans**: TBD
**UI hint**: yes

---

### Phase 5: Advanced Features
**Goal**: deoxy produces smarter doc comments with context-aware naming, aligned parameter docs, custom tags, and git-aware operation
**Depends on**: Phase 3 (can be built independently of Phase 4)
**Key files**:
  - `internal/smarttext/namer.go` — Name inference engine (camelCase/snake_case splitting, getter/setter/ctor detection)
  - `internal/smarttext/namer_test.go` — Edge case tests (single-letter names, abbreviations, acronyms)
  - `internal/template/alignment.go` — Column-aligned `@param`, `@return` tag formatting
  - `internal/template/customtags.go` — User-defined tag injection (`@note`, `@example`)
  - `internal/git/diff.go` — Git-aware file filtering (`git diff --name-only`)
  - `internal/smarttext/registry.go` — Common name mappings (`count → "Number of items"`, `name → "The name"`)
**Estimated complexity**: M
**Success Criteria** (what must be TRUE):
  1. `isGetter("GetName")` and `isSetter("SetName")` return true; generated comments say "Gets the name" / "Sets the name"
  2. `isConstructor("NewConfig")` returns true; generated comment says "NewConfig creates a new Config"
  3. Parameter name `count` infers description "Number of items" (from registry)
  4. CamelCase names like `maxRetryCount` split into readable words for description inference
  5. `@param` tags can be aligned with padding for clean column formatting
  6. Users can define custom tags in `.deoxy.yaml` (e.g., `custom_tags: ["note", "example"]`)
  7. `deoxy generate --git-aware` only processes files modified since the last commit
  8. All smart text features are opt-in (disabled by default; enabled via config flag)
**Anti-goals**:
  - Do NOT add LLM/AI-powered description generation (optional future plugin at most)
  - Do NOT add `@throws`/`@exception` detection from function bodies (deferred to v2.0)
  - Do NOT add deprecation detection (low value, easy to add later)
  - Do NOT add full documentation site rendering (out of scope permanently)
  - Do NOT support all 40+ tree-sitter grammars (stick to 8 major languages)
**Plans**: TBD

---

## Dependency Map

```
Phase 0 (Foundation)
   └──→ Phase 1 (Parser Engine)
            └──→ Phase 2 (Template Engine)
                     └──→ Phase 3 (CLI)
                              ├──→ Phase 4 (VS Code Extension)
                              └──→ Phase 5 (Advanced Features)
```

Phase 4 depends on Phase 3 (needs the sidecar `serve` command).
Phase 5 depends on Phase 3 (needs CLI config) but not Phase 4 (works without VS Code).

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 0. Foundation | 8/8 | ✅ Completed | 2026-06-22 |
| 1. Core Parser Engine | 14/14 | ✅ Completed | 2026-06-22 |
| 2. Template Engine | 8/8 | ✅ Completed | 2026-06-22 |
| 3. CLI | 0/0 | Not started | - |
| 4. VS Code Extension | 0/0 | Not started | - |
| 5. Advanced Features | 0/0 | Not started | - |
