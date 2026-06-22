# Feature Landscape

**Domain:** Multi-language documentation comment generator
**Researched:** 2026-06-22

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Function/method signature detection | Must know what to document | Low | tree-sitter queries capture `function_declaration`, `method_declaration`, etc. Built-in from Phase 1. |
| Parameter extraction | Must list `@param` tags | Low | Named parameters with types from tree-sitter captures |
| Return type detection | Must populate `@return` tags | Low | `result` field in function AST nodes |
| Comment insertion at correct position | Must not break code | Med | Insert before function declaration, respecting blank lines and existing comments |
| Respect existing comments | Must not overwrite user's comments | Low-Med | Check for preceding comment blocks; optionally skip, replace, or merge |
| Per-language comment format | Go uses `//`, JSDoc uses `/** */`, Python uses `"""` | Low | Template selection by language — core feature |
| CLI: `deoxy generate [path]` | Standard CLI UX | Low | cobra subcommand, glob file patterns |
| Config file (`.deoxy.yaml`) | Customize behavior per project | Low | viper, YAML |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Multi-language in one tool | No need to switch between doxdocgen (C/C++), rustdoc (Rust), etc. | Med | Key differentiator. Single binary handles Go, Rust, Python, JS/TS, Java, C/C++. |
| VS Code extension | IDE integration with real-time doc generation | High | Go sidecar binary spawned by extension. JSON-RPC over stdio. |
| Smart text generation | Infer parameter descriptions from parameter names (e.g., `count` → "Number of items") | Med-High | doxdocgen's "generateSmartText" — split camelCase, snake_case into words. |
| Custom template ordering | Users want control over tag order (brief, param, return, etc.) | Med | doxdocgen's `generic.order` is a good model. Each language has a default order. |
| Batch mode (entire project) | Document all functions in a codebase in one command | Low | Walk directory, parse all matching files, generate. |
| Watch mode (`deoxy watch`) | Auto-regenerate on file change | Med | fsnotify + incremental tree-sitter re-parse. |
| Git-aware mode | Only document changed files | Med | `git diff --name-only` to filter targets. |
| `@throws` / `@exception` detection | Identify panics/exceptions from function body | High | Requires deeper AST analysis (find `panic()`, `throw`, `raise` calls). Phase 6+. |
| Deprecation detection | Identify `@deprecated` / `// Deprecated:` patterns | Low | Regex on existing comments. |
| AI-assisted descriptions | Use LLM to generate description text from function body | High | Phase 6+; optional plugin. |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Full documentation site generator | out of scope; tools like godoc, rustdoc, JSDoc already handle rendering | Focus on generating source comments; let other tools render |
| Language server protocol (LSP) | Overkill for a doc generator. Heavy process per language. | Use tree-sitter for parsing, JSON-RPC for VS Code communication |
| AI dependency | Makes tool unreliable without API key, adds cost, latency, nondeterminism | Make AI optional extension, not core feature |
| Support all 40+ tree-sitter grammars | Maintenance burden for niche languages | Focus on 8 major languages; community can contribute more |
| Documentation format conversion (doc → HTML/PDF) | Separately solved problem | Output native doc comments; let existing renderers handle conversion |

## Feature Dependencies

```
Function detection → Parameter extraction → Comment generation → Source insertion
    (Phase 1)          (Phase 1)               (Phase 2)            (Phase 3)

Config system → Language-specific templates
  (Phase 4)        (Phase 2)

CLI framework → All user-facing commands
  (Phase 4)        (Phase 1-3, 5-6)

VS Code extension ← Go sidecar binary ← JSON-RPC interface
  (Phase 5)              (Phase 5)          (Phase 5)
```

## MVP Recommendation

Prioritize:
1. Parse Go, Python, JS/TS files and extract function signatures (Phase 1)
2. Generate GoDoc, JSDoc, Python docstring comments (Phase 2)
3. Insert comments into source files (Phase 3)
4. CLI with `deoxy generate` command (Phase 4)

Defer:
- VS Code extension (Phase 5): Important for adoption but complex
- AI integration: Revisit after MVP validates
- Watch mode: Add in v1.1 after core is solid

## Sources

- doxdocgen VS Code extension — configuration structure, feature scope (HIGH confidence)
- JSDoc official docs — JS/TS doc format (HIGH confidence)
- Go doc comments spec — go.dev/doc/comment (HIGH confidence)
- PEP 257 — Python docstring conventions (HIGH confidence)
- Rust doc comments — doc.rust-lang.org/stable/reference/comments.html (HIGH confidence)
- Oracle Javadoc spec — JDK 26 spec (HIGH confidence)
- Doxygen manual — doxygen.nl (HIGH confidence)
