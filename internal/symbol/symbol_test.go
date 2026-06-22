package symbol

import "testing"

func TestKindString(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{KindFunction, "function"},
		{KindMethod, "method"},
		{KindStruct, "struct"},
		{KindClass, "class"},
		{KindInterface, "interface"},
		{KindEnum, "enum"},
		{Kind(-1), "unknown"},
		{Kind(99), "unknown"},
	}
	for _, tt := range tests {
		got := tt.kind.String()
		if got != tt.want {
			t.Errorf("Kind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestParamZeroValue(t *testing.T) {
	var p Param
	if p.Name != "" {
		t.Errorf("expected empty Name, got %q", p.Name)
	}
	if p.Type != "" {
		t.Errorf("expected empty Type, got %q", p.Type)
	}
}

func TestSymbolInfoDefaults(t *testing.T) {
	var s SymbolInfo
	if s.Name != "" {
		t.Errorf("expected empty Name, got %q", s.Name)
	}
	if s.Kind != KindFunction {
		t.Errorf("expected KindFunction (0), got %v", s.Kind)
	}
	if s.Params != nil {
		t.Errorf("expected nil Params, got %v", s.Params)
	}
	if s.Returns != nil {
		t.Errorf("expected nil Returns, got %v", s.Returns)
	}
	if s.Receiver != nil {
		t.Errorf("expected nil Receiver, got %v", s.Receiver)
	}
}

func TestKindValues(t *testing.T) {
	if KindFunction != 0 {
		t.Errorf("KindFunction should be 0, got %d", KindFunction)
	}
	if KindMethod != KindFunction+1 {
		t.Errorf("KindMethod should be KindFunction+1")
	}
	if KindEnum != 5 {
		t.Errorf("KindEnum should be 5, got %d", KindEnum)
	}
}

func TestParamStruct(t *testing.T) {
	p := Param{Name: "count", Type: "int"}
	if p.Name != "count" {
		t.Errorf("expected Name 'count', got %q", p.Name)
	}
	if p.Type != "int" {
		t.Errorf("expected Type 'int', got %q", p.Type)
	}
}

func TestSymbolInfoWithData(t *testing.T) {
	s := SymbolInfo{
		Name: "Add",
		Kind: KindFunction,
		Params: []Param{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		Returns: []string{"int"},
		StartLine: 3,
		EndLine:   5,
	}
	if s.Name != "Add" {
		t.Errorf("expected Name 'Add', got %q", s.Name)
	}
	if len(s.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(s.Params))
	}
	if len(s.Returns) != 1 || s.Returns[0] != "int" {
		t.Errorf("expected Returns [int], got %v", s.Returns)
	}
	if s.StartLine != 3 || s.EndLine != 5 {
		t.Errorf("expected start=3 end=5, got start=%d end=%d", s.StartLine, s.EndLine)
	}
}

func TestMethodSymbol(t *testing.T) {
	s := SymbolInfo{
		Name: "Greet",
		Kind: KindMethod,
		Receiver: &Param{Name: "p", Type: "Person"},
		Params: []Param{
			{Name: "greeting", Type: "string"},
		},
		Returns: []string{"string"},
	}
	if s.Kind != KindMethod {
		t.Errorf("expected KindMethod, got %v", s.Kind)
	}
	if s.Receiver == nil {
		t.Fatal("expected non-nil Receiver")
	}
	if s.Receiver.Name != "p" || s.Receiver.Type != "Person" {
		t.Errorf("expected receiver p Person, got %v", *s.Receiver)
	}
}
