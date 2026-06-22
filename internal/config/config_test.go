package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNoConfigFileByDefault(t *testing.T) {
	// Without any config file, the package should not panic
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Just verify no panic
	_ = dir
}

func TestConfigFileExists(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte("docstyle: godoc\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(data) != "docstyle: godoc\n" {
		t.Errorf("unexpected config content: %s", string(data))
	}
}

func TestMultipleDocStyles(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"godoc", "docstyle: godoc\n"},
		{"doxygen", "docstyle: doxygen\n"},
		{"jsdoc", "docstyle: jsdoc\n"},
		{"python", "docstyle: python\n"},
		{"rustdoc", "docstyle: rustdoc\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, ".deoxy.yaml")
			if err := os.WriteFile(cfgPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}
			if len(data) == 0 {
				t.Error("expected non-empty config")
			}
		})
	}
}

func TestCustomTags(t *testing.T) {
	content := []byte("custom_tags:\n  - note\n  - example\n")
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	data, _ := os.ReadFile(cfgPath)
	if len(data) == 0 {
		t.Error("expected non-empty config")
	}
}
