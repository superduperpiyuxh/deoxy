package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	content := `version: '1'
default_style: godoc
default_tag_order: brief-first
languages:
  go:
    docstyle: godoc
  c:
    docstyle: doxygen
  python:
    docstyle: pydoc
  rust:
    docstyle: rustdoc
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig should return non-nil Config")
	}
	if cfg.DefaultDocStyle != "godoc" {
		t.Errorf("DefaultDocStyle = %q, want %q", cfg.DefaultDocStyle, "godoc")
	}
	if cfg.DefaultTagOrder != "brief-first" {
		t.Errorf("DefaultTagOrder = %q, want %q", cfg.DefaultTagOrder, "brief-first")
	}
	if cfg.Languages == nil {
		t.Fatal("Languages should not be nil")
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	cfg := LoadDefaultConfig()

	if cfg == nil {
		t.Fatal("LoadDefaultConfig should return non-nil Config")
	}
	if cfg.DefaultDocStyle != "godoc" {
		t.Errorf("DefaultDocStyle = %q, want %q", cfg.DefaultDocStyle, "godoc")
	}
	if cfg.DefaultTagOrder != "brief-first" {
		t.Errorf("DefaultTagOrder = %q, want %q", cfg.DefaultTagOrder, "brief-first")
	}
	if cfg.Languages == nil {
		t.Fatal("Languages should not be nil in default config")
	}
	if len(cfg.Languages) != 5 {
		t.Errorf("Default config should have 5 languages, got %d", len(cfg.Languages))
	}

	// Verify specific language defaults
	for lang, expectedDocStyle := range map[string]string{
		"go":     "godoc",
		"python": "pydoc",
		"c":      "doxygen",
		"cpp":    "doxygen",
		"rust":   "rustdoc",
	} {
		if langCfg, ok := cfg.Languages[lang]; ok {
			if langCfg.DocStyle != expectedDocStyle {
				t.Errorf("Language %q docstyle = %q, want %q", lang, langCfg.DocStyle, expectedDocStyle)
			}
		} else {
			t.Errorf("Language %q not found in default config", lang)
		}
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/.deoxy.yaml")
	if err == nil {
		t.Fatal("LoadConfig on non-existent file should return error")
	}
	if !strings.Contains(err.Error(), "cannot read") && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("error message should mention file not found, got: %v", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte("invalid: [yaml: broken"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("LoadConfig on invalid YAML should return error")
	}
}

func TestGetLanguageConfigMerge(t *testing.T) {
	cfg := &Config{
		DefaultDocStyle: "godoc",
		DefaultTagOrder: "brief-first",
		Languages: map[string]LanguageConfig{
			"c": {DocStyle: "doxygen", TagOrder: "brief-first"},
		},
	}

	// Go should get default docstyle
	goCfg := cfg.GetLanguageConfig("go")
	if goCfg.DocStyle != "godoc" {
		t.Errorf("Go DocStyle = %q, want %q", goCfg.DocStyle, "godoc")
	}

	// C should get doxygen from override
	cCfg := cfg.GetLanguageConfig("c")
	if cCfg.DocStyle != "doxygen" {
		t.Errorf("C DocStyle = %q, want %q", cCfg.DocStyle, "doxygen")
	}

	// Go should not inherit C's override
	if goCfg.DocStyle == "doxygen" {
		t.Error("Go should not inherit C's docstyle override")
	}
}

func TestGetDocStyleParsing(t *testing.T) {
	tests := []struct {
		style   string
		want    DocStyle
		wantErr bool
	}{
		{"godoc", DocStyleGoDoc, false},
		{"doxygen", DocStyleDoxygen, false},
		{"pydoc", DocStylePyDoc, false},
		{"rustdoc", DocStyleRustdoc, false},
		{"unknown", DocStyleGoDoc, true},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			got, err := ParseDocStyle(tt.style)
			if tt.wantErr && err == nil {
				t.Error("expected error for invalid docstyle")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseDocStyle(%q) = %v, want %v", tt.style, got, tt.want)
			}
		})
	}
}

func TestGetDocStyleFromConfig(t *testing.T) {
	cfg := LoadDefaultConfig()

	docStyles := map[string]DocStyle{
		"go":     DocStyleGoDoc,
		"python": DocStylePyDoc,
		"c":      DocStyleDoxygen,
		"cpp":    DocStyleDoxygen,
		"rust":   DocStyleRustdoc,
	}

	for lang, want := range docStyles {
		t.Run(lang, func(t *testing.T) {
			got := cfg.GetDocStyle(lang)
			if got != want {
				t.Errorf("GetDocStyle(%q) = %v, want %v", lang, got, want)
			}
		})
	}
}

func TestUnknownKeyTolerance(t *testing.T) {
	// Write a config with unknown top-level key
	content := `version: '1'
default_style: godoc
experimental_feature: true
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig with unknown key should not error, got: %v", err)
	}
	if cfg.DefaultDocStyle != "godoc" {
		t.Errorf("DefaultDocStyle = %q, want %q", cfg.DefaultDocStyle, "godoc")
	}
	if cfg.Version != "1" {
		t.Errorf("Version = %q, want %q", cfg.Version, "1")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	original := &Config{
		DefaultDocStyle: "godoc",
		DefaultTagOrder: "brief-first",
		Languages: map[string]LanguageConfig{
			"go": {DocStyle: "godoc", TagOrder: "brief-first"},
			"c":  {DocStyle: "doxygen", TagOrder: "params-first"},
		},
	}

	data, err := original.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("MarshalYAML returned empty data")
	}

	// Write and re-load
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig after round-trip failed: %v", err)
	}
	if loaded.DefaultDocStyle != original.DefaultDocStyle {
		t.Errorf("round-trip: DefaultDocStyle = %q, want %q", loaded.DefaultDocStyle, original.DefaultDocStyle)
	}
	if loaded.DefaultTagOrder != original.DefaultTagOrder {
		t.Errorf("round-trip: DefaultTagOrder = %q, want %q", loaded.DefaultTagOrder, original.DefaultTagOrder)
	}
}

func TestEmptyConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error for empty config, got: %v", err)
	}
	if cfg.DefaultDocStyle != "godoc" {
		t.Errorf("DefaultDocStyle = %q, want %q", cfg.DefaultDocStyle, "godoc")
	}
	if cfg.DefaultTagOrder != "brief-first" {
		t.Errorf("DefaultTagOrder = %q, want %q", cfg.DefaultTagOrder, "brief-first")
	}
}

func TestPerLanguageTemplateOverrides(t *testing.T) {
	content := `version: '1'
languages:
  go:
    brief_template: 'Custom: {{.Name}}'
    param_template: '- {{.Name}}'
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error, got: %v", err)
	}

	goCfg := cfg.GetLanguageConfig("go")
	if goCfg.BriefTemplate != "Custom: {{.Name}}" {
		t.Errorf("BriefTemplate = %q, want %q", goCfg.BriefTemplate, "Custom: {{.Name}}")
	}
	if goCfg.ParamTemplate != "- {{.Name}}" {
		t.Errorf("ParamTemplate = %q, want %q", goCfg.ParamTemplate, "- {{.Name}}")
	}

	// Other languages should have empty template overrides
	cCfg := cfg.GetLanguageConfig("c")
	if cCfg.BriefTemplate != "" {
		t.Errorf("C BriefTemplate should be empty, got %q", cCfg.BriefTemplate)
	}
}

func TestCustomTags(t *testing.T) {
	content := `version: '1'
languages:
  cpp:
    custom_tags:
      - note
      - warning
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error, got: %v", err)
	}

	cppCfg := cfg.GetLanguageConfig("cpp")
	if len(cppCfg.CustomTags) != 2 {
		t.Fatalf("C++ CustomTags length = %d, want 2", len(cppCfg.CustomTags))
	}
	if cppCfg.CustomTags[0] != "note" {
		t.Errorf("CustomTags[0] = %q, want %q", cppCfg.CustomTags[0], "note")
	}
	if cppCfg.CustomTags[1] != "warning" {
		t.Errorf("CustomTags[1] = %q, want %q", cppCfg.CustomTags[1], "warning")
	}

	// Other languages should have nil/empty CustomTags
	goCfg := cfg.GetLanguageConfig("go")
	if goCfg.CustomTags != nil {
		t.Errorf("Go CustomTags should be nil, got %v", goCfg.CustomTags)
	}
}
