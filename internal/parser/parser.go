// Package parser manages a pool of tree-sitter parsers and provides
// language-aware source code parsing.
//
// WARNING: tree-sitter allocates C heap memory. Always defer obj.Close() for
// every Parser, Tree, Query, and QueryCursor. Never rely on runtime.SetFinalizer.
package parser

import (
	"fmt"
	"sync"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Manager manages a pool of tree-sitter parsers, one per supported language.
// Implements io.Closer.
type Manager struct {
	parsers map[string]*sitter.Parser
	mu      sync.Mutex
}

// NewManager creates a Manager with one tree-sitter parser per supported language.
// It configures each parser with the appropriate language grammar.
// If any grammar fails to initialize, previously opened parsers are closed
// before returning the error.
func NewManager() (*Manager, error) {
	m := &Manager{
		parsers: make(map[string]*sitter.Parser),
	}

	for _, lang := range getSupportedLanguageNames() {
		cfg, ok := GetLanguageConfig(lang)
		if !ok {
			// Should never happen since getSupportedLanguageNames is the source of truth
			m.Close()
			return nil, fmt.Errorf("parser: no config for language %q", lang)
		}

		parser := sitter.NewParser()
		langObj := sitter.NewLanguage(cfg.Grammar())
		if err := parser.SetLanguage(langObj); err != nil {
			parser.Close()
			m.Close()
			return nil, fmt.Errorf("parser: failed to set language %q: %w", lang, err)
		}

		m.parsers[lang] = parser
	}

	return m, nil
}

// Parse parses source code for the given language, returning a new tree.
// The caller is responsible for calling tree.Close() on the returned tree.
// Returns an error if the language is not supported.
//
// The mutex is held across the entire C call to prevent Close() from freeing
// the underlying C parser (ts_parser*) while Parse is using it, which would
// cause a use-after-free in C heap memory. This serializes Parse calls across
// all languages, which is acceptable because tree-sitter C code may not be
// fully thread-safe.
func (m *Manager) Parse(lang string, src []byte) (*sitter.Tree, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parser, ok := m.parsers[lang]
	if !ok {
		return nil, fmt.Errorf("parser: unsupported language %q", lang)
	}

	tree := parser.Parse(src, nil)
	if tree == nil {
		return nil, nil
	}
	return tree, nil
}

// Languages returns a slice of registered language names.
func (m *Manager) Languages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	langs := make([]string, 0, len(m.parsers))
	for lang := range m.parsers {
		langs = append(langs, lang)
	}
	return langs
}

// Close iterates all parsers and calls Close() on each.
// Implements io.Closer.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for lang, parser := range m.parsers {
		parser.Close()
		delete(m.parsers, lang)
	}
	return nil
}

// getSupportedLanguageNames returns the canonical list of supported language names.
func getSupportedLanguageNames() []string {
	return []string{"go", "python", "c", "cpp", "rust"}
}
