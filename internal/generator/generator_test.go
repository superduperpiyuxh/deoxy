package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/superduperpiyuxh/deoxy/internal/config"
)

// fixtureRoot returns the absolute path to the test fixtures directory.
func fixtureRoot() string {
	return filepath.Join("..", "..", "testdata", "fixtures")
}

// skipIfCgoDisabled skips the test if CGO is disabled.
func skipIfCgoDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("skipping test: CGO_ENABLED=0")
	}
}

// TestGenerateGoDoc tests that Go fixture files generate GoDoc-style // comments.
func TestGenerateGoDoc(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "go")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	var goFiles int
	for _, f := range result.Files {
		if f.Language == "go" {
			goFiles++
			if len(f.Comments) == 0 {
				continue
			}
			for _, c := range f.Comments {
				if c.DocComment == "" {
					continue
				}
				// Verify all comment lines start with '//'
				for _, line := range strings.Split(c.DocComment, "\n") {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
						t.Errorf("Go comment line should start with '//': %q in file %s", line, f.Path)
					}
				}
			}
		}
	}

	if goFiles == 0 {
		t.Error("expected at least 1 Go file in results")
	}
	if result.SymbolCount == 0 {
		t.Error("expected at least 1 symbol across all files")
	}
}

// TestGeneratePythonDoc tests that Python fixture files generate Google-style docstrings.
func TestGeneratePythonDoc(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "python")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	var pyFiles int
	for _, f := range result.Files {
		if f.Language == "python" {
			pyFiles++
			for _, c := range f.Comments {
				if c.DocComment == "" {
					continue
				}
				// Verify docstring delimiters
				if !strings.HasPrefix(c.DocComment, "\"\"\"") {
					t.Errorf("Python comment should start with '\"\"\"', got: %q", c.DocComment[:10])
				}
				// Verify Args section (for functions with params)
				if len(c.Symbol.Params) > 0 && !strings.Contains(c.DocComment, "Args:") {
					t.Errorf("Python comment for %q should contain 'Args:' section", c.Symbol.Name)
				}
				// Verify Returns section (for functions with return type)
				if len(c.Symbol.Returns) > 0 && !strings.Contains(c.DocComment, "Returns:") {
					t.Errorf("Python comment for %q should contain 'Returns:' section", c.Symbol.Name)
				}
			}
		}
	}

	if pyFiles == 0 {
		t.Error("expected at least 1 Python file in results")
	}
}

// TestGenerateCDoxygen tests that C fixture files generate Doxygen-style comments.
func TestGenerateCDoxygen(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "c")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	var cFiles int
	for _, f := range result.Files {
		if f.Language == "c" {
			cFiles++
			for _, c := range f.Comments {
				if c.DocComment == "" {
					continue
				}
				// Verify Doxygen structure
				if !strings.HasPrefix(c.DocComment, "/**") {
					t.Errorf("C comment should start with '/**', got: %q", c.DocComment[:10])
				}
				if !strings.Contains(c.DocComment, "@brief") {
					t.Errorf("C comment for %q should contain '@brief'", c.Symbol.Name)
				}
				// Verify @param for functions with parameters
				if len(c.Symbol.Params) > 0 && !strings.Contains(c.DocComment, "@param") {
					t.Errorf("C comment for %q should contain '@param'", c.Symbol.Name)
				}
			}
		}
	}

	if cFiles == 0 {
		t.Error("expected at least 1 C file in results")
	}
}

// TestGenerateCppDoxygen tests that C++ fixture files generate Doxygen-style comments.
func TestGenerateCppDoxygen(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "cpp")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	var cppFiles int
	for _, f := range result.Files {
		if f.Language == "cpp" {
			cppFiles++
			for _, c := range f.Comments {
				if c.DocComment == "" {
					continue
				}
				// Verify Doxygen structure
				if !strings.HasPrefix(c.DocComment, "/**") {
					t.Errorf("C++ comment should start with '/**'")
				}
				if !strings.Contains(c.DocComment, "@brief") {
					t.Errorf("C++ comment for %q should contain '@brief'", c.Symbol.Name)
				}
			}
		}
	}

	if cppFiles == 0 {
		t.Error("expected at least 1 C++ file in results")
	}
}

// TestGenerateRustdoc tests that Rust fixture files generate Rustdoc-style comments.
func TestGenerateRustdoc(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "rust")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	var rustFiles int
	for _, f := range result.Files {
		if f.Language == "rust" {
			rustFiles++
			for _, c := range f.Comments {
				if c.DocComment == "" {
					continue
				}
				// Verify all lines start with '///'
				for _, line := range strings.Split(c.DocComment, "\n") {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" && !strings.HasPrefix(trimmed, "///") {
						t.Errorf("Rust comment line should start with '///': %q", line)
					}
				}
			}
		}
	}

	if rustFiles == 0 {
		t.Error("expected at least 1 Rust file in results")
	}
}

// TestGenerateAllLanguages tests that all 5 language fixture directories
// are processed correctly.
func TestGenerateAllLanguages(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{fixtureRoot()})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	// Collect which languages were seen
	seenLangs := make(map[string]bool)
	var totalComments int
	for _, f := range result.Files {
		seenLangs[f.Language] = true
		totalComments += len(f.Comments)
	}

	// Verify all 5 languages are present
	expectedLangs := []string{"go", "python", "c", "cpp", "rust"}
	for _, lang := range expectedLangs {
		if !seenLangs[lang] {
			t.Errorf("expected files for language %q in results", lang)
		}
	}

	if result.SymbolCount == 0 {
		t.Error("expected at least 1 symbol across all files")
	}

	// Verify no parse errors for fixture files (they are valid syntax)
	if result.ErrorCount > 0 {
		t.Errorf("expected 0 parse errors for fixture files, got %d", result.ErrorCount)
		for _, f := range result.Files {
			if f.ParseError != nil {
				t.Logf("  %s: %v", f.Path, f.ParseError)
			}
		}
	}
}

// TestGenerateWithConfigOverrides tests that Go can generate Doxygen-style
// comments when the config is set to use Doxygen for Go.
func TestGenerateWithConfigOverrides(t *testing.T) {
	skipIfCgoDisabled(t)

	// Create a config with Go set to Doxygen style
	cfg := &config.Config{
		DefaultDocStyle: "godoc",
		DefaultTagOrder: "brief-first",
		Languages: map[string]config.LanguageConfig{
			"go": {DocStyle: "doxygen", TagOrder: "brief-first"},
		},
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{filepath.Join(fixtureRoot(), "go")})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount == 0 {
		t.Fatal("expected at least 1 file in result")
	}

	// Verify Go files now produce Doxygen-style comments
	var doxygenCount int
	for _, f := range result.Files {
		if f.Language == "go" {
			for _, c := range f.Comments {
				if strings.HasPrefix(c.DocComment, "/**") {
					doxygenCount++
				}
			}
		}
	}

	if doxygenCount == 0 {
		t.Error("expected at least 1 Doxygen-style comment with Go overridden to doxygen style")
	}
}

// TestGenerateEmptyFileGraceful tests that an empty file produces no errors.
func TestGenerateEmptyFileGraceful(t *testing.T) {
	skipIfCgoDisabled(t)

	// Create a temporary empty .go file
	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.go")
	if err := os.WriteFile(emptyPath, []byte("package empty\n"), 0644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	result, err := gen.Run([]string{emptyPath})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.FileCount != 1 {
		t.Errorf("expected 1 file in result, got %d", result.FileCount)
	}

	for _, f := range result.Files {
		if f.ParseError != nil {
			t.Errorf("expected no parse error for empty file, got: %v", f.ParseError)
		}
		if len(f.Comments) != 0 {
			t.Errorf("expected 0 comments for empty file, got %d", len(f.Comments))
		}
	}
}

// TestGenerateNonExistentDirectory tests error handling for non-existent paths.
func TestGenerateNonExistentDirectory(t *testing.T) {
	skipIfCgoDisabled(t)

	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer gen.Close()

	_, err = gen.Run([]string{"/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}
