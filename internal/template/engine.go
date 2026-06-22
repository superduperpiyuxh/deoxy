// Package template provides a Go text/template-based engine for rendering
// doc comments from parsed symbol information. It supports five languages
// (Go, Python, C, C++, Rust) with language-appropriate comment styles
// (GoDoc prose, Google-style Python docstrings, Doxygen for C/C++, Rustdoc).
//
// Thread-safe after construction — Engine is read-only once created.
package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

// TemplateData holds the data context passed to each template during Render.
// Fields are derived from SymbolInfo and computed by the engine's helper functions.
type TemplateData struct {
	Name       string          // Symbol name (function, struct, etc.)
	Kind       string          // Kind string (function, method, struct, etc.)
	Params     []symbol.Param  // Function/method parameters
	Returns    []string        // Return types (empty if no return)
	TypeParams []symbol.Param  // Generic/type parameters
	Receiver   *symbol.Param   // Method receiver (nil for functions)
	HasReceiver bool           // True if Receiver is non-nil
	Brief      string          // Pre-computed brief description
	Lang       string          // Language identifier passed to Render
}

// Engine is a template engine that renders SymbolInfo into language-specific
// doc comments using Go text/template. Thread-safe after construction.
type Engine struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
}

// New creates an Engine from a map of raw template strings keyed by language
// name (e.g., "go", "python", "c", "cpp", "rust").
//
// Each template string is compiled with the engine's shared FuncMap.
// Returns an error if any template fails to parse.
func New(tpls map[string]string) (*Engine, error) {
	e := &Engine{
		templates: make(map[string]*template.Template, len(tpls)),
		funcs:     sharedFuncs(),
	}

	for lang, tplStr := range tpls {
		tmpl, err := template.New(lang).Funcs(e.funcs).Parse(tplStr)
		if err != nil {
			return nil, fmt.Errorf("template: failed to parse %q template: %w", lang, err)
		}
		e.templates[lang] = tmpl
	}

	// No nil map check — if tpls is nil, we get an empty templates map,
	// which is valid: Render will return "unknown language" for any lang.

	return e, nil
}

// Render executes the template for the given language against SymbolInfo data
// and returns the formatted doc comment string.
//
// Returns an error if:
//   - the language is unknown (not in the templates map)
//   - template execution fails
func (e *Engine) Render(info symbol.SymbolInfo, lang string) (string, error) {
	tmpl, ok := e.templates[lang]
	if !ok {
		return "", fmt.Errorf("unknown language: %s", lang)
	}

	if info.Name == "" {
		return "", nil
	}

	data := TemplateData{
		Name:        info.Name,
		Kind:        info.Kind.String(),
		Params:      info.Params,
		Returns:     info.Returns,
		TypeParams:  info.TypeParams,
		Receiver:    info.Receiver,
		HasReceiver: info.Receiver != nil,
		Lang:        lang,
	}

	// Pre-compute brief
	data.Brief = brief(info.Name, info.Params, info.Returns)

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template: execution failed for %q rendering %q: %w", lang, info.Name, err)
	}

	return buf.String(), nil
}

// sharedFuncs returns the FuncMap shared by all templates.
func sharedFuncs() template.FuncMap {
	return template.FuncMap{
		"brief":         brief,
		"paramDesc":     paramDesc,
		"returnDesc":    returnDesc,
		"joinParams":    joinParams,
		"commentPrefix": commentPrefix,
		"hasParams":     hasParams,
		"hasReturns":    hasReturns,
		"hasTypeParams": hasTypeParams,
	}
}

// brief generates a brief description sentence from the symbol name and
// signature. For functions: lowercases first letter of name, appends
// parameter names with 'and' joining, and appends return phrase.
//
// Examples:
//
//	brief("Add", [{a,int},{b,int}], ["int"]) → "adds a and b and returns the result."
//	brief("greet", [{name,str},{age,int}], ["str"]) → "greets a person with the given name and age."
//	brief("DoNothing", [], []) → "does nothing."
func brief(name string, params []symbol.Param, returns []string) string {
	if name == "" {
		return ""
	}

	verb := lowerFirst(name)
	if !strings.HasSuffix(verb, "s") && !strings.HasSuffix(verb, "ed") {
		verb = verb + "s"
	}

	parts := []string{verb}

	if len(params) > 0 {
		parts = append(parts, joinParams(params))
	}

	if len(returns) > 0 {
		parts = append(parts, "and returns the result.")
	} else if len(params) > 0 {
		// Add closing period for sentences with params but no returns
	}

	if !strings.HasSuffix(parts[len(parts)-1], ".") {
		parts[len(parts)-1] = parts[len(parts)-1] + "."
	}

	return strings.Join(parts, " ")
}

// paramDesc generates a description for a parameter at the given ordinal
// position. Uses the param name as a hint if it looks descriptive.
//
// Examples:
//
//	paramDesc([{a,int}], 0) → "the first operand"
//	paramDesc([{name,str}], 0) → "the name"
func paramDesc(params []symbol.Param, ordinal int) string {
	if ordinal < 0 || ordinal >= len(params) {
		return ""
	}

	name := params[ordinal].Name

	// Use the parameter name itself as a description hint
	if name != "" && !isGenericParamName(name) {
		return "the " + name
	}

	// Fall back to ordinal-based description
	ordinals := []string{"first", "second", "third", "fourth", "fifth"}
	if ordinal < len(ordinals) {
		return "the " + ordinals[ordinal] + " operand"
	}
	return fmt.Sprintf("the %dth parameter", ordinal+1)
}

// isGenericParamName returns true for common generic/short parameter names
// that don't make good descriptions.
func isGenericParamName(name string) bool {
	switch name {
	case "a", "b", "c", "x", "y", "z", "i", "j", "k", "n", "m", "p", "q", "r", "s", "t", "v", "w":
		return true
	}
	return false
}

// returnDesc generates a return description based on return types.
//
// Examples:
//
//	returnDesc(["int"]) → "the result"
//	returnDesc(["int", "error"]) → "the result and an error"
func returnDesc(returns []string) string {
	if len(returns) == 0 {
		return ""
	}

	parts := make([]string, len(returns))
	for i, r := range returns {
		parts[i] = "the " + r + " result"
	}

	if len(parts) == 1 {
		return parts[0]
	}

	// Join with ", " and final "and"
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}

// joinParams joins parameter names with ", " and " and " for use in brief sentences.
//
// Examples:
//
//	joinParams([{a,int},{b,int}]) → "a and b"
//	joinParams([{x,int},{y,int},{z,int}]) → "x, y, and z"
func joinParams(params []symbol.Param) string {
	if len(params) == 0 {
		return ""
	}

	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}

	if len(names) == 1 {
		return names[0]
	}

	if len(names) == 2 {
		return names[0] + " and " + names[1]
	}

	return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
}

// commentPrefix returns the comment prefix for the given language.
//
//	"go" → "//"
//	"python" → "#"
//	"rust" → "///"
//	"c", "cpp" → " *"  (inside /** */ block)
func commentPrefix(lang string) string {
	switch lang {
	case "go":
		return "//"
	case "python":
		return "#"
	case "rust":
		return "///"
	case "c", "cpp":
		return " *"
	default:
		return "//"
	}
}

// hasParams returns true if the parameter slice is non-empty.
func hasParams(params []symbol.Param) bool {
	return len(params) > 0
}

// hasReturns returns true if the returns slice is non-empty.
func hasReturns(returns []string) bool {
	return len(returns) > 0
}

// hasTypeParams returns true if the type parameter slice is non-empty.
func hasTypeParams(tparams []symbol.Param) bool {
	return len(tparams) > 0
}

// lowerFirst lowercases the first character of a string.
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
