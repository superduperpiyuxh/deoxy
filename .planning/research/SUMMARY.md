# Research Summary: deoxy

**Domain:** Multi-language documentation comment generator (CLI + VS Code extension)
**Researched:** 2026-06-22
**Overall confidence:** HIGH

## Executive Summary

deoxy is a multi-language documentation generator inspired by the VS Code extension `doxdocgen` (which auto-generates Doxygen comments for C/C++/C#), but reimagined as a cross-language tool written in Go. The core idea: parse source code with tree-sitter to extract function/method/type signatures, then inject properly-formatted doc comments (GoDoc, JSDoc, Javadoc, Rustdoc, Python docstrings, Doxygen) back into the source.

The ecosystem research reveals three critical findings:

**1. tree-sitter in Go is production-ready but has a migration in progress.** The community-maintained `github.com/smacker/go-tree-sitter` (560 stars, last updated Aug 2024) is being superseded by the official `github.com/tree-sitter/go-tree-sitter` (v0.25.0, Feb 2025, maintained by the tree-sitter org). The official bindings have a modular architecture — each language grammar is a separate module you `go get` individually (e.g., `github.com/tree-sitter/tree-sitter-go`). **Recommendation: use the official bindings from day one.** The API is nearly identical; the migration is straightforward.

**2. No existing tool does what deoxy proposes.** The Go ecosystem has `godoc` / `go/doc` / `go/doc/comment` for extracting and rendering Go documentation, and various AI-powered doc generators (autodocs, gocomments, codedocgen), but none are a deterministic, tree-sitter-based, multi-language doc comment generator that inserts comments into source files. This is a genuine gap.

**3. The VS Code + Go sidecar pattern is well-established.** The official `golang/vscode-go` extension ships the `vscgo` Go binary as a sidecar (distributed with the VSIX, installed via `go install` into the extension directory). Communication uses JSON-RPC 2.0 over stdio (newline-delimited JSON on stdin/stdout). The `agent-lsp` project provides a clean reference architecture.

## Key Findings

**Stack:** Go 1.23+ using `github.com/tree-sitter/go-tree-sitter` (official bindings v0.25+) with per-language grammar modules. Cobra/Viper for CLI, TOML/YAML for per-language config.

**Architecture:** Layered: (1) tree-sitter node extraction layer, (2) language-specific query files (`.scm`), (3) comment template engine, (4) output formatter with source injection, (5) VS Code sidecar using JSON-RPC over stdio.

**Critical pitfall:** CGo memory management — tree-sitter allocates on the C heap. The `runtime.SetFinalizer` approach is unreliable for long-lived processes. **Always call `Close()` explicitly** on Parser, Tree, TreeCursor, Query, and QueryCursor via `defer`. Memory leaks in CGo are invisible to Go's pprof.

## Implications for Roadmap

Based on research, suggested phase structure:

1. **Phase 1: Core tree-sitter integration** — Parse Go, Python, JS/TS, Rust, Java files and extract function signatures, parameter names/types, return types. Build the query files (`.scm`). Validate on real codebases.
   - Addresses: FEATURES.md table stakes (function detection)
   - Avoids: PITFALLS.md CGo memory leak by establishing `Close()` discipline early

2. **Phase 2: Comment generation engine** — Template system for each language's doc format. `@param`, `@return`, `@brief` templates, configurable order.
   - Addresses: FEATURES.md core feature (auto-generate doc comments)
   - Avoids: PITFALLS.md template flexibility issues

3. **Phase 3: Source injection** — Insert generated comments into source files at correct positions. Handle existing comments (skip/update/replace modes).
   - Addresses: FEATURES.md differentiators (in-place insertion)
   - Avoids: PITFALLS.md comment insertion position errors

4. **Phase 4: CLI + config system** — `deoxy init`, `deoxy generate`, per-language config files (`.deoxy.yaml`/`.deoxy.toml`).
   - Addresses: FEATURES.md CLI UX
   - Avoids: PITFALLS.md configuration complexity

5. **Phase 5: VS Code extension** — TypeScript extension that spawns Go binary as sidecar, JSON-RPC 2.0 over stdio.
   - Addresses: FEATURES.md VS Code integration
   - Avoids: PITFALLS.md process lifecycle management

6. **Phase 6: Advanced features** — Smart text generation, git integration, batch mode, watch mode.
   - Addresses: FEATURES.md differentiators

**Research flags for phases:**
- Phase 1: May need deeper tree-sitter query debugging per language — some grammars have quirks (e.g., Go's `map` types in return values)
- Phase 2: Standard patterns, unlikely to need research
- Phase 3: Comment detection (find existing comments) is straightforward; inserting at correct byte offsets needs care
- Phase 4: Standard patterns
- Phase 5: VS Code extension packaging (platform-specific Go binaries per target triple) needs build infrastructure
- Phase 6: Smart text generation may warrant LLM integration research; defer to later

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | tree-sitter/go-tree-sitter v0.25+ is confirmed production-ready. CGo gotchas are documented. |
| Features | HIGH | doxdocgen provides clear reference. Doc formats per language are well-documented standards. |
| Architecture | HIGH | Layered architecture is well-patterned. VS Code + Go sidecar has multiple production references. |
| Pitfalls | HIGH | CGo memory management is the #1 risk. All other issues are well-documented in existing projects. |

## Gaps to Address

- **tree-sitter query coverage for all target languages needs manual validation** — Go grammar is well-understood; Python, Rust, Java, JS/TS need `.scm` file authoring and testing
- **VS Code extension packaging** — Cross-compiling Go binaries for linux/darwin/windows × amd64/arm64 is standard but requires CI matrix setup
- **Existing comment detection** — Need to decide on strategy for comments that already exist (replace? append? skip?)
