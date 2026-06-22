package writer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper: create a temp .go file with the given content
func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

// Helper: read file content, trimming trailing whitespace per line for comparison
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func TestInsertComment(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		lineIdx int
		comment string
		indent  string
		want    []string
	}{
		{
			name: "simple insert before function",
			lines: []string{
				"package main",
				"",
				"func Hello() int {",
				"    return 42",
				"}",
			},
			lineIdx: 2,
			comment: "// Hello does something",
			indent:  "",
			want: []string{
				"package main",
				"",
				"// Hello does something",
				"func Hello() int {",
				"    return 42",
				"}",
			},
		},
		{
			name: "insert indented comment",
			lines: []string{
				"type Foo struct {",
				"    Bar() int",
				"}",
			},
			lineIdx: 1,
			comment: "// Bar does bar stuff",
			indent:  "    ",
			want: []string{
				"type Foo struct {",
				"    // Bar does bar stuff",
				"    Bar() int",
				"}",
			},
		},
		{
			name: "insert multi-line comment",
			lines: []string{
				"package main",
				"",
				"func Add(a, b int) int {",
				"    return a + b",
				"}",
			},
			lineIdx: 2,
			comment: "// Add adds a and b and returns the result.\n//\n// a - the first operand\n// b - the second operand",
			indent:  "",
			want: []string{
				"package main",
				"",
				"// Add adds a and b and returns the result.",
				"//",
				"// a - the first operand",
				"// b - the second operand",
				"func Add(a, b int) int {",
				"    return a + b",
				"}",
			},
		},
		{
			name: "adds leading blank line when adjacent to code",
			lines: []string{
				"package main",
				"var x = 1",
				"func Hello() int {",
				"    return x",
				"}",
			},
			lineIdx: 2,
			comment: "// Hello does a thing",
			indent:  "",
			want: []string{
				"package main",
				"var x = 1",
				"",
				"// Hello does a thing",
				"func Hello() int {",
				"    return x",
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertComment(tt.lines, tt.lineIdx, tt.comment, tt.indent)
			if len(got) != len(tt.want) {
				t.Fatalf("len(got)=%d, want %d\ngot:  %#v\nwant: %#v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d:\ngot:  %q\nwant: %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestHasExistingComment(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		lineIdx int
		want    bool
	}{
		{
			name: "no comment above",
			lines: []string{
				"package main",
				"",
				"func Foo() {}",
			},
			lineIdx: 2,
			want:    false,
		},
		{
			name: "Go comment above",
			lines: []string{
				"package main",
				"",
				"// Foo does a thing",
				"func Foo() {}",
			},
			lineIdx: 3,
			want:    true,
		},
		{
			name: "Rust doc comment",
			lines: []string{
				"/// Foo does a thing",
				"fn foo() {}",
			},
			lineIdx: 1,
			want:    true,
		},
		{
			name: "block comment",
			lines: []string{
				"/**",
				" * @brief does a thing",
				" */",
				"int foo(void);",
			},
			lineIdx: 3,
			want:    true,
		},
		{
			name: "Python comment",
			lines: []string{
				"",
				"# this is a comment",
				"def foo():",
			},
			lineIdx: 2,
			want:    true,
		},
		{
			name: "Python docstring",
			lines: []string{
				"",
				"\"\"\"Does something.\"\"\"",
				"def foo():",
			},
			lineIdx: 2,
			want:    true,
		},
		{
			name: "blank lines before function",
			lines: []string{
				"package main",
				"",
				"",
				"func Foo() {}",
			},
			lineIdx: 3,
			want:    false,
		},
		{
			name: "lineIdx 0 returns false",
			lines: []string{
				"func Foo() {}",
			},
			lineIdx: 0,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasExistingComment(tt.lines, tt.lineIdx)
			if got != tt.want {
				t.Errorf("hasExistingComment(%v) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestExtractIndent(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"no indent", "func Foo()", ""},
		{"4 spaces", "    func Foo()", "    "},
		{"tab", "\tfunc Foo()", "\t"},
		{"8 spaces", "        func Foo()", "        "},
		{"empty", "", ""},
		{"mixed", "\t  func Foo()", "\t  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIndent(tt.line)
			if got != tt.want {
				t.Errorf("extractIndent(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestToLinesAndFromLines(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantLines     int
		wantSep       string
		wantTrailing  bool
	}{
		{"LF no trailing", "line1\nline2\nline3", 3, "\n", false},
		{"LF with trailing", "line1\nline2\nline3\n", 3, "\n", true},
		{"CRLF no trailing", "line1\r\nline2\r\nline3", 3, "\r\n", false},
		{"CRLF with trailing", "line1\r\nline2\r\nline3\r\n", 3, "\r\n", true},
		{"empty", "", 0, "\n", false},
		{"single line", "hello", 1, "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, sep, trailing := toLines([]byte(tt.input))
			if len(lines) != tt.wantLines {
				t.Errorf("len(lines)=%d, want %d", len(lines), tt.wantLines)
			}
			if sep != tt.wantSep {
				t.Errorf("sep=%q, want %q", sep, tt.wantSep)
			}
			if trailing != tt.wantTrailing {
				t.Errorf("trailing=%v, want %v", trailing, tt.wantTrailing)
			}

			// Roundtrip
			got := fromLines(lines, sep, trailing)
			if string(got) != tt.input {
				t.Errorf("roundtrip:\ngot:  %q\nwant: %q", string(got), tt.input)
			}
		})
	}
}

func TestIntegrationInsertAndDetect(t *testing.T) {
	// Full integration: write a Go file, run hasExistingComment + insertComment
	dir := t.TempDir()

	// File without doc comment
	src := "package main\n\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	path := writeGoFile(t, dir, "main.go", src)

	// Read it back
	got := readFile(t, path)
	if got != src {
		t.Fatalf("file content mismatch:\ngot:  %q\nwant: %q", got, src)
	}

	// Test hasExistingComment on the raw source
	lines, sep, trailing := toLines([]byte(src))
	if hasExistingComment(lines, 2) {
		t.Error("expected no existing comment before Add")
	}

	// Insert a comment
	comment := "// Add adds a and b and returns the result."
	modifiedLines := insertComment(lines, 2, comment, "")
	modified := fromLines(modifiedLines, sep, trailing)

	// Verify insertion
	expected := "package main\n\n// Add adds a and b and returns the result.\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	if string(modified) != expected {
		t.Errorf("insert result:\ngot:  %q\nwant: %q", string(modified), expected)
	}

	// Now hasExistingComment should detect it
	lines2, _, _ := toLines(modified)
	if !hasExistingComment(lines2, 3) {
		t.Error("expected existing comment after insertion")
	}
}

func TestBuildDirectivesPreserved(t *testing.T) {
	// Verify that //go:build directives are preserved above generated comments
	dir := t.TempDir()

	src := "//go:build linux\n\npackage main\n\nfunc Foo() {}\n"
	writeGoFile(t, dir, "main.go", src)

	// Read lines
	lines, sep, trailing := toLines([]byte(src))

	// hasExistingComment on Foo (line 4) should see the empty line above it
	if hasExistingComment(lines, 4) {
		t.Error("expected no comment before Foo")
	}

	// Insert a comment before Foo
	modifiedLines := insertComment(lines, 4, "// Foo does something", "")
	modified := fromLines(modifiedLines, sep, trailing)

	// Verify the build directive is preserved and comment is inserted
	if !strings.Contains(string(modified), "//go:build linux") {
		t.Error("build directive removed after insertion")
	}
	if !strings.Contains(string(modified), "// Foo does something") {
		t.Error("comment not inserted")
	}

	// Verify the order: build directive, blank line, package, blank line, comment, function
	expected := "//go:build linux\n\npackage main\n\n// Foo does something\nfunc Foo() {}\n"
	if string(modified) != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", string(modified), expected)
	}
}

func TestCRLFPreserved(t *testing.T) {
	dir := t.TempDir()

	src := "package main\r\n\r\nfunc Hello() int {\r\n    return 42\r\n}\r\n"
	writeGoFile(t, dir, "main.go", src)

	// toLines should detect CRLF
	lines, sep, trailing := toLines([]byte(src))
	if sep != "\r\n" {
		t.Errorf("expected CRLF separator, got %q", sep)
	}
	if !trailing {
		t.Error("expected trailing newline")
	}

	// Insert a comment
	comment := "// Hello does a thing"
	modifiedLines := insertComment(lines, 2, comment, "")
	modified := fromLines(modifiedLines, sep, trailing)

	// Verify CRLF is preserved
	if !bytes.Contains(modified, []byte("\r\n")) {
		t.Error("CRLF not preserved in output")
	}

	// Verify correct content
	expected := "package main\r\n\r\n// Hello does a thing\r\nfunc Hello() int {\r\n    return 42\r\n}\r\n"
	if string(modified) != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", string(modified), expected)
	}
}
