# Architecture Patterns

**Domain:** Multi-language documentation comment generator
**Researched:** 2026-06-22

## Recommended Architecture

```
┌─────────────────────────────────────────────────────┐
│                    CLI (cobra)                       │
│  deoxy generate ./... --lang go,py,rs               │
│  deoxy init                                         │
│  deoxy watch ./...                                  │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│               Core Engine (Go library)               │
│                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │ Scanner  │  │  Parser  │  │  Query/Runner     │   │
│  │ (glob    │─→│ (tree-   │─→│  (captures →      │   │
│  │  files)  │  │  sitter) │  │   SymbolInfo)     │   │
│  └──────────┘  └──────────┘  └────────┬─────────┘   │
│                                        │             │
│  ┌─────────────────────────────────────▼──────────┐  │
│  │              Template Engine                     │  │
│  │  LanguageConfig → CommentTemplate → filled text │  │
│  └─────────────────────────────────────┬──────────┘  │
│                                        │             │
│  ┌─────────────────────────────────────▼──────────┐  │
│  │              Source Writer                       │  │
│  │  Insert/update/replace comments in source files  │  │
│  └──────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│              VS Code Extension (TypeScript)           │
│                                                       │
│  ┌──────────────┐       ┌─────────────────────────┐  │
│  │ Extension    │──────→│ Go sidecar (deoxy serve) │  │
│  │ activates →  │       │ JSON-RPC over stdio      │  │
│  │ spawn binary │←──────│ response + streaming     │  │
│  └──────────────┘       └─────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| **Scanner** | Discover source files by glob pattern, filter by language extension | Filesystem |
| **Parser** | Wrap tree-sitter parser lifecycle (create, set language, parse, close) | tree-sitter C library via CGo |
| **QueryRunner** | Load `.scm` query files, execute against parsed tree, return typed `SymbolInfo` structs | Parser, query files |
| **TemplateEngine** | Given `SymbolInfo` + language config, render doc comment text | Config |
| **SourceWriter** | Read source file, insert/modify comment at correct position, write back | Filesystem (in-place edit) |
| **Config** | Load `.deoxy.yaml`, provide per-language settings | Filesystem |
| **VS Code Extension** | Spawn Go binary, manage lifecycle, send/receive JSON-RPC | Go sidecar |
| **Go sidecar** | Accept JSON-RPC document, return generated comments, handle keepalive | VS Code Extension |

### Data Flow (CLI mode)

```
1. deoxy generate ./src
2. Scanner: walk ./src, collect *.go, *.py, *.js, *.ts files
3. Parser: for each file, create tree-sitter parser for detected language
4. QueryRunner: run language-specific .scm queries, produce []SymbolInfo
5. SymbolInfo contains: Name, Kind (function/method/struct), Params, Returns, TypeParams, Doc
6. TemplateEngine: for each SymbolInfo, render doc comment per language format
7. SourceWriter: for each file, insert/replace comment blocks, write result
8. stdout: summary (X files processed, Y comments generated)
```

### Data Flow (VS Code mode)

```
1. User opens file in editor, triggers "Generate Doc" command
2. Extension sends JSON-RPC request: { method: "generate", params: { filePath, language, source } }
3. Go sidecar parses source with tree-sitter, generates comments
4. Response: { jsonrpc: "2.0", result: { edits: [{ offset, length, text }] } }
5. Extension applies edits to the document via VS Code API
```

## Patterns to Follow

### Pattern 1: Language Registry
**What:** A central registry mapping file extensions → language configs → tree-sitter grammars → query files.
**When:** At application startup.
**Example:**
```go
type LanguageConfig struct {
    Name             string
    FileExtensions   []string
    TreeSitterGrammar unsafe.Pointer // ts.Language()
    QueryFile        string          // embedded .scm content
    CommentStyle     CommentStyle    // prefix, suffix, open, close
    TagTemplates     map[string]string // @param → "/** @param {name} {type} - {desc} */"
    DefaultOrder     []string        // ["brief", "tparam", "param", "return"]
}

var registry = map[string]*LanguageConfig{
    ".go": {
        Name:             "go",
        FileExtensions:   []string{".go"},
        TreeSitterGrammar: tree_sitter_go.GetLanguage(),
        CommentStyle:     CommentStyle{Prefix: "// "},
        DefaultOrder:     []string{"brief", "param", "return"},
    },
    ".py": {
        Name:             "python",
        FileExtensions:   []string{".py"},
        TreeSitterGrammar: tree_sitter_python.GetLanguage(),
        CommentStyle:     CommentStyle{Open: `"""`, Close: `"""`},
        DefaultOrder:     []string{"summary", "args", "returns", "raises"},
    },
}
```

### Pattern 2: Query Files as Embedded Assets
**What:** Store `.scm` query files in `queries/` directory, embed them at build time with `//go:embed`.
**When:** Each language has its own `.scm` file for function/method/struct detection.
**Example:**
```go
//go:embed queries/go/docs.scm
var goQuery string

// queries/go/docs.scm
// (function_declaration
//   name: (identifier) @function.name
//   parameters: (parameter_list
//     (parameter_declaration
//       name: (identifier) @function.param.name
//       type: (_) @function.param.type))
//   result: (_) @function.return.type)
```

### Pattern 3: JSON-RPC over stdio for VS Code Sidecar
**What:** Newline-delimited JSON-RPC 2.0 on stdin/stdout. Stderr reserved for logging.
**When:** VS Code extension spawns Go binary with `child_process.spawn()`.
**Example (Go sidecar):**
```go
// stdin: {"jsonrpc":"2.0","id":1,"method":"generate","params":{"source":"...","language":"go"}}
// stdout: {"jsonrpc":"2.0","id":1,"result":{"edits":[...]}}

// Stderr is used for logging (visible in VS Code Developer Tools console):
// [2026-06-22 10:00:00] INFO: parsed 150 functions in main.go
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Per-request Parser Instantiation
**What:** Creating a new tree-sitter `Parser` for every file parsed.
**Why bad:** tree-sitter parser initialization allocates C memory. Doing it per-file in a batch of 1000 files causes GC pressure and CGo overhead.
**Instead:** Reuse parsers per language. Create one parser per language and reuse across files:
```go
// Good
parserPool := map[string]*ts.Parser{}
for _, lang := range languages {
    p := ts.NewParser()
    p.SetLanguage(ts.NewLanguage(lang.Grammar()))
    parserPool[lang.Name] = p
    defer p.Close()
}
```

### Anti-Pattern 2: Modifying Source via String Operations
**What:** Using string concatenation to insert comments into source code.
**Why bad:** Breaks on files with non-UTF8 content, BOM, mixed line endings. Byte offsets from tree-sitter are byte-oriented, not rune-oriented.
**Instead:** Work with `[]byte`, track byte offsets, use `bytes.Buffer` for assembly. Verify with `gofmt`/`rustfmt` that output is valid.

### Anti-Pattern 3: Blocking the VS Code UI Thread
**What:** Running tree-sitter parsing synchronously in the extension host process.
**Why bad:** Large files will freeze the editor. tree-sitter uses CGo which blocks the calling goroutine's OS thread.
**Instead:** Always run the Go sidecar as a separate process. The extension sends requests and receives async responses. If needed, the sidecar can use goroutines internally.

## Scalability Considerations

| Concern | At 100 files | At 10K files | At 100K files |
|---------|--------------|--------------|---------------|
| Parser memory | ~5MB (per-language parser) | ~5MB parsers + peak tree memory | Same; files processed sequentially |
| Batch processing | Single goroutine is fine | Worker pool with `errgroup` | Worker pool + file-level parallelism |
| tree-sitter parse time | ~0.1ms per small file | ~1-2 seconds total | ~10-20 seconds; use incremental parsing? |
| VS Code extension | Instant | Sub-second per file | May need progress reporting |

**Key insight:** tree-sitter is extremely fast (0.1-1ms per file even for moderate-sized files). The bottleneck will be file I/O and writer operations, not parsing. No special scalability architecture needed for MVP.

## Sources

- doxdocgen architecture (TypeScript extension only) — reference for feature behavior (HIGH confidence)
- `golang/vscode-go` sidecar (vscgo) — process management pattern (HIGH confidence)
- `github.com/blackwell-systems/agent-lsp` — JSON-RPC over stdio Go sidecar architecture (HIGH confidence)
- `github.com/gsd-build/gsd-2` — VS Code extension spawning Go binary with JSON-RPC (HIGH confidence)
- Tree-sitter API docs — parser/query/cursor lifecycle (HIGH confidence)
- `github.com/strings77wzq/claude-code-Go` — MCP stdio transport pattern (HIGH confidence)
