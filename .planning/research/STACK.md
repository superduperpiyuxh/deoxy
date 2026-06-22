# Technology Stack

**Project:** deoxy
**Researched:** 2026-06-22

## Recommended Stack

### Core Framework
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.23+ | Core runtime | Single-binary distribution, cross-compilation, excellent CGo support for tree-sitter, stdlib JSON-RPC. 1.23 required by `tree-sitter/go-tree-sitter` v0.25+. |
| `github.com/tree-sitter/go-tree-sitter` | v0.25.0+ | Tree-sitter Go bindings (official) | **Official** bindings from tree-sitter org. Modular grammars. Active maintenance (last push Nov 2025, Apr 2026). Replaces the community `smacker/go-tree-sitter`. |
| `github.com/spf13/cobra` | latest | CLI framework | De facto standard for Go CLIs. Subcommands, help, completion. |
| `github.com/spf13/viper` | latest | Configuration | Supports YAML, TOML, JSON config files. Environment variable binding. |

### Language Grammar Packages (tree-sitter)
| Module | Purpose |
|--------|---------|
| `github.com/tree-sitter/tree-sitter-go` | Go parser |
| `github.com/tree-sitter/tree-sitter-python` | Python parser |
| `github.com/tree-sitter/tree-sitter-javascript` | JavaScript parser |
| `github.com/tree-sitter/tree-sitter-typescript` | TypeScript parser |
| `github.com/tree-sitter/tree-sitter-rust` | Rust parser |
| `github.com/tree-sitter/tree-sitter-java` | Java parser |
| `github.com/tree-sitter/tree-sitter-c` | C parser |
| `github.com/tree-sitter/tree-sitter-cpp` | C++ parser |

**Note:** Each grammar is imported and registered separately. This is intentional — users only pay for what they use. The `Language()` function from each grammar is passed to `tree_sitter.NewLanguage()`.

### Supporting Libraries
| Library | Purpose | When to Use |
|---------|---------|-------------|
| `go.uber.org/zap` | Structured logging | CLI debug/stderr output, VS Code sidecar stderr logging |
| `github.com/charmbracelet/glamour` | Markdown rendering for terminal | If deoxy supports `--format markdown` output |
| `gopkg.in/yaml.v3` | YAML config parsing | If viper is too heavy for embedded config; otherwise viper handles this |
| Standard `encoding/json` | JSON-RPC (VS Code sidecar) | No external dependency needed for JSON-RPC over stdio |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| tree-sitter bindings | `tree-sitter/go-tree-sitter` (official) | `smacker/go-tree-sitter` | smacker is frozen (last push Aug 2024). Official bindings have modular grammars, better maintenance, and are the project's future. Migration paths exist (see uzomuzo-oss#272). |
| tree-sitter bindings (cont.) | official | Writing custom CGo per language | Infinitely more work. tree-sitter handles incremental parsing, error recovery, and has 40+ grammars. |
| CLI framework | cobra | `alecthomas/kong`, `urfave/cli` | Cobra is the most widely adopted, has the best help generation, and integrates with viper for config. |
| Config format | YAML/TOML | JSON, HCL | YAML is familiar to developers. TOML is a good alternative. JSON is poor for human-edited config. |
| VS Code sidecar protocol | JSON-RPC over stdio | LSP, HTTP, WebSocket | JSON-RPC over stdio is the simplest, most battle-tested pattern. No port allocation needed. No network security concerns. LSP is overkill (deoxy isn't a language server). |
| Doc generation approach | tree-sitter AST | Regex, language server protocol | Regex fails on edge cases (nested generics, multiline signatures). LSP is heavy (requires running a language server per file). tree-sitter gives exactly the right level of structural analysis. |

## Installation

```bash
# Core runtime
go get github.com/tree-sitter/go-tree-sitter@v0.25.0

# Language grammars (as needed)
go get github.com/tree-sitter/tree-sitter-go@latest
go get github.com/tree-sitter/tree-sitter-python@latest
go get github.com/tree-sitter/tree-sitter-javascript@latest
go get github.com/tree-sitter/tree-sitter-typescript@latest
go get github.com/tree-sitter/tree-sitter-rust@latest
go get github.com/tree-sitter/tree-sitter-java@latest
go get github.com/tree-sitter/tree-sitter-c@latest
go get github.com/tree-sitter/tree-sitter-cpp@latest

# CLI framework
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
```

## Sources

- `github.com/tree-sitter/go-tree-sitter` — official Go bindings, v0.25.0, Feb 2025 (HIGH confidence)
- `github.com/smacker/go-tree-sitter` — community bindings, frozen Aug 2024 (HIGH confidence)
- `github.com/future-architect/uzomuzo-oss/issues/272` — migration from smacker to official (HIGH confidence)
- Tree-sitter Query API docs (HIGH confidence)
- `pkg.go.dev/github.com/tree-sitter/go-tree-sitter` — API reference (HIGH confidence)
