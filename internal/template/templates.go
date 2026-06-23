package template

func GetDefaultTemplates() map[string]string {
	return map[string]string{
		"go":     goTemplate,
		"python": pythonTemplate,
		"c":      cTemplate,
		"cpp":    cppTemplate,
		"rust":   rustTemplate,
	}
}

func GetTemplate(lang string) (string, bool) {
	tpls := GetDefaultTemplates()
	tpl, ok := tpls[lang]
	return tpl, ok
}

var goTemplate = `// {{.Name}} {{.Brief}}
{{- if .Aligned}}
//
{{- range .Aligned}}
// {{.Name}} - {{.Desc}}
{{- end}}
{{- end}}
{{- if .Returns}}
//
{{- range $i, $r := .Returns}}
// returns {{$r}}: {{returnDesc $.Returns}}
{{- end}}
{{- end}}
{{- if .CustomTags}}
//
{{- range .CustomTags}}
// {{.}}:
{{- end}}
{{- end}}`

var pythonTemplate = `"""{{.Brief}}
{{- if .Params}}

Args:
{{- range $i, $p := .Params}}
    {{$p.Name}} ({{$p.Type}}): {{paramDesc $.Params $i}}
{{- end}}
{{- end}}
{{- if .Returns}}

Returns:
{{- range $i, $r := .Returns}}
    {{$r}}: {{returnDesc $.Returns}}
{{- end}}
{{- end}}
{{- if .CustomTags}}
{{- range .CustomTags}}

{{.}}:
{{- end}}
{{- end}}
"""`

var cTemplate = `/**
 * @brief {{.Brief}}
{{- if .Aligned}}
{{- range .Aligned}}
 * @param {{.Name}} {{.Desc}}
{{- end}}
{{- end}}
{{- if .Returns}}
 * @return {{index .Returns 0}} {{returnDesc $.Returns}}
{{- end}}
{{- if .CustomTags}}
{{- range .CustomTags}}
 * @{{.}}
{{- end}}
{{- end}}
 */`

var cppTemplate = `/**
 * @brief {{.Brief}}
{{- if .Aligned}}
{{- range .Aligned}}
 * @param {{.Name}} {{.Desc}}
{{- end}}
{{- end}}
{{- if .Returns}}
 * @return {{index .Returns 0}} {{returnDesc $.Returns}}
{{- end}}
{{- if .CustomTags}}
{{- range .CustomTags}}
 * @{{.}}
{{- end}}
{{- end}}
 */`

var rustTemplate = `/// {{.Brief}}` + "\n" +
	`{{- if .Params}}` + "\n" +
	`///` + "\n" +
	`{{- range $i, $p := .Params}}` + "\n" +
	"/// * `{{$p.Name}}` - {{paramDesc $.Params $i}}" + "\n" +
	`{{- end}}` + "\n" +
	`{{- end}}` + "\n" +
	`{{- if .Returns}}` + "\n" +
	`{{- if .Params}}` + "\n" +
	`///` + "\n" +
	`{{- end}}` + "\n" +
	`/// Returns: {{returnDesc $.Returns}}` + "\n" +
	`{{- end}}` +
	`{{- if .CustomTags}}` + "\n" +
	`///` + "\n" +
	`{{- range .CustomTags}}` + "\n" +
	`/// {{.}}:` + "\n" +
	`{{- end}}` + "\n" +
	`{{- end}}`
