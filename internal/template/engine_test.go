package template

import (
	"strings"
	"testing"

	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

// ---- Engine Creation Tests ----

func TestNewEngineCompiles(t *testing.T) {
	t.Run("valid templates", func(t *testing.T) {
		e, err := New(map[string]string{
			"go": "{{.Name}} does things.",
		})
		if err != nil {
			t.Fatalf("New with valid template should not error, got: %v", err)
		}
		if e == nil {
			t.Fatal("New should return non-nil Engine")
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		_, err := New(map[string]string{
			"go": "{{.Name",
		})
		if err == nil {
			t.Fatal("New with invalid template should return error")
		}
		if !strings.Contains(err.Error(), "failed to parse") {
			t.Errorf("error should mention parse failure, got: %v", err)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		e, err := New(nil)
		if err != nil {
			t.Fatalf("New with nil map should not error, got: %v", err)
		}
		if e == nil {
			t.Fatal("New with nil map should return non-nil Engine")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		e, err := New(map[string]string{})
		if err != nil {
			t.Fatalf("New with empty map should not error, got: %v", err)
		}
		if e == nil {
			t.Fatal("New with empty map should return non-nil Engine")
		}
	})
}

// ---- Render Tests ----

func TestRenderBasicFunction(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} does things.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:    "DoSomething",
		Kind:    symbol.KindFunction,
		Params:  []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
		Returns: []string{"int"},
	}

	output, err := e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render should not error, got: %v", err)
	}
	if output == "" {
		t.Fatal("Render should return non-empty string")
	}
	if !strings.Contains(output, "DoSomething") {
		t.Errorf("output should contain symbol name, got: %s", output)
	}
	if !strings.Contains(output, "does things") {
		t.Errorf("output should contain template text, got: %s", output)
	}
}

func TestRenderEmptyParams(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} does things.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:    "Foo",
		Kind:    symbol.KindFunction,
		Params:  nil,
		Returns: nil,
	}

	output, err := e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render should not error, got: %v", err)
	}
	if output == "" {
		t.Fatal("Render should return non-empty string")
	}
	if strings.Contains(output, "param") {
		t.Errorf("output should not contain param text, got: %s", output)
	}
}

func TestRenderEmptyReturns(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} does things.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:    "Bar",
		Kind:    symbol.KindFunction,
		Params:  []symbol.Param{{Name: "x", Type: "int"}},
		Returns: nil,
	}

	_, err = e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render should not error, got: %v", err)
	}
}

func TestRenderWithReceiver(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} is a method on {{if .HasReceiver}}{{.Receiver.Type}}{{end}}.",
	})
	if err != nil {
		t.Fatal(err)
	}

	receiver := &symbol.Param{Name: "p", Type: "*Person"}
	info := symbol.SymbolInfo{
		Name:     "Greet",
		Kind:     symbol.KindMethod,
		Receiver: receiver,
		Params:   []symbol.Param{{Name: "name", Type: "string"}},
		Returns:  []string{"string"},
	}

	output, err := e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render should not error, got: %v", err)
	}
	if !strings.Contains(output, "*Person") {
		t.Errorf("output should reference receiver type, got: %s", output)
	}
}

func TestRenderWithNilReceiver(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} works.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:     "Foo",
		Kind:     symbol.KindFunction,
		Receiver: nil,
		Params:   []symbol.Param{{Name: "x", Type: "int"}},
	}

	// Should not panic
	_, err = e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render with nil receiver should not error, got: %v", err)
	}
}

func TestRenderUnknownLanguage(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}}.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name: "Test",
	}

	_, err = e.Render(info, "python")
	if err == nil {
		t.Fatal("Render with unknown language should return error")
	}
	if !strings.Contains(err.Error(), "unknown language") {
		t.Errorf("error should mention 'unknown language', got: %v", err)
	}
}

func TestRenderEmptyName(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}}.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name: "",
	}

	output, err := e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render with empty name should not error, got: %v", err)
	}
	if output != "" {
		t.Errorf("Render with empty name should return empty string, got: %q", output)
	}
}

func TestRenderSpecialCharacters(t *testing.T) {
	e, err := New(map[string]string{
		"go": "// {{.Name}} {{.Brief}}",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:    "__init__",
		Params:  []symbol.Param{{Name: "self", Type: "Self"}},
		Returns: nil,
	}

	output, err := e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render should not error, got: %v", err)
	}
	if !strings.Contains(output, "__init__") {
		t.Errorf("output should preserve special characters in name, got: %s", output)
	}
}

func TestRenderGenericTypeParams(t *testing.T) {
	e, err := New(map[string]string{
		"go": "{{.Name}} works with {{.TypeParams}}.",
	})
	if err != nil {
		t.Fatal(err)
	}

	info := symbol.SymbolInfo{
		Name:       "FirstOrDefault",
		Kind:       symbol.KindFunction,
		Params:     []symbol.Param{{Name: "items", Type: "[]T"}, {Name: "defaultVal", Type: "T"}},
		Returns:    []string{"T"},
		TypeParams: []symbol.Param{{Name: "T", Type: "any"}},
	}

	_, err = e.Render(info, "go")
	if err != nil {
		t.Fatalf("Render with TypeParams should not error, got: %v", err)
	}
}

// ---- Helper Function Tests ----

func TestBriefFunction(t *testing.T) {
	tests := []struct {
		name    string
		params  []symbol.Param
		returns []string
		want    []string // substrings that must appear
		notWant []string // substrings that must NOT appear
	}{
		{
			name:    "Add",
			params:  []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			returns: []string{"int"},
			want:    []string{"adds", "a", "b"},
		},
		{
			name:    "greet",
			params:  []symbol.Param{{Name: "name", Type: "str"}, {Name: "age", Type: "int"}},
			returns: []string{"str"},
			want:    []string{"greets"},
		},
		{
			name:    "DoNothing",
			params:  nil,
			returns: nil,
			// Smart camelCase splitting deferred to Phase 5
			want:    []string{"doNothings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := brief(tt.name, tt.params, tt.returns)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("brief(%q, ...) = %q, want it to contain %q", tt.name, got, w)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("brief(%q, ...) = %q, should NOT contain %q", tt.name, got, nw)
				}
			}
		})
	}
}

func TestParamDescFunction(t *testing.T) {
	t.Run("with valid params", func(t *testing.T) {
		params := []symbol.Param{{Name: "a", Type: "int"}}
		got := paramDesc(params, 0)
		if got == "" {
			t.Error("paramDesc should return non-empty string")
		}
	})

	t.Run("with empty params", func(t *testing.T) {
		got := paramDesc(nil, 0)
		if got != "" {
			t.Errorf("paramDesc with empty params should return empty string, got: %q", got)
		}
	})

	t.Run("with descriptive param name", func(t *testing.T) {
		params := []symbol.Param{{Name: "name", Type: "string"}}
		got := paramDesc(params, 0)
		if !strings.Contains(got, "name") {
			t.Errorf("paramDesc should contain param name, got: %q", got)
		}
	})
}

func TestReturnDescFunction(t *testing.T) {
	t.Run("with returns", func(t *testing.T) {
		got := returnDesc([]string{"int"})
		if got == "" {
			t.Error("returnDesc with returns should return non-empty string")
		}
	})

	t.Run("empty returns", func(t *testing.T) {
		got := returnDesc(nil)
		if got != "" {
			t.Errorf("returnDesc with nil should return empty string, got: %q", got)
		}
	})

	t.Run("multiple returns", func(t *testing.T) {
		got := returnDesc([]string{"int", "error"})
		if !strings.Contains(got, "and") {
			t.Errorf("multiple returns should include 'and', got: %q", got)
		}
	})
}

func TestJoinParams(t *testing.T) {
	tests := []struct {
		params []symbol.Param
		want   string
	}{
		{[]symbol.Param{{Name: "a", Type: "int"}}, "a"},
		{[]symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}}, "a and b"},
		{[]symbol.Param{{Name: "x", Type: "int"}, {Name: "y", Type: "int"}, {Name: "z", Type: "int"}}, "x, y, and z"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := joinParams(tt.params)
			if got != tt.want {
				t.Errorf("joinParams(%v) = %q, want %q", tt.params, got, tt.want)
			}
		})
	}
}

func TestCommentPrefix(t *testing.T) {
	tests := []struct {
		lang string
		want string
	}{
		{"go", "//"},
		{"python", "#"},
		{"rust", "///"},
		{"c", " *"},
		{"cpp", " *"},
		{"unknown", "//"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := commentPrefix(tt.lang)
			if got != tt.want {
				t.Errorf("commentPrefix(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}
