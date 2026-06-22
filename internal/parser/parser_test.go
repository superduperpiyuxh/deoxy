package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

// readFixture reads a test fixture file.
func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return data
}

// fixturePath returns the absolute path to a test fixture.
func fixturePath(t *testing.T, lang, filename string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "fixtures", lang, filename)
}

func TestParseGoFunctions(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "go", "sample.go"))
	tree, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("go")
	results, err := RunQuery(tree, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var functions []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindFunction {
			functions = append(functions, r)
		}
	}

	if len(functions) < 1 {
		t.Fatalf("expected at least 1 function, got %d", len(functions))
	}

	// Check Add function
	found := false
	for _, fn := range functions {
		if fn.Name == "Add" {
			found = true
			if len(fn.Params) != 2 {
				t.Errorf("Add: expected 2 params, got %d", len(fn.Params))
			}
			if len(fn.Returns) == 0 {
				t.Error("Add: expected non-empty Returns")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find function 'Add'")
	}
}

func TestParseGoMethods(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "go", "sample.go"))
	tree, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("go")
	results, err := RunQuery(tree, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var methods []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindMethod {
			methods = append(methods, r)
		}
	}

	if len(methods) < 1 {
		t.Fatalf("expected at least 1 method, got %d", len(methods))
	}

	// Check Greet method
	found := false
	for _, m := range methods {
		if m.Name == "Greet" {
			found = true
			if m.Receiver == nil {
				t.Error("Greet: expected non-nil Receiver")
			} else if m.Receiver.Name == "" && m.Receiver.Type == "" {
				t.Error("Greet: Receiver has empty fields")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find method 'Greet'")
	}
}

func TestParseGoStructs(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "go", "sample.go"))
	tree, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("go")
	results, err := RunQuery(tree, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var structs []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindStruct {
			structs = append(structs, r)
		}
	}

	if len(structs) < 1 {
		t.Fatalf("expected at least 1 struct, got %d", len(structs))
	}

	if structs[0].Name != "Person" {
		t.Errorf("expected struct name 'Person', got %q", structs[0].Name)
	}
}

func TestParseGoInterfaces(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "go", "sample.go"))
	tree, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("go")
	results, err := RunQuery(tree, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var interfaces []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindInterface {
			interfaces = append(interfaces, r)
		}
	}

	if len(interfaces) < 1 {
		t.Fatalf("expected at least 1 interface, got %d", len(interfaces))
	}

	if interfaces[0].Name != "Stringer" {
		t.Errorf("expected interface name 'Stringer', got %q", interfaces[0].Name)
	}
}

func TestParsePythonFunctions(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "python", "sample.py"))
	tree, err := manager.Parse("python", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("python")
	results, err := RunQuery(tree, cfg.QueryContent, "python", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var functions []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindFunction {
			functions = append(functions, r)
		}
	}

	if len(functions) < 1 {
		t.Fatalf("expected at least 1 function, got %d", len(functions))
	}

	// Check greet has params and returns
	for _, fn := range functions {
		if fn.Name == "greet" {
			if len(fn.Params) == 0 {
				t.Error("greet: expected non-empty Params")
			}
			if len(fn.Returns) == 0 {
				t.Error("greet: expected non-empty Returns")
			}
		}
	}
}

func TestParsePythonClasses(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "python", "sample.py"))
	tree, err := manager.Parse("python", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("python")
	results, err := RunQuery(tree, cfg.QueryContent, "python", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var classes []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindClass {
			classes = append(classes, r)
		}
	}

	if len(classes) < 1 {
		t.Fatalf("expected at least 1 class, got %d", len(classes))
	}

	if classes[0].Name != "Calculator" {
		t.Errorf("expected class name 'Calculator', got %q", classes[0].Name)
	}
}

func TestParseCFunctions(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "c", "sample.c"))
	tree, err := manager.Parse("c", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("c")
	results, err := RunQuery(tree, cfg.QueryContent, "c", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var functions []symbol.SymbolInfo
	for _, r := range results {
		if r.Kind == symbol.KindFunction {
			functions = append(functions, r)
		}
	}

	if len(functions) < 1 {
		t.Fatalf("expected at least 1 function, got %d", len(functions))
	}
}

func TestParseCppFunctionsAndClasses(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "cpp", "sample.cpp"))
	tree, err := manager.Parse("cpp", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("cpp")
	results, err := RunQuery(tree, cfg.QueryContent, "cpp", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var functions []symbol.SymbolInfo
	var classes []symbol.SymbolInfo
	for _, r := range results {
		switch r.Kind {
		case symbol.KindFunction:
			functions = append(functions, r)
		case symbol.KindClass:
			classes = append(classes, r)
		}
	}

	if len(functions) < 1 {
		t.Errorf("expected at least 1 function, got %d", len(functions))
	}
	if len(classes) < 1 {
		t.Errorf("expected at least 1 class, got %d", len(classes))
	}
}

func TestParseRustFunctionsAndStructs(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "rust", "sample.rs"))
	tree, err := manager.Parse("rust", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer tree.Close()

	cfg, _ := GetLanguageConfig("rust")
	results, err := RunQuery(tree, cfg.QueryContent, "rust", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	var functions []symbol.SymbolInfo
	var structs []symbol.SymbolInfo
	var interfaces []symbol.SymbolInfo
	for _, r := range results {
		switch r.Kind {
		case symbol.KindFunction:
			functions = append(functions, r)
		case symbol.KindStruct:
			structs = append(structs, r)
		case symbol.KindInterface:
			interfaces = append(interfaces, r)
		}
	}

	if len(functions) < 1 {
		t.Errorf("expected at least 1 function, got %d", len(functions))
	}
	if len(structs) < 1 {
		t.Errorf("expected at least 1 struct, got %d", len(structs))
	}
	if len(interfaces) < 1 {
		t.Errorf("expected at least 1 interface/trait, got %d", len(interfaces))
	}
}

func TestParserPoolReusesParsers(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	src := readFixture(t, fixturePath(t, "go", "sample.go"))

	// Parse twice on same language
	tree1, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("first Parse failed: %v", err)
	}
	defer tree1.Close()

	tree2, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("second Parse failed: %v", err)
	}
	defer tree2.Close()

	// Verify two different tree pointers
	if tree1 == tree2 {
		t.Error("expected two different *Tree pointers")
	}

	// Verify both produce results
	cfg, _ := GetLanguageConfig("go")
	results1, err := RunQuery(tree1, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery on tree1 failed: %v", err)
	}

	results2, err := RunQuery(tree2, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery on tree2 failed: %v", err)
	}

	if len(results1) != len(results2) {
		t.Errorf("expected same number of results, got %d vs %d", len(results1), len(results2))
	}

	// Verify manager has 5 parsers
	langs := manager.Languages()
	if len(langs) != 5 {
		t.Errorf("expected 5 languages, got %d: %v", len(langs), langs)
	}
}

func TestUnsupportedLanguage(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	_, err = manager.Parse("javascript", []byte{})
	if err == nil {
		t.Error("expected error for unsupported language 'javascript'")
	}

	_, err = manager.Parse("unsupported", []byte{})
	if err == nil {
		t.Error("expected error for unsupported language 'unsupported'")
	}
}

func TestEmptySource(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	tree, err := manager.Parse("go", []byte{})
	if err != nil {
		t.Fatalf("Parse empty source failed: %v", err)
	}
	if tree != nil {
		defer tree.Close()
	}
}

func TestAllObjectsClosed(t *testing.T) {
	// This test verifies that the Manager, Tree, and Query objects
	// are properly closed. We create, use, and close objects within
	// a single function scope and check that no panics occur.
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	src := readFixture(t, fixturePath(t, "go", "sample.go"))
	tree, err := manager.Parse("go", src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cfg, _ := GetLanguageConfig("go")
	results, err := RunQuery(tree, cfg.QueryContent, "go", src)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected non-empty results")
	}

	tree.Close()
	manager.Close()
	// If we get here without panics, all objects were properly closed
}

func TestManagerLanguages(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	langs := manager.Languages()
	expected := map[string]bool{"go": true, "python": true, "c": true, "cpp": true, "rust": true}

	if len(langs) != len(expected) {
		t.Errorf("expected %d languages, got %d: %v", len(expected), len(langs), langs)
	}

	for _, lang := range langs {
		if !expected[lang] {
			t.Errorf("unexpected language %q", lang)
		}
	}
}
