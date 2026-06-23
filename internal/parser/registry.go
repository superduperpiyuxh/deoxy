package parser

import (
	_ "embed"
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

//go:embed queries/go.scm
var queryGo string

//go:embed queries/python.scm
var queryPython string

//go:embed queries/c.scm
var queryC string

//go:embed queries/cpp.scm
var queryCpp string

//go:embed queries/rust.scm
var queryRust string

// LanguageConfig holds configuration for a single supported language.
type LanguageConfig struct {
	// Name is the canonical language name, e.g. "go", "python".
	Name string
	// Extensions are the file extensions for this language, e.g. []string{".go"}.
	Extensions []string
	// Grammar is a getter function returning the tree-sitter language pointer.
	Grammar func() unsafe.Pointer
	// QueryContent is the embedded .scm query file content.
	QueryContent string
	// language is a cached *sitter.Language wrapper, initialized once.
	language *sitter.Language
}

// GrammarAsLanguage wraps the grammar pointer as a *sitter.Language for use with queries.
// The result is cached to avoid allocating a new wrapper on every call.
func (c *LanguageConfig) GrammarAsLanguage() *sitter.Language {
	if c.language == nil {
		c.language = sitter.NewLanguage(c.Grammar())
	}
	return c.language
}

// registry is the internal map of language configs, populated at init time.
var registry map[string]*LanguageConfig
var extRegistry map[string]*LanguageConfig

// registerLanguage adds a language to both name-based and extension-based registries.
func registerLanguage(cfg *LanguageConfig) {
	// Pre-cache the *sitter.Language wrapper since the grammar pointer is static.
	cfg.language = sitter.NewLanguage(cfg.Grammar())
	registry[cfg.Name] = cfg
	for _, ext := range cfg.Extensions {
		// First registration wins (handles .h → C, not C++)
		if _, exists := extRegistry[ext]; !exists {
			extRegistry[ext] = cfg
		}
	}
}

// GetLanguageConfig looks up a language configuration by its canonical name
// (e.g., "go", "python"). Returns the config and true if found, nil and false otherwise.
func GetLanguageConfig(lang string) (*LanguageConfig, bool) {
	cfg, ok := registry[lang]
	return cfg, ok
}

// GetConfigForExtension looks up a language configuration by file extension
// (e.g., ".go", ".py", ".h"). Returns the config and true if found, nil and false otherwise.
// Note: ".h" resolves to C by default (registered first); C++ headers use ".hpp".
func GetConfigForExtension(ext string) (*LanguageConfig, bool) {
	cfg, ok := extRegistry[ext]
	return cfg, ok
}

// SupportedLanguages returns all registered language names.
func SupportedLanguages() []string {
	langs := make([]string, 0, len(registry))
	for name := range registry {
		langs = append(langs, name)
	}
	return langs
}

// SupportedExtensions returns all registered file extensions.
func SupportedExtensions() []string {
	extSet := make(map[string]struct{})
	for ext := range extRegistry {
		extSet[ext] = struct{}{}
	}
	exts := make([]string, 0, len(extSet))
	for ext := range extSet {
		exts = append(exts, ext)
	}
	return exts
}

func init() {
	registry = make(map[string]*LanguageConfig)
	extRegistry = make(map[string]*LanguageConfig)

	registerLanguage(&LanguageConfig{
		Name:         "go",
		Extensions:   []string{".go"},
		Grammar:      tree_sitter_go.Language,
		QueryContent: queryGo,
	})
	registerLanguage(&LanguageConfig{
		Name:         "python",
		Extensions:   []string{".py"},
		Grammar:      tree_sitter_python.Language,
		QueryContent: queryPython,
	})
	registerLanguage(&LanguageConfig{
		Name:         "c",
		Extensions:   []string{".c", ".h"},
		Grammar:      tree_sitter_c.Language,
		QueryContent: queryC,
	})
	registerLanguage(&LanguageConfig{
		Name:         "cpp",
		Extensions:   []string{".cpp", ".cc", ".cxx", ".hpp"},
		Grammar:      tree_sitter_cpp.Language,
		QueryContent: queryCpp,
	})
	registerLanguage(&LanguageConfig{
		Name:         "rust",
		Extensions:   []string{".rs"},
		Grammar:      tree_sitter_rust.Language,
		QueryContent: queryRust,
	})
}
