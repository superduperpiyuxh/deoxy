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
	Name       string
	Kind       string
	Params     []symbol.Param
	Returns    []string
	TypeParams []symbol.Param
	Receiver   *symbol.Param
	HasReceiver bool
	Brief      string
	Lang       string
	Aligned    []AlignedParam
	CustomTags []string
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
		return e.paramDesc(info.Params, info.Returns, i)
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

	data.Brief = e.brief(info)

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template: execution failed for %q rendering %q: %w", lang, info.Name, err)
	}
	return buf.String(), nil
}

func (e *Engine) sharedFuncs() template.FuncMap {
	return template.FuncMap{
		"paramDesc":     func(params []symbol.Param, ordinal int) string { return e.paramDesc(params, nil, ordinal) },
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

func (e *Engine) brief(info symbol.SymbolInfo) string {
	name := info.Name
	if name == "" {
		return ""
	}

	switch info.Kind {
	case symbol.KindStruct:
		if e.smartTextEnabled {
			if d := describeStruct(name, info.Params); d != "" {
				return d
			}
		}
		return lowerFirst(name) + " represents a " + lowerFirst(name) + "."

	case symbol.KindInterface:
		return lowerFirst(name) + " describes the contract for a " + lowerFirst(name) + "."

	case symbol.KindEnum:
		return lowerFirst(name) + " represents a " + lowerFirst(name) + " value."
	}

	if e.smartTextEnabled {
		desc := smarttext.Describe(name, info.Params, info.Returns, e.smartReg)
		if desc != "" {
			return desc
		}
	}

	if info.Receiver != nil {
		recvType := receiverBaseType(info.Receiver.Type)
		if info.Kind == symbol.KindMethod && recvType != "" {
			if len(info.Returns) > 0 {
				return lowerFirst(name) + " returns a result for the " + lowerFirst(recvType) + "."
			}
			if len(info.Params) == 0 {
				return lowerFirst(name) + " performs an operation on the " + lowerFirst(recvType) + "."
			}
			return lowerFirst(name) + " performs an operation on the " + lowerFirst(recvType) +
				" with the given " + joinParamNames(info.Params) + "."
		}
	}

	verb := lowerFirst(name)
	if !strings.HasSuffix(verb, "s") && !strings.HasSuffix(verb, "ed") {
		verb = verb + "s"
	}

	parts := []string{verb}
	if len(info.Params) > 0 {
		parts = append(parts, joinParams(info.Params))
	}
	if len(info.Returns) > 0 {
		parts = append(parts, "and returns the result")
	}

	last := parts[len(parts)-1]
	if !strings.HasSuffix(last, ".") {
		last += "."
	}
	parts[len(parts)-1] = last

	return strings.Join(parts, " ")
}

func (e *Engine) paramDesc(params []symbol.Param, returns []string, ordinal int) string {
	if ordinal < 0 || ordinal >= len(params) {
		return ""
	}

	name := params[ordinal].Name
	ptype := params[ordinal].Type

	if e.smartTextEnabled && name != "" {
		if desc := e.smartReg.Get(name); desc != "" {
			return desc
		}
	}

	if name != "" && !isGenericParamName(name) {
		if ptype != "" {
			return describeTypedParam(name, ptype)
		}
		return "the " + name
	}

	if name != "" {
		return "the " + name
	}

	ordinals := []string{"first", "second", "third", "fourth", "fifth"}
	if ordinal < len(ordinals) {
		return "the " + ordinals[ordinal] + " operand"
	}
	return fmt.Sprintf("the %dth parameter", ordinal+1)
}

func describeTypedParam(name, ptype string) string {
	switch {
	case strings.HasPrefix(ptype, "*"):
		return "a pointer to the " + name
	case strings.HasPrefix(ptype, "[]"):
		return "a slice of " + name + " values"
	case ptype == "error" || ptype == "Error":
		return "an error indicating " + name
	case ptype == "bool" || ptype == "Bool" || ptype == "boolean":
		return "whether to " + name
	case ptype == "string" || ptype == "String":
		return "the " + name + " string"
	case ptype == "int" || ptype == "int64" || ptype == "int32":
		return "the " + name + " value"
	case ptype == "float64" || ptype == "float32":
		return "the " + name + " value"
	case ptype == "context.Context" || ptype == "Context" || ptype == "ctx.Context":
		return "the " + name
	case ptype == "time.Duration" || ptype == "Duration":
		return "the " + name + " duration"
	case ptype == "[]byte":
		return "the " + name + " data"
	default:
		return "the " + name
	}
}

func receiverBaseType(recvType string) string {
	recvType = strings.TrimLeft(recvType, "*&")
	if idx := strings.LastIndex(recvType, "."); idx >= 0 {
		return recvType[idx+1:]
	}
	return recvType
}

func describeStruct(name string, fields []symbol.Param) string {
	if len(fields) == 0 {
		return lowerFirst(name) + " represents a " + lowerFirst(name) + "."
	}
	var fieldNames []string
	for _, f := range fields {
		if f.Name != "" {
			fieldNames = append(fieldNames, f.Name)
		}
	}
	if len(fieldNames) == 0 {
		return lowerFirst(name) + " represents a " + lowerFirst(name) + "."
	}
	if len(fieldNames) == 1 {
		return lowerFirst(name) + " holds a " + fieldNames[0] + "."
	}
	last := fieldNames[len(fieldNames)-1]
	joined := strings.Join(fieldNames[:len(fieldNames)-1], ", ") + " and " + last
	return lowerFirst(name) + " holds the " + joined + "."
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
		parts[i] = typeReturnDesc(r, i)
	}

	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}

func typeReturnDesc(rtype string, idx int) string {
	switch {
	case rtype == "error" || rtype == "Error":
		return "an error, if any"
	case rtype == "bool" || rtype == "Bool" || rtype == "boolean":
		return "true if the operation succeeded, false otherwise"
	case rtype == "string" || rtype == "String":
		return "the resulting string"
	case rtype == "int" || rtype == "int64" || rtype == "int32" || rtype == "uint" || rtype == "uint64":
		return "the computed result"
	case rtype == "float64" || rtype == "float32":
		return "the computed floating-point result"
	case strings.HasPrefix(rtype, "*"):
		base := strings.TrimPrefix(rtype, "*")
		return "a pointer to the resulting " + lowerFirst(base)
	case strings.HasPrefix(rtype, "[]"):
		elem := strings.TrimPrefix(rtype, "[]")
		return "a slice of " + lowerFirst(elem) + " values"
	case strings.HasPrefix(rtype, "map["):
		return "a map of results"
	case strings.HasPrefix(rtype, "("):
		return "the result values"
	default:
		return "the " + rtype + " result"
	}
}

func joinParams(params []symbol.Param) string {
	return joinParamNames(params)
}

func joinParamNames(params []symbol.Param) string {
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
