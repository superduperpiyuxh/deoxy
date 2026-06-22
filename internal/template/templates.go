package template

// GetDefaultTemplates returns the default template map for all 5 supported
// languages. Keys: "go", "python", "c", "cpp", "rust".
// Each value is a Go text/template string that uses the engine's FuncMap
// helpers (brief, paramDesc, returnDesc, joinParams, commentPrefix) and
// receives a TemplateData context (Name, Brief, Params, Returns, etc.).
func GetDefaultTemplates() map[string]string {
	return map[string]string{
		"go":    goTemplate,
		"python": pythonTemplate,
		"c":     cTemplate,
		"cpp":   cppTemplate,
		"rust":  rustTemplate,
	}
}

// GetTemplate returns the default template for a single language.
// Returns empty string and false if the language is unknown.
func GetTemplate(lang string) (string, bool) {
	tpls := GetDefaultTemplates()
	tpl, ok := tpls[lang]
	return tpl, ok
}

// goTemplate is the GoDoc prose-style comment template.
// Uses "//" prefix. Brief sentence followed by blank line then
// per-parameter descriptions. No @param or @return tags — Go convention
// is prose description.
//
// Example output for func Add(a int, b int) int:
//
//	// Add adds a and b and returns the result.
//	//
//	// a - the first operand
//	// b - the second operand
var goTemplate = `// {{.Name}} {{.Brief}}
{{- if .Params}}
//
{{- range $i, $p := .Params}}
// {{$p.Name}} - {{paramDesc $.Params $i}}
{{- end}}
{{- end}}
{{- if .Returns}}
//
{{- range $i, $r := .Returns}}
// returns {{$r}}: {{returnDesc $.Returns}}
{{- end}}
{{- end}}`

// pythonTemplate is the Google-style Python docstring template.
// Uses """ delimiters with Args: and Returns: sections.
//
// Example output for def greet(name: str, age: int) -> str:
//
//	"""Greets a person with the given name and age.
//
//	Args:
//	    name (str): the name
//	    age (int): the age
//
//	Returns:
//	    str: the str result
//	"""
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
"""`

// cTemplate is the Doxygen-style block comment for C.
// Uses /** ... */ with * prefix on each line.
//
// Example output for int add(int a, int b):
//
//	/**
//	 * @brief adds a and b and returns the result.
//	 * @param a the first operand
//	 * @param b the second operand
//	 * @return int the result
//	 */
var cTemplate = `/**
 * @brief {{.Brief}}
{{- if .Params}}
{{- range $i, $p := .Params}}
 * @param {{$p.Name}} {{paramDesc $.Params $i}}
{{- end}}
{{- end}}
{{- if .Returns}}
 * @return {{index .Returns 0}} {{returnDesc $.Returns}}
{{- end}}
 */`

// cppTemplate is the Doxygen-style block comment for C++.
// Same format as C — Doxygen conventions apply to both.
//
// Example output for int free_function(int x, int y):
//
//	/**
//	 * @brief adds x and y and returns the result.
//	 * @param x the first operand
//	 * @param y the second operand
//	 * @return int the result
//	 */
var cppTemplate = `/**
 * @brief {{.Brief}}
{{- if .Params}}
{{- range $i, $p := .Params}}
 * @param {{$p.Name}} {{paramDesc $.Params $i}}
{{- end}}
{{- end}}
{{- if .Returns}}
 * @return {{index .Returns 0}} {{returnDesc $.Returns}}
{{- end}}
 */`

// rustTemplate is the Rustdoc-style with "///" prefix.
// Uses bullet-point parameter descriptions with backtick-quoted names.
//
// Example output for fn add(a: i32, b: i32) -> i32:
//
//	/// adds a and b and returns the result.
//	///
//	/// * `a` - the first operand
//	/// * `b` - the second operand
//	///
//	/// Returns: the result
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
	`{{- end}}`
