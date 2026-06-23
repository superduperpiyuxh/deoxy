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
		label   string
		funcName string
		params  []symbol.Param
		returns []string
		want    []string // substrings that must appear
		notWant []string // substrings that must NOT appear
	}{
		{
			label:    "Add",
			funcName: "Add",
			params:   []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			returns:  []string{"int"},
			want:     []string{"adds", "a", "b"},
		},
		{
			label:    "greet",
			funcName: "greet",
			params:   []symbol.Param{{Name: "name", Type: "str"}, {Name: "age", Type: "int"}},
			returns:  []string{"str"},
			want:     []string{"greets"},
		},
		{
			label:    "DoNothing",
			funcName: "DoNothing",
			params:   nil,
			returns:  nil,
			want:     []string{"doNothings"},
		},
		{
			label:    "empty name",
			funcName: "",
			params:   nil,
			returns:  nil,
			want:     []string{""},
		},
		{
			label:    "single char F",
			funcName: "F",
			params:   []symbol.Param{{Name: "x", Type: "int"}},
			returns:  []string{"int"},
			want:     []string{"fs", "x", "returns"},
		},
		{
			label:    "name ending in s",
			funcName: "Process",
			params:   nil,
			returns:  nil,
			want:     []string{"process"},
		},
		{
			label:    "name ending in ed",
			funcName: "Need",
			params:   nil,
			returns:  nil,
			want:     []string{"need"},
		},
		{
			label:    "params no returns",
			funcName: "Test",
			params:   []symbol.Param{{Name: "a", Type: "int"}},
			returns:  nil,
			want:     []string{"tests", "a"},
			notWant:  []string{"returns"},
		},
		{
			label:    "returns no params",
			funcName: "Test",
			params:   nil,
			returns:  []string{"int"},
			want:     []string{"tests", "returns"},
		},
		{
			label:    "multiple returns",
			funcName: "Test",
			params:   []symbol.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			returns:  []string{"int", "error"},
			want:     []string{"a", "b", "returns"},
		},
	}

	eng := &Engine{}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := eng.brief(symbol.SymbolInfo{
				Name:    tt.funcName,
				Kind:    symbol.KindFunction,
				Params:  tt.params,
				Returns: tt.returns,
			})
			if tt.want[0] == "" && got != "" {
				t.Errorf("brief(%q, ...) = %q, want empty string", tt.funcName, got)
			}
			for _, w := range tt.want {
				if w != "" && !strings.Contains(got, w) {
					t.Errorf("brief(%q, ...) = %q, want it to contain %q", tt.funcName, got, w)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("brief(%q, ...) = %q, should NOT contain %q", tt.funcName, got, nw)
				}
			}
		})
	}
}

func TestParamDescFunction(t *testing.T) {
	eng := &Engine{}

	t.Run("with valid params", func(t *testing.T) {
		params := []symbol.Param{{Name: "a", Type: "int"}}
		got := eng.paramDesc(params, nil, 0)
		if got == "" {
			t.Error("paramDesc should return non-empty string")
		}
	})

	t.Run("with empty params", func(t *testing.T) {
		got := eng.paramDesc(nil, nil, 0)
		if got != "" {
			t.Errorf("paramDesc with empty params should return empty string, got: %q", got)
		}
	})

	t.Run("with descriptive param name", func(t *testing.T) {
		params := []symbol.Param{{Name: "name", Type: "string"}}
		got := eng.paramDesc(params, nil, 0)
		if !strings.Contains(got, "name") {
			t.Errorf("paramDesc should contain param name, got: %q", got)
		}
	})

	t.Run("generic param name uses name fallback", func(t *testing.T) {
		params := []symbol.Param{{Name: "a", Type: "int"}}
		got := eng.paramDesc(params, nil, 0)
		if !strings.Contains(got, "the a") {
			t.Errorf("generic param 'a' should describe by name, got: %q", got)
		}
	})

	t.Run("sixth param uses bare ordinal", func(t *testing.T) {
		params := []symbol.Param{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
			{Name: "c", Type: "int"},
			{Name: "d", Type: "int"},
			{Name: "e", Type: "int"},
			{Name: "f", Type: "int"},
		}
		got := eng.paramDesc(params, nil, 5)
		if !strings.Contains(got, "the") {
			t.Errorf("6th param should start with 'the', got: %q", got)
		}
	})

	t.Run("negative ordinal returns empty", func(t *testing.T) {
		params := []symbol.Param{{Name: "a", Type: "int"}}
		got := eng.paramDesc(params, nil, -1)
		if got != "" {
			t.Errorf("negative ordinal should return empty, got: %q", got)
		}
	})

	t.Run("ordinal beyond length returns empty", func(t *testing.T) {
		params := []symbol.Param{{Name: "a", Type: "int"}}
		got := eng.paramDesc(params, nil, 5)
		if got != "" {
			t.Errorf("out-of-bounds ordinal should return empty, got: %q", got)
		}
	})

	t.Run("empty param name uses ordinal", func(t *testing.T) {
		params := []symbol.Param{{Name: "", Type: "int"}}
		got := eng.paramDesc(params, nil, 0)
		if !strings.Contains(got, "first operand") {
			t.Errorf("empty name should use ordinal, got: %q", got)
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

	t.Run("three returns uses oxford comma", func(t *testing.T) {
		got := returnDesc([]string{"int", "string", "error"})
		if !strings.Contains(got, ", and") {
			t.Errorf("three returns should use oxford comma, got: %q", got)
		}
	})

	t.Run("single return includes type", func(t *testing.T) {
		got := returnDesc([]string{"*Person"})
		if !strings.Contains(got, "person") {
			t.Errorf("single return should include type, got: %q", got)
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
		{nil, ""},
		{[]symbol.Param{}, ""},
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
