package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/superduperpiyuxh/deoxy/internal/smarttext"
	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

type AlignedParam struct {
	Name string
	Type string
	Desc string
}

type TemplateData struct {
	Name        string
	Kind        string
	Params      []symbol.Param
	Returns     []string
	TypeParams  []symbol.Param
	Receiver    *symbol.Param
	HasReceiver bool
	Brief       string
	Lang        string
	Aligned     []AlignedParam
	CustomTags  []string
}

type Engine struct {
	templates        map[string]*template.Template
	funcs            template.FuncMap
	smartTextEnabled bool
	smartReg         *smarttext.Registry
}

type Option func(*Engine)

func WithSmartText(enabled bool, descriptions map[string]string) Option {
	return func(e *Engine) {
		e.smartTextEnabled = enabled
		if enabled {
			e.smartReg = smarttext.NewRegistryWithOverrides(descriptions)
		}
	}
}

func New(tpls map[string]string, opts ...Option) (*Engine, error) {
	e := &Engine{
		templates: make(map[string]*template.Template, len(tpls)),
	}
	for _, opt := range opts {
		opt(e)
	}
	e.funcs = e.sharedFuncs()

	for lang, tplStr := range tpls {
		tmpl, err := template.New(lang).Funcs(e.funcs).Parse(tplStr)
		if err != nil {
			return nil, fmt.Errorf("template: failed to parse %q template: %w", lang, err)
		}
		e.templates[lang] = tmpl
	}

	return e, nil
}

func (e *Engine) Render(info symbol.SymbolInfo, lang string, customTags ...string) (string, error) {
	tmpl, ok := e.templates[lang]
	if !ok {
		return "", fmt.Errorf("unknown language: %s", lang)
	}

	if info.Name == "" {
		return "", nil
	}

	aligned := makeAligned(info.Params, func(i int) string {
		return e.paramDesc(info.Params, i)
	})

	data := TemplateData{
		Name:       info.Name,
		Kind:       info.Kind.String(),
		Params:     info.Params,
		Returns:    info.Returns,
		TypeParams: info.TypeParams,
		Receiver:   info.Receiver,
		HasReceiver: info.Receiver != nil,
		Lang:       lang,
		Aligned:    aligned,
		CustomTags: customTags,
	}

	data.Brief = e.brief(info.Name, info.Params, info.Returns)

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template: execution failed for %q rendering %q: %w", lang, info.Name, err)
	}

	return buf.String(), nil
}

func (e *Engine) sharedFuncs() template.FuncMap {
	return template.FuncMap{
		"brief":         e.brief,
		"paramDesc":     e.paramDesc,
		"returnDesc":    returnDesc,
		"joinParams":    joinParams,
		"commentPrefix": commentPrefix,
		"hasParams":     hasParams,
		"hasReturns":    hasReturns,
		"hasTypeParams": hasTypeParams,
	}
}

func makeAligned(params []symbol.Param, descs func(int) string) []AlignedParam {
	if len(params) == 0 {
		return nil
	}
	maxName := 0
	for _, p := range params {
		if len(p.Name) > maxName {
			maxName = len(p.Name)
		}
	}
	result := make([]AlignedParam, len(params))
	for i, p := range params {
		result[i] = AlignedParam{
			Name: p.Name + strings.Repeat(" ", maxName-len(p.Name)),
			Type: p.Type,
			Desc: descs(i),
		}
	}
	return result
}

func (e *Engine) brief(name string, params []symbol.Param, returns []string) string {
	if name == "" {
		return ""
	}

	if e.smartTextEnabled {
		desc := smarttext.Describe(name, params, returns, e.smartReg)
		if desc != "" {
			return desc
		}
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
		parts = append(parts, "and returns the result")
	}

	last := parts[len(parts)-1]
	if !strings.HasSuffix(last, ".") {
		last += "."
	}
	parts[len(parts)-1] = last

	return strings.Join(parts, " ")
}

func (e *Engine) paramDesc(params []symbol.Param, ordinal int) string {
	if ordinal < 0 || ordinal >= len(params) {
		return ""
	}

	name := params[ordinal].Name

	if e.smartTextEnabled && name != "" {
		if desc := e.smartReg.Get(name); desc != "" {
			return desc
		}
	}

	if name != "" && !isGenericParamName(name) {
		return "the " + name
	}

	ordinals := []string{"first", "second", "third", "fourth", "fifth"}
	if ordinal < len(ordinals) {
		return "the " + ordinals[ordinal] + " operand"
	}
	return fmt.Sprintf("the %dth parameter", ordinal+1)
}

func isGenericParamName(name string) bool {
	switch name {
	case "a", "b", "c", "x", "y", "z", "i", "j", "k", "n", "m", "p", "q", "r", "s", "t", "v", "w":
		return true
	}
	return false
}

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
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}

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

func hasParams(params []symbol.Param) bool {
	return len(params) > 0
}

func hasReturns(returns []string) bool {
	return len(returns) > 0
}

func hasTypeParams(tparams []symbol.Param) bool {
	return len(tparams) > 0
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
