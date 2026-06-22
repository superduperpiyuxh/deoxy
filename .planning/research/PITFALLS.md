# Domain Pitfalls

**Domain:** Multi-language documentation comment generator
**Researched:** 2026-06-22

## Critical Pitfalls

Mistakes that cause rewrites or major issues.

### Pitfall 1: CGo Memory Leaks from tree-sitter
**What goes wrong:** tree-sitter allocates Parser, Tree, TreeCursor, Query, and QueryCursor objects on the C heap via CGo. If not explicitly closed, memory is reclaimed only when Go GC runs finalizers — but `runtime.SetFinalizer` has known bugs with CGo memory (golang-nuts/LIWj6Gl--es). In long-lived processes (VS Code sidecar, watch mode), this causes unbounded memory growth that is **invisible to Go pprof** because the C heap isn't tracked by the Go allocator.

**Why it happens:** The `go-tree-sitter` library uses `runtime.SetFinalizer` as a backstop, but finalizers may not run promptly or at all for C-allocated memory. The library's own README (v0.25.0) states: "Due to bugs with runtime.SetFinalizer and CGO, you must always call Close on an object that allocates memory from C."

**Consequences:** In long-running processes, RSS grows to hundreds of MB or GB. Process may be OOM-killed. Bug reports for the official bindings (issue #52) confirm 700MiB+ memory surges when processing large files.

**Prevention:**
- **Always** use `defer obj.Close()` for every Parser, Tree, Query, QueryCursor, TreeCursor — without exception
- Never rely on `runtime.SetFinalizer`
- In the VS Code sidecar, consider a process-per-request model (cheap process restart for memory hygiene)
- Test with `GODEBUG=cgocheck=2` and monitor RSS in CI

**Detection:** `pmap <pid> | grep total` shows RSS. If RSS grows while Go heap (`runtime.ReadMemStats().HeapInuse`) stays flat, it's CGo memory leak.

### Pitfall 2: tree-sitter Grammar Quirks Per Language
**What goes wrong:** Each tree-sitter grammar has unique node type names and AST structures. There is no universal "function name" capture — each language grammar uses different node types:
- Go: `function_declaration` → `name: (identifier)`
- Python: `function_definition` → `name: (identifier)`
- Rust: `function_item` → `name: (identifier)`
- JavaScript: `function_declaration` → `name: (identifier)`
- Java: `method_declaration` → `name: (identifier)`

**Why it happens:** Each language grammar is independently maintained. Node type naming is not standardized across grammars.

**Consequences:** Query files must be hand-crafted per language. A grammar update can rename node types (breaking change). For example, the Go grammar had a known bug (issue #128) with `map` types in function return values that required a grammar fix.

**Prevention:**
- Maintain separate `.scm` query files per language
- Pin grammar versions in go.mod
- Write a test suite that parses known-good code and verifies captures
- Subscribe to tree-sitter grammar repos for breaking change notifications

### Pitfall 3: Comment Insertion Position Errors
**What goes wrong:** Inserting a doc comment at the wrong position breaks the file. The comment must go immediately before the declaration (no blank lines between comment and declaration), after any package-level comments, and must not interfere with build tags (`//go:build`, `// +build`) or `go:generate` directives.

**Why it happens:** Tree-sitter provides byte offsets for nodes, but blank lines, shebangs (`#!/usr/bin/env python`), BOM, and preprocessor directives require careful handling.

**Consequences:** Generated files fail to compile. Worse: they compile but behave differently (e.g., build tags separated from file by a comment).

**Prevention:**
- When inserting, check for existing comments immediately above the target node
- Handle `//go:` directives (directive lines must not have blank lines inserted before them)
- For files with build tags, insert doc comments after the build tag block
- Validate output: compile or syntax-check after insertion
- Support `--dry-run` and `--diff` modes so users can review

## Moderate Pitfalls

### Pitfall 1: Go's Self-Documenting Convention is Minimalist
**What goes wrong:** Go doc comments use `// PackageName` on the first line, not `@param` tags. A user who expects `@param name type description` from doxdocgen will be surprised by Go's convention: `// FuncName does X. // // name is the user's name. // age is the user's age.`

**Prevention:** Generate GoDoc-style comments (plain prose, no tags) as the default for Go. Support a `--doxygen-style` flag for Go files if users want `@param` tags (some Go projects do use Doxygen-style comments).

### Pitfall 2: Python Docstrings Have Two Competing Conventions
**What goes wrong:** PEP 257 defines docstring syntax but not param tag format. Google style, NumPy style, Sphinx/reST style are all common. Users will want a specific convention.

**Prevention:** Make docstring style configurable per project:
```yaml
languages:
  python:
    docstring_style: "google"  # or "numpy", "sphinx", "pep257"
```

### Pitfall 3: VS Code Extension Binary Distribution
**What goes wrong:** The Go binary must be compiled for each platform (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, win32-x64). Forgetting one platform breaks the extension for those users. The official `golang/vscode-go` extension follows this pattern: binaries are stored in `extension/bin/{platform}/` within the VSIX.

**Prevention:**
- Use Go's `GOOS`/`GOARCH` cross-compilation in CI
- Build matrix: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Include all platform binaries in the VSIX
- Refer to Tauri sidecar naming convention for reference (platform triple suffix)

### Pitfall 4: JSON-RPC Framing in the Sidecar
**What goes wrong:** JSON-RPC over stdio needs message framing. The simplest approach (newline-delimited JSON) breaks if JSON contains newlines (e.g., in source code strings).

**Prevention:** Use either:
- `Content-Length: N\r\n\r\n{json}` framing (LSP standard) — robust for arbitrary JSON
- Newline-delimited JSON with escaped newlines in strings — simpler but less robust

Recommend: Use the LSP-style Content-Length header framing from the start. It's the same framing used by `gopls`, `typescript-language-server`, and `agent-lsp`.

## Minor Pitfalls

### Pitfall 1: Binary Size
**What goes wrong:** A statically-linked Go binary with tree-sitter + grammars can be 10-20MB. This matters for VS Code extension download size.

**Prevention:** Use UPX compression for binaries in the VSIX. Tree-sitter's C library is the main size contributor; grammars are linked separately so only ship the grammars needed.

### Pitfall 2: Unicode/Rune Handling
**What goes wrong:** tree-sitter byte offsets don't map 1:1 to Go `string` indices for multi-byte characters (Chinese, emoji in comments).

**Prevention:** Always use `[]byte` and byte offsets for tree-sitter operations. Convert to string only for template rendering.

### Pitfall 3: Editor Line Ending Mismatch
**What goes wrong:** Files with CRLF line endings (`\r\n`) have different byte offsets than LF files. Inserting text with only `\n` breaks CRLF files.

**Prevention:** Detect line endings from the source file. Preserve original line endings in the output.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Phase 1: tree-sitter parsing | CGo memory leak (Pitfall 1) | `defer Close()` everywhere; test with long-running parser reuse |
| Phase 1: Language grammars | Grammar quirks per language (Pitfall 2) | Build `.scm` query files and test suite per language |
| Phase 2: Go doc generation | Go's minimalist doc style (Moderate 1) | Default to GoDoc prose; offer Doxygen-style as option |
| Phase 2: Python doc generation | Multiple style conventions (Moderate 2) | Configurable `docstring_style` setting |
| Phase 3: Source insertion | Comment position errors (Pitfall 3) | Handle build tags, directives; validate output |
| Phase 5: VS Code extension | Binary distribution per platform (Moderate 3) | CI build matrix; test on all platforms |
| Phase 5: VS Code extension | JSON-RPC framing (Moderate 4) | Use Content-Length header framing from day one |

## Sources

- `github.com/tree-sitter/go-tree-sitter` README — CGo Close() warning (HIGH confidence)
- `github.com/tree-sitter/go-tree-sitter/issues/52` — memory leak report (HIGH confidence)
- `golang-nuts/LIWj6Gl--es` — runtime.SetFinalizer + CGo bugs (HIGH confidence)
- `github.com/smacker/go-tree-sitter/issues/181` — memory behavior with finalizers (HIGH confidence)
- `github.com/tree-sitter/tree-sitter-go/issues/128` — grammar bug with map types (HIGH confidence)
- `github.com/parsiya/blog` — tree-sitter query deep dive, Go return type queries (MEDIUM confidence)
- `golang/vscode-go` extension — sidecar distribution pattern (HIGH confidence)
- Doxygen manual — comment format variations (HIGH confidence)
- PEP 257 — Python docstring conventions (HIGH confidence)
