package smarttext

import (
	"testing"
)

func TestIsGetter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"standard getter", "GetName", true},
		{"getter with multi-word", "GetMaxRetryCount", true},
		{"not getter (get lowercase)", "getName", false},
		{"not getter (short)", "Get", false},
		{"not getter (no prefix)", "getName", false},
		{"getter single word after", "GetX", true},
		{"getter with acronym", "GetHTTPSPort", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGetter(tt.input); got != tt.want {
				t.Errorf("IsGetter(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSetter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"standard setter", "SetName", true},
		{"setter with multi-word", "SetMaxRetryCount", true},
		{"not setter (lowercase)", "setName", false},
		{"not setter (short)", "Set", false},
		{"setter with underscore", "Set_thing", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSetter(tt.input); got != tt.want {
				t.Errorf("IsSetter(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsConstructor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"standard constructor", "NewConfig", true},
		{"constructor multi-word", "NewHTTPServer", true},
		{"not constructor (lowercase)", "newConfig", false},
		{"not constructor (short)", "New", false},
		{"not constructor", "NewlyDiscovered", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConstructor(tt.input); got != tt.want {
				t.Errorf("IsConstructor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple camelCase", "maxRetryCount", []string{"max", "retry", "count"}},
		{"PascalCase", "GetName", []string{"get", "name"}},
		{"single word", "Config", []string{"config"}},
		{"empty", "", nil},
		{"snake_case", "parse_json", []string{"parse", "json"}},
		{"single snake", "value", []string{"value"}},
		{"acronym at end", "ParseXML", []string{"parse", "xml"}},
		{"acronym at start", "XMLParser", []string{"xml", "parser"}},
		{"all caps", "ID", []string{"id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitCamelCase(tt.input)
			if len(got) != len(tt.want) {
			t.Errorf("SplitCamelCase(%q) = %v (len=%d), want %v (len=%d)", tt.input, got, len(got), tt.want, len(tt.want))
			return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitCamelCase(%q) = %v, want %v", tt.input, got, tt.want)
					break
				}
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		name string
		fn   string
		want string
	}{
		{"constructor", "NewConfig", "NewConfig creates a new config"},
		{"getter", "GetName", "Gets the name"},
		{"setter", "SetMaxRetryCount", "Sets the max retry count"},
		{"not special", "DoSomething", ""},
		{"constructor multi-word", "NewHTTPServer", "NewHTTPServer creates a new http server"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Describe(tt.fn, nil, nil, reg)
			if got != tt.want {
				t.Errorf("Describe(%q) = %q, want %q", tt.fn, got, tt.want)
			}
		})
	}
}
