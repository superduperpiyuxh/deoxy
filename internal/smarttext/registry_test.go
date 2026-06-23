package smarttext

import (
	"testing"
)

func TestRegistryDefaultEntries(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"count", "count", "Number of items"},
		{"name", "name", "The name"},
		{"err", "err", "An error"},
		{"ctx", "ctx", "The context"},
		{"nonexistent", "nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Get(tt.key); got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestRegistrySet(t *testing.T) {
	r := NewRegistry()
	r.Set("custom", "A custom value")
	if got := r.Get("custom"); got != "A custom value" {
		t.Errorf("Get after Set = %q, want %q", got, "A custom value")
	}
}

func TestRegistryOverride(t *testing.T) {
	overrides := map[string]string{"count": "Custom count description"}
	r := NewRegistryWithOverrides(overrides)
	if got := r.Get("count"); got != "Custom count description" {
		t.Errorf("Get(overridden count) = %q, want %q", got, "Custom count description")
	}
	if got := r.Get("name"); got != "The name" {
		t.Errorf("Get(default name) = %q, want %q", got, "The name")
	}
}
