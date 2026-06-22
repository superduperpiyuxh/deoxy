package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempDir creates a temporary directory for test setup.
// Returns the path and a cleanup function.
func createTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// writeFile writes content to a file at path, creating parent dirs as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestDiscoverGoFiles(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.go"), "package main\n")
	writeFile(t, filepath.Join(dir, "helper.go"), "package helper\n")

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}

	for _, f := range result.Files {
		if f.Language != "go" {
			t.Errorf("expected language 'go' for %s, got %q", f.Path, f.Language)
		}
	}
}

func TestDiscoverAllLanguages(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	files := map[string]string{
		"main.go":     "package main\n",
		"main.py":     "def foo(): pass\n",
		"main.c":      "int main() { return 0; }\n",
		"main.cpp":    "int main() { return 0; }\n",
		"main.rs":     "fn main() {}\n",
	}

	for name, content := range files {
		writeFile(t, filepath.Join(dir, name), content)
	}

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 5 {
		t.Errorf("expected 5 files, got %d", len(result.Files))
	}

	if len(result.ByLanguage) != 5 {
		t.Errorf("expected 5 language groups, got %d", len(result.ByLanguage))
	}
}

func TestSkipsHidden(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, ".hidden", "main.go"), "package main\n")
	writeFile(t, filepath.Join(dir, "visible", "main.py"), "def foo(): pass\n")

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.Files))
	}

	if len(result.Files) > 0 && result.Files[0].Language != "python" {
		t.Errorf("expected python file, got %q", result.Files[0].Language)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		wantLang string
		wantOk   bool
	}{
		{"main.go", "go", true},
		{"main.py", "python", true},
		{"main.rs", "rust", true},
		{"main.c", "c", true},
		{"main.cpp", "cpp", true},
		{"main.h", "c", true},
		{"main.hpp", "cpp", true},
		{"main.cc", "cpp", true},
		{"main.js", "", false},
		{"main.ts", "", false},
		{"Makefile", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			gotLang, gotOk := DetectLanguage(tt.path)
			if gotOk != tt.wantOk {
				t.Errorf("DetectLanguage(%q) ok = %v, want %v", tt.path, gotOk, tt.wantOk)
			}
			if gotLang != tt.wantLang {
				t.Errorf("DetectLanguage(%q) lang = %q, want %q", tt.path, gotLang, tt.wantLang)
			}
		})
	}
}

func TestUnknownExtensionSkipped(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.js"), "console.log('hi');\n")
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n")

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file (.go only), got %d", len(result.Files))
	}
}

func TestDeduplication(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.go"), "package main\n")

	// Scan with both relative and absolute paths
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("Abs failed: %v", err)
	}

	result, err := Scan(dir, absDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	goFiles := 0
	for _, f := range result.Files {
		if f.Language == "go" {
			goFiles++
		}
	}

	if goFiles != 1 {
		t.Errorf("expected 1 go file (deduplicated), got %d", goFiles)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 total file (deduplicated), got %d", len(result.Files))
	}
}

func TestNonExistentDirectory(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/definitely/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestByLanguageGrouping(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	files := map[string]string{
		"main.go":  "package main\n",
		"main.py":  "def foo(): pass\n",
		"main.c":   "int main() { return 0; }\n",
		"main.cpp": "int main() { return 0; }\n",
		"main.rs":  "fn main() {}\n",
	}

	for name, content := range files {
		writeFile(t, filepath.Join(dir, name), content)
	}

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expectedLangs := map[string]int{
		"go":     1,
		"python": 1,
		"c":      1,
		"cpp":    1,
		"rust":   1,
	}

	for lang, expectedCount := range expectedLangs {
		group, ok := result.ByLanguage[lang]
		if !ok {
			t.Errorf("expected language group %q in ByLanguage", lang)
			continue
		}
		if len(group) != expectedCount {
			t.Errorf("ByLanguage[%q] has %d files, want %d", lang, len(group), expectedCount)
		}
	}
}

func TestEmptyDirectory(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 0 {
		t.Errorf("expected 0 files in empty directory, got %d", len(result.Files))
	}
	if len(result.ByLanguage) != 0 {
		t.Errorf("expected 0 language groups in empty directory, got %d", len(result.ByLanguage))
	}
}

func TestDirectoryWithOnlyUnsupportedFiles(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.js"), "console.log('hi');\n")
	writeFile(t, filepath.Join(dir, "main.ts"), "console.log('hi');\n")
	writeFile(t, filepath.Join(dir, "main.css"), "body {}\n")

	result, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Files) != 0 {
		t.Errorf("expected 0 files for unsupported extensions, got %d", len(result.Files))
	}
}

func TestScanWithExtensions(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.go"), "package main\n")
	writeFile(t, filepath.Join(dir, "main.py"), "def foo(): pass\n")
	writeFile(t, filepath.Join(dir, "main.c"), "int main() { return 0; }\n")

	result, err := ScanWithExtensions([]string{dir}, []string{".go", ".py"})
	if err != nil {
		t.Fatalf("ScanWithExtensions failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("expected 2 files (.go, .py), got %d", len(result.Files))
	}

	for _, f := range result.Files {
		if f.Language != "go" && f.Language != "python" {
			t.Errorf("unexpected language %q, expected 'go' or 'python'", f.Language)
		}
	}
}

func TestScanGoSpecificExtension(t *testing.T) {
	dir, cleanup := createTempDir(t)
	defer cleanup()

	writeFile(t, filepath.Join(dir, "main.go"), "package main\n")
	writeFile(t, filepath.Join(dir, "main.py"), "def foo(): pass\n")

	result, err := ScanWithExtensions([]string{dir}, []string{".go"})
	if err != nil {
		t.Fatalf("ScanWithExtensions failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file (.go only), got %d", len(result.Files))
	}

	if len(result.Files) > 0 && result.Files[0].Language != "go" {
		t.Errorf("expected 'go' language, got %q", result.Files[0].Language)
	}
}
