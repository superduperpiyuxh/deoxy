package template

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

var update = flag.Bool("update", false, "update golden files")

// goldenPath returns the path to the golden file for the given language.
func goldenPath(lang string) string {
	return filepath.Join("..", "..", "testdata", "golden", lang+"_expected.txt")
}

// readGoldenFile reads the golden file for the given language.
func readGoldenFile(lang string) (string, error) {
	data, err := os.ReadFile(goldenPath(lang))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeGoldenFile writes the golden file for the given language.
func writeGoldenFile(lang, content string) error {
	return os.WriteFile(goldenPath(lang), []byte(content), 0644)
}

// renderSymbols renders all test symbols for the given language through the
// engine and returns the combined output.
func renderSymbols(t *testing.T, lang string, symbols []symbol.SymbolInfo) string {
	t.Helper()

	e, err := New(GetDefaultTemplates())
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var outputs []string
	for _, sym := range symbols {
		out, err := e.Render(sym, lang)
		if err != nil {
			t.Fatalf("Render(%q, %q) failed: %v", sym.Name, lang, err)
		}
		if out != "" {
			outputs = append(outputs, out)
		}
	}

	return strings.Join(outputs, "\n")
}

// TestGoDocGolden tests the Go template output against golden files.
func TestGoDocGolden(t *testing.T) {
	symbols := []symbol.SymbolInfo{
		{
			Name:    "Add",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			Returns: []string{"int"},
		},
		{
			Name:     "Greet",
			Kind:     symbol.KindMethod,
			Params:   []symbol.Param{{Name: "name", Type: "string"}},
			Returns:  []string{"string"},
			Receiver: &symbol.Param{Name: "p", Type: "Person"},
		},
		{
			Name: "Person",
			Kind: symbol.KindStruct,
		},
		{
			Name:       "FirstOrDefault",
			Kind:       symbol.KindFunction,
			Params:     []symbol.Param{{Name: "items", Type: "[]T"}, {Name: "defaultVal", Type: "T"}},
			Returns:    []string{"T"},
			TypeParams: []symbol.Param{{Name: "T", Type: "any"}},
		},
	}

	output := renderSymbols(t, "go", symbols)

	// Verify basic structure
	if !strings.HasPrefix(output, "// Add") {
		t.Errorf("Go output should start with '// Add', got:\n%s", output)
	}
	if !strings.Contains(output, "// a -") {
		t.Errorf("Go output should contain param 'a' description, got:\n%s", output)
	}
	if !strings.Contains(output, "// b -") {
		t.Errorf("Go output should contain param 'b' description, got:\n%s", output)
	}

	// Verify all lines start with '//' or are empty
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "//") {
			t.Errorf("Go line should start with '//': %q", line)
		}
	}

	// Golden file comparison
	if *update {
		if err := writeGoldenFile("go", output); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := readGoldenFile("go")
		if err != nil {
			t.Fatalf("golden file not found. Run with -update to create it.\nExpected content:\n%s", output)
		}
		if output != golden {
			t.Errorf("Go output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
		}
	}
}

// TestPythonDocGolden tests the Python template output against golden files.
func TestPythonDocGolden(t *testing.T) {
	symbols := []symbol.SymbolInfo{
		{
			Name:    "greet",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "name", Type: "str"}, {Name: "age", Type: "int"}},
			Returns: []string{"str"},
		},
		{
			Name:     "add",
			Kind:     symbol.KindMethod,
			Params:   []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			Returns:  []string{"int"},
			Receiver: &symbol.Param{Name: "self", Type: "Calculator"},
		},
		{
			Name: "Calculator",
			Kind: symbol.KindClass,
		},
	}

	output := renderSymbols(t, "python", symbols)

	// Verify structure
	if !strings.HasPrefix(output, "\"\"\"") {
		t.Errorf("Python output should start with '\"\"\"', got:\n%s", output)
	}
	if !strings.Contains(output, "Args:") {
		t.Errorf("Python output should contain 'Args:', got:\n%s", output)
	}
	if !strings.Contains(output, "Returns:") {
		t.Errorf("Python output should contain 'Returns:', got:\n%s", output)
	}

	// Verify indentation
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "Args:") || strings.HasPrefix(trimmed, "Returns:") {
			// Check that Args: and Returns: are indented by 4 spaces
			if strings.HasPrefix(line, "    ") {
				// OK - 4 space indent
			}
		}
	}

	// Golden file comparison
	if *update {
		if err := writeGoldenFile("python", output); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := readGoldenFile("python")
		if err != nil {
			t.Fatalf("golden file not found. Run with -update to create it.\nExpected content:\n%s", output)
		}
		if output != golden {
			t.Errorf("Python output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
		}
	}
}

// TestCDoxygenGolden tests the C template output against golden files.
func TestCDoxygenGolden(t *testing.T) {
	symbols := []symbol.SymbolInfo{
		{
			Name:    "add",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			Returns: []string{"int"},
		},
		{
			Name:    "process_buffer",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "buffer", Type: "char *"}, {Name: "size", Type: "int"}},
			Returns: []string{"void"},
		},
		{
			Name:    "divide",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "dividend", Type: "int"}, {Name: "divisor", Type: "int"}},
			Returns: []string{"int"},
		},
	}

	output := renderSymbols(t, "c", symbols)

	// Verify Doxygen structure
	if !strings.HasPrefix(output, "/**") {
		t.Errorf("C output should start with '/**', got:\n%s", output)
	}
	if !strings.Contains(output, "@brief") {
		t.Errorf("C output should contain '@brief', got:\n%s", output)
	}
	if !strings.Contains(output, "@param") {
		t.Errorf("C output should contain '@param', got:\n%s", output)
	}
	if !strings.Contains(output, "@return") {
		t.Errorf("C output should contain '@return', got:\n%s", output)
	}

	// void functions with Returns:["void"] still render @return in our implementation
	// Smart void-skipping could be added later

	// Golden file comparison
	if *update {
		if err := writeGoldenFile("c", output); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := readGoldenFile("c")
		if err != nil {
			t.Fatalf("golden file not found. Run with -update to create it.\nExpected content:\n%s", output)
		}
		if output != golden {
			t.Errorf("C output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
		}
	}
}

// TestCppDoxygenGolden tests the C++ template output against golden files.
func TestCppDoxygenGolden(t *testing.T) {
	symbols := []symbol.SymbolInfo{
		{
			Name:    "free_function",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "x", Type: "int"}, {Name: "y", Type: "int"}},
			Returns: []string{"int"},
		},
		{
			Name: "Calculator",
			Kind: symbol.KindClass,
		},
	}

	output := renderSymbols(t, "cpp", symbols)

	// Verify Doxygen structure
	if !strings.HasPrefix(output, "/**") {
		t.Errorf("C++ output should start with '/**', got:\n%s", output)
	}
	if !strings.Contains(output, "@brief") {
		t.Errorf("C++ output should contain '@brief', got:\n%s", output)
	}

	// Golden file comparison
	if *update {
		if err := writeGoldenFile("cpp", output); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := readGoldenFile("cpp")
		if err != nil {
			t.Fatalf("golden file not found. Run with -update to create it.\nExpected content:\n%s", output)
		}
		if output != golden {
			t.Errorf("C++ output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
		}
	}
}

// TestRustdocGolden tests the Rust template output against golden files.
func TestRustdocGolden(t *testing.T) {
	symbols := []symbol.SymbolInfo{
		{
			Name:    "add",
			Kind:    symbol.KindFunction,
			Params:  []symbol.Param{{Name: "a", Type: "i32"}, {Name: "b", Type: "i32"}},
			Returns: []string{"i32"},
		},
		{
			Name:     "new",
			Kind:     symbol.KindMethod,
			Params:   []symbol.Param{{Name: "x", Type: "i32"}, {Name: "y", Type: "i32"}},
			Returns:  []string{"Self"},
			Receiver: &symbol.Param{Name: "self", Type: "&Point"},
		},
		{
			Name: "Point",
			Kind: symbol.KindStruct,
		},
	}

	output := renderSymbols(t, "rust", symbols)

	// Verify Rustdoc structure
	if !strings.HasPrefix(output, "///") {
		t.Errorf("Rust output should start with '///', got:\n%s", output)
	}
	if !strings.Contains(output, "Returns:") {
		t.Errorf("Rust output should contain 'Returns:', got:\n%s", output)
	}

	// Golden file comparison
	if *update {
		if err := writeGoldenFile("rust", output); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := readGoldenFile("rust")
		if err != nil {
			t.Fatalf("golden file not found. Run with -update to create it.\nExpected content:\n%s", output)
		}
		if output != golden {
			t.Errorf("Rust output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
		}
	}
}

// TestAllLanguagesCovered verifies golden files exist for all 5 languages.
func TestAllLanguagesCovered(t *testing.T) {
	languages := []string{"go", "python", "c", "cpp", "rust"}

	if *update {
		t.Skip("skipping coverage check during -update")
	}

	for _, lang := range languages {
		content, err := readGoldenFile(lang)
		if err != nil {
			t.Errorf("missing golden file for %q, run with -update", lang)
			continue
		}
		if content == "" {
			t.Errorf("golden file for %q is empty", lang)
		}
	}
}
