// Package generator is the top-level orchestrator that drives the deoxy
// pipeline: load config, scan directories for source files, parse files
// with tree-sitter, render doc comments via the template engine, and
// return structured results.
//
// The generator does NOT write to files (that is Phase 3 — CLI integration).
// It returns structured results that Phase 3's CLI commands can use for
// file injection, diff display, and dry-run modes.
package generator

import (
	"fmt"
	"os"
	"sort"

	"github.com/superduperpiyuxh/deoxy/internal/config"
	"github.com/superduperpiyuxh/deoxy/internal/parser"
	"github.com/superduperpiyuxh/deoxy/internal/scanner"
	"github.com/superduperpiyuxh/deoxy/internal/symbol"
	"github.com/superduperpiyuxh/deoxy/internal/template"
)

// SymbolComment represents a single generated doc comment for a symbol
// at a specific source location.
type SymbolComment struct {
	// Symbol is the parsed symbol this comment was generated for.
	Symbol symbol.SymbolInfo
	// DocComment is the generated doc comment text, including comment delimiters.
	DocComment string
	// StartLine is the 0-indexed line where this comment should be inserted
	// (before the symbol).
	StartLine int
	// EndLine is the 0-indexed line where the symbol definition ends.
	EndLine int
}

// GeneratedFileResult holds results for a single source file.
type GeneratedFileResult struct {
	// Path is the absolute file path.
	Path string
	// Language is the canonical language identifier.
	Language string
	// Source is the original source bytes (for Phase 3 injection).
	Source []byte
	// Comments are the generated comments per symbol, in source order.
	Comments []SymbolComment
	// ParseError is non-nil if parsing failed for this file.
	ParseError error
}

// GeneratorResult holds complete pipeline results.
type GeneratorResult struct {
	// Files is the list of results per file.
	Files []GeneratedFileResult
	// FileCount is the total number of files processed.
	FileCount int
	// SymbolCount is the total number of symbols documented across all files.
	SymbolCount int
	// ErrorCount is the number of files with errors.
	ErrorCount int
}

// Generator is the top-level orchestrator for the deoxy pipeline.
// It wires together config, scanner, parser, and template engine.
type Generator struct {
	config *config.Config
	parser *parser.Manager
	engine *template.Engine
}

// New creates a Generator with the given configuration.
// It initializes the parser manager and template engine with default templates.
// Returns an error if the parser manager or template engine initialization fails.
func New(cfg *config.Config) (*Generator, error) {
	if cfg == nil {
		cfg = config.LoadDefaultConfig()
	}

	pm, err := parser.NewManager()
	if err != nil {
		return nil, fmt.Errorf("generator: failed to create parser manager: %w", err)
	}

	eng, err := template.New(template.GetDefaultTemplates())
	if err != nil {
		pm.Close()
		return nil, fmt.Errorf("generator: failed to create template engine: %w", err)
	}

	return &Generator{
		config: cfg,
		parser: pm,
		engine: eng,
	}, nil
}

// Run executes the full pipeline: scan → parse → render.
// It returns structured results without modifying any files.
func (g *Generator) Run(paths []string) (*GeneratorResult, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("generator: no paths provided")
	}

	// Step 1: Scan
	scanResult, err := scanner.Scan(paths...)
	if err != nil {
		return nil, fmt.Errorf("generator: scan failed: %w", err)
	}

	if len(scanResult.Files) == 0 {
		return &GeneratorResult{}, nil
	}

	// Step 2: Parse and render each file sequentially
	// Sequential processing is used because tree-sitter parsers are not
	// thread-safe. Parallel processing can be added in Phase 3 when
	// per-file parser instances are introduced.

	result := &GeneratorResult{}
	var fileResults []GeneratedFileResult

	for _, entry := range scanResult.Files {
		res := g.processFile(entry)
		fileResults = append(fileResults, res)
		if res.ParseError != nil {
			result.ErrorCount++
		}
		result.SymbolCount += len(res.Comments)
	}

	// Sort files by path for deterministic output
	sort.Slice(fileResults, func(i, j int) bool {
		return fileResults[i].Path < fileResults[j].Path
	})

	result.Files = fileResults
	result.FileCount = len(fileResults)

	return result, nil
}

// processFile handles a single file: read source, parse, run query, render.
func (g *Generator) processFile(entry scanner.FileEntry) GeneratedFileResult {
	res := GeneratedFileResult{
		Path:     entry.Path,
		Language: entry.Language,
	}

	// Read source
	src, err := os.ReadFile(entry.Path)
	if err != nil {
		res.ParseError = fmt.Errorf("generator: failed to read %q: %w", entry.Path, err)
		return res
	}
	res.Source = src

	// Parse
	tree, err := g.parser.Parse(entry.Language, src)
	if err != nil {
		res.ParseError = fmt.Errorf("generator: failed to parse %q: %w", entry.Path, err)
		return res
	}
	if tree == nil {
		// Empty file or tree-sitter returned nil
		return res
	}
	defer tree.Close()

	// Get query content for this language
	langCfg, ok := parser.GetLanguageConfig(entry.Language)
	if !ok {
		res.ParseError = fmt.Errorf("generator: unsupported language %q", entry.Language)
		return res
	}

	// Run query to extract symbols
	symbols, err := parser.RunQuery(tree, langCfg.QueryContent, entry.Language, src)
	if err != nil {
		res.ParseError = fmt.Errorf("generator: query failed for %q: %w", entry.Path, err)
		return res
	}

	// Determine which template to use based on config
	templateKey := g.templateKeyForLanguage(entry.Language)

	// Render each symbol
	var comments []SymbolComment
	for _, sym := range symbols {
		doc, err := g.engine.Render(sym, templateKey)
		if err != nil {
			// Skip symbols that fail to render
			continue
		}
		if doc == "" {
			continue
		}

		comments = append(comments, SymbolComment{
			Symbol:     sym,
			DocComment: doc,
			StartLine:  sym.StartLine,
			EndLine:    sym.EndLine,
		})
	}

	res.Comments = comments
	return res
}

// templateKeyForLanguage returns the appropriate template key for a language
// based on the generator's configuration.
func (g *Generator) templateKeyForLanguage(lang string) string {
	// Map config docstyle to template key
	docStyle := g.config.GetDocStyle(lang)

	switch docStyle {
	case config.DocStyleGoDoc:
		return "go"
	case config.DocStyleDoxygen:
		// For C and C++, use the language-specific template
		if lang == "c" || lang == "cpp" {
			return lang
		}
		// For other languages configured to use Doxygen, use the generic 'c' template
		return "c"
	case config.DocStylePyDoc:
		return "python"
	case config.DocStyleRustdoc:
		return "rust"
	default:
		// Fall back to language-specific default
		return lang
	}
}

// ProcessContent parses and renders doc comments for in-memory content.
// Used by the LSP server to process files opened in the editor without writing to disk.
func (g *Generator) ProcessContent(path string, content []byte, lang string) ([]SymbolComment, error) {
	tree, err := g.parser.Parse(lang, content)
	if err != nil {
		return nil, fmt.Errorf("generator: parse failed for %q: %w", path, err)
	}
	if tree == nil {
		return nil, nil
	}
	defer tree.Close()

	langCfg, ok := parser.GetLanguageConfig(lang)
	if !ok {
		return nil, fmt.Errorf("generator: unsupported language %q", lang)
	}

	symbols, err := parser.RunQuery(tree, langCfg.QueryContent, lang, content)
	if err != nil {
		return nil, fmt.Errorf("generator: query failed for %q: %w", path, err)
	}

	templateKey := g.templateKeyForLanguage(lang)

	var comments []SymbolComment
	for _, sym := range symbols {
		doc, err := g.engine.Render(sym, templateKey)
		if err != nil {
			continue
		}
		if doc == "" {
			continue
		}
		comments = append(comments, SymbolComment{
			Symbol:     sym,
			DocComment: doc,
			StartLine:  sym.StartLine,
			EndLine:    sym.EndLine,
		})
	}

	return comments, nil
}

// Close releases all parser resources.
// Implements io.Closer.
func (g *Generator) Close() error {
	return g.parser.Close()
}
