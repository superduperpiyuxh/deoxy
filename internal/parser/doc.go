// Package parser wraps github.com/tree-sitter/go-tree-sitter to provide
// per-language AST parsing with explicit Close() discipline.
//
// It owns the parser pool, tree-sitter query execution, and the extraction
// of typed SymbolInfo structs from source files. Full implementation begins
// in Phase 1.
package parser
