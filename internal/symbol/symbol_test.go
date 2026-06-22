package symbol

import (
	"testing"
)

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
		{Kind(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("Kind(%d).String() = %q, want %q", int(tt.kind), got, tt.want)
			}
		})
	}
}

func TestParamDefaults(t *testing.T) {
	p := Param{}
	if p.Name != "" {
		t.Errorf("default Param.Name should be empty, got %q", p.Name)
	}
	if p.Type != "" {
		t.Errorf("default Param.Type should be empty, got %q", p.Type)
	}
}

func TestSymbolInfoDefaults(t *testing.T) {
	s := SymbolInfo{}
	if s.Name != "" {
		t.Errorf("default Name should be empty, got %q", s.Name)
	}
	if s.Kind != KindFunction {
		t.Errorf("default Kind should be KindFunction(0), got %v", s.Kind)
	}
	if s.Params != nil {
		t.Errorf("default Params should be nil, got %v", s.Params)
	}
	if s.Returns != nil {
		t.Errorf("default Returns should be nil, got %v", s.Returns)
	}
	if s.Receiver != nil {
		t.Errorf("default Receiver should be nil, got %v", s.Receiver)
	}
	if s.StartLine != 0 {
		t.Errorf("default StartLine should be 0, got %d", s.StartLine)
	}
	if s.EndLine != 0 {
		t.Errorf("default EndLine should be 0, got %d", s.EndLine)
	}
}

func TestSymbolInfoConstruction(t *testing.T) {
	s := SymbolInfo{
		Name:    "Add",
		Kind:    KindFunction,
		Params:  []Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
		Returns: []string{"int"},
		TypeParams: []Param{{Name: "T", Type: "any"}},
		Receiver: &Param{Name: "p", Type: "Point"},
		StartLine: 10,
		EndLine:   20,
	}

	if s.Name != "Add" {
		t.Errorf("Name = %q, want %q", s.Name, "Add")
	}
	if s.Kind != KindFunction {
		t.Errorf("Kind = %v, want KindFunction", s.Kind)
	}
	if len(s.Params) != 2 {
		t.Errorf("len(Params) = %d, want 2", len(s.Params))
	}
	if s.Params[0].Name != "a" {
		t.Errorf("Params[0].Name = %q, want %q", s.Params[0].Name, "a")
	}
	if s.Params[1].Type != "int" {
		t.Errorf("Params[1].Type = %q, want %q", s.Params[1].Type, "int")
	}
	if len(s.Returns) != 1 {
		t.Errorf("len(Returns) = %d, want 1", len(s.Returns))
	}
	if s.Returns[0] != "int" {
		t.Errorf("Returns[0] = %q, want %q", s.Returns[0], "int")
	}
	if len(s.TypeParams) != 1 {
		t.Errorf("len(TypeParams) = %d, want 1", len(s.TypeParams))
	}
	if s.TypeParams[0].Name != "T" {
		t.Errorf("TypeParams[0].Name = %q, want %q", s.TypeParams[0].Name, "T")
	}
	if s.Receiver == nil {
		t.Fatal("Receiver should not be nil")
	}
	if s.Receiver.Name != "p" {
		t.Errorf("Receiver.Name = %q, want %q", s.Receiver.Name, "p")
	}
	if s.StartLine != 10 {
		t.Errorf("StartLine = %d, want 10", s.StartLine)
	}
	if s.EndLine != 20 {
		t.Errorf("EndLine = %d, want 20", s.EndLine)
	}
}

func TestParamConstruction(t *testing.T) {
	p := Param{Name: "count", Type: "int"}
	if p.Name != "count" {
		t.Errorf("Name = %q, want %q", p.Name, "count")
	}
	if p.Type != "int" {
		t.Errorf("Type = %q, want %q", p.Type, "int")
	}
}

func TestKindValues(t *testing.T) {
	if KindFunction != 0 {
		t.Errorf("KindFunction should be 0 (iota), got %d", KindFunction)
	}
	if KindMethod != 1 {
		t.Errorf("KindMethod should be 1, got %d", KindMethod)
	}
	if KindEnum != 5 {
		t.Errorf("KindEnum should be 5, got %d", KindEnum)
	}
}

func TestKindStringAll(t *testing.T) {
	// Ensure all iota values produce non-empty strings
	kinds := []Kind{KindFunction, KindMethod, KindStruct, KindClass, KindInterface, KindEnum}
	for _, k := range kinds {
		if k.String() == "" {
			t.Errorf("Kind(%d).String() returned empty string", int(k))
		}
	}
}

func TestSymbolInfoNilReceiver(t *testing.T) {
	s := SymbolInfo{
		Name:     "Foo",
		Kind:     KindFunction,
		Receiver: nil,
	}
	if s.Receiver != nil {
		t.Error("Receiver should be nil for functions")
	}
}

func TestSymbolInfoEmptySlices(t *testing.T) {
	// nil slices vs empty slices
	s1 := SymbolInfo{Name: "Foo"}
	s2 := SymbolInfo{
		Name:    "Bar",
		Params:  []Param{},
		Returns: []string{},
	}

	if s1.Params == nil && s2.Params != nil {
		t.Log("nil vs empty Params — both valid, ensure templates handle both")
	}

	// Both should be treated as "no params" by len()
	if len(s1.Params) != 0 {
		t.Errorf("len(nil Params) = %d, want 0", len(s1.Params))
	}
	if len(s2.Params) != 0 {
		t.Errorf("len(empty Params) = %d, want 0", len(s2.Params))
	}
	if len(s1.Returns) != 0 {
		t.Errorf("len(nil Returns) = %d, want 0", len(s1.Returns))
	}
	if len(s2.Returns) != 0 {
		t.Errorf("len(empty Returns) = %d, want 0", len(s2.Returns))
	}
}
