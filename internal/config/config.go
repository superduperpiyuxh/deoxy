// Package config provides configuration loading and management for deoxy.
//
// It handles reading .deoxy.yaml files, per-language settings, docstyle
// selection (GoDoc, Doxygen, Python docstrings, Rustdoc), tag ordering,
// and user-defined custom tags.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DocStyle represents the documentation comment style for a language.
type DocStyle int

const (
	// DocStyleGoDoc is prose-style GoDoc comments (// FuncName does X).
	DocStyleGoDoc DocStyle = iota
	// DocStyleDoxygen is Doxygen/JSDoc-style (/** @brief ... */).
	DocStyleDoxygen
	// DocStylePyDoc is Google-style Python docstrings ("""... Args: ... Returns: ...""").
	DocStylePyDoc
	// DocStyleRustdoc is Rust-style /// doc comments.
	DocStyleRustdoc
)

// String returns the string representation of a DocStyle.
func (d DocStyle) String() string {
	switch d {
	case DocStyleGoDoc:
		return "godoc"
	case DocStyleDoxygen:
		return "doxygen"
	case DocStylePyDoc:
		return "pydoc"
	case DocStyleRustdoc:
		return "rustdoc"
	default:
		return "unknown"
	}
}

// ParseDocStyle parses a docstyle string into the DocStyle enum.
// Returns an error if the string is not a valid docstyle.
func ParseDocStyle(s string) (DocStyle, error) {
	switch s {
	case "godoc":
		return DocStyleGoDoc, nil
	case "doxygen":
		return DocStyleDoxygen, nil
	case "pydoc":
		return DocStylePyDoc, nil
	case "rustdoc":
		return DocStyleRustdoc, nil
	default:
		return DocStyleGoDoc, fmt.Errorf("config: unknown docstyle %q", s)
	}
}

// TagOrder controls the order of sections in generated doc comments.
type TagOrder int

const (
	// TagOrderBriefFirst places brief description before parameter/return tags (default).
	TagOrderBriefFirst TagOrder = iota
	// TagOrderParamsFirst places parameter tags before brief description.
	TagOrderParamsFirst
)

// String returns the string representation of a TagOrder.
func (t TagOrder) String() string {
	switch t {
	case TagOrderBriefFirst:
		return "brief-first"
	case TagOrderParamsFirst:
		return "params-first"
	default:
		return "unknown"
	}
}

// ParseTagOrder parses a tag order string into the TagOrder enum.
func ParseTagOrder(s string) (TagOrder, error) {
	switch s {
	case "brief-first":
		return TagOrderBriefFirst, nil
	case "params-first":
		return TagOrderParamsFirst, nil
	default:
		return TagOrderBriefFirst, fmt.Errorf("config: unknown tag order %q", s)
	}
}

// LanguageConfig holds per-language configuration overrides.
type LanguageConfig struct {
	DocStyle       string   `yaml:"docstyle,omitempty"`
	TagOrder       string   `yaml:"tag_order,omitempty"`
	BriefTemplate  string   `yaml:"brief_template,omitempty"`
	ParamTemplate  string   `yaml:"param_template,omitempty"`
	ReturnTemplate string   `yaml:"return_template,omitempty"`
	CustomTags     []string `yaml:"custom_tags,omitempty"`
}

// Config is the top-level deoxy configuration.
type Config struct {
	Version         string                    `yaml:"version,omitempty"`
	Languages       map[string]LanguageConfig `yaml:"languages,omitempty"`
	DefaultDocStyle string                    `yaml:"default_style,omitempty"`
	DefaultTagOrder string                    `yaml:"default_tag_order,omitempty"`
}

// LoadDefaultConfig returns a Config with all default values.
// This represents the configuration used when no .deoxy.yaml exists.
func LoadDefaultConfig() *Config {
	return &Config{
		DefaultDocStyle: "godoc",
		DefaultTagOrder: "brief-first",
		Languages: map[string]LanguageConfig{
			"go":     {DocStyle: "godoc", TagOrder: "brief-first"},
			"python": {DocStyle: "pydoc", TagOrder: "brief-first"},
			"c":      {DocStyle: "doxygen", TagOrder: "brief-first"},
			"cpp":    {DocStyle: "doxygen", TagOrder: "brief-first"},
			"rust":   {DocStyle: "rustdoc", TagOrder: "brief-first"},
		},
	}
}

// LoadConfig loads and parses a .deoxy.yaml file from the given path.
// Returns the parsed Config and nil error on success.
// Returns an error if the file does not exist or the YAML is invalid.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read %q: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: failed to parse %q: %w", path, err)
	}

	// Apply defaults for empty values
	if cfg.DefaultDocStyle == "" {
		cfg.DefaultDocStyle = "godoc"
	}
	if cfg.DefaultTagOrder == "" {
		cfg.DefaultTagOrder = "brief-first"
	}
	if cfg.Languages == nil {
		cfg.Languages = make(map[string]LanguageConfig)
	}

	return cfg, nil
}

// GetLanguageConfig returns the effective configuration for a language,
// merging per-language overrides onto the defaults. Returns the default
// LanguageConfig if no per-language override exists.
func (c *Config) GetLanguageConfig(lang string) LanguageConfig {
	// Start with language-specific defaults based on docstyle
	var defaults LanguageConfig

	// Apply default docstyle per language
	switch lang {
	case "go":
		defaults = LanguageConfig{DocStyle: "godoc", TagOrder: "brief-first"}
	case "python":
		defaults = LanguageConfig{DocStyle: "pydoc", TagOrder: "brief-first"}
	case "c", "cpp":
		defaults = LanguageConfig{DocStyle: "doxygen", TagOrder: "brief-first"}
	case "rust":
		defaults = LanguageConfig{DocStyle: "rustdoc", TagOrder: "brief-first"}
	default:
		defaults = LanguageConfig{DocStyle: c.DefaultDocStyle, TagOrder: c.DefaultTagOrder}
	}

	// Override with per-language config if present
	if c.Languages != nil {
		if langCfg, ok := c.Languages[lang]; ok {
			if langCfg.DocStyle != "" {
				defaults.DocStyle = langCfg.DocStyle
			}
			if langCfg.TagOrder != "" {
				defaults.TagOrder = langCfg.TagOrder
			}
			if langCfg.BriefTemplate != "" {
				defaults.BriefTemplate = langCfg.BriefTemplate
			}
			if langCfg.ParamTemplate != "" {
				defaults.ParamTemplate = langCfg.ParamTemplate
			}
			if langCfg.ReturnTemplate != "" {
				defaults.ReturnTemplate = langCfg.ReturnTemplate
			}
			if langCfg.CustomTags != nil {
				defaults.CustomTags = langCfg.CustomTags
			}
		}
	}

	return defaults
}

// GetDocStyle returns the parsed DocStyle enum for a language,
// considering defaults and per-language overrides.
func (c *Config) GetDocStyle(lang string) DocStyle {
	langCfg := c.GetLanguageConfig(lang)
	dstyle := langCfg.DocStyle

	// Fall back to default docstyle if per-language is not set
	if dstyle == "" {
		dstyle = c.DefaultDocStyle
	}

	docStyle, _ := ParseDocStyle(dstyle)
	return docStyle
}

// MarshalYAML serializes the config to YAML bytes.
func (c *Config) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(c)
}
