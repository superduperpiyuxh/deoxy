package smarttext

import (
	"strings"
	"unicode"

	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

func IsGetter(name string) bool {
	return strings.HasPrefix(name, "Get") && len(name) > 3 && !isLower(name[3])
}

func IsSetter(name string) bool {
	return strings.HasPrefix(name, "Set") && len(name) > 3 && !isLower(name[3])
}

func IsConstructor(name string) bool {
	return strings.HasPrefix(name, "New") && len(name) > 3 && !isLower(name[3])
}

func SplitCamelCase(name string) []string {
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i, p := range parts {
			parts[i] = strings.ToLower(p)
		}
		return parts
	}
	if len(name) == 0 {
		return nil
	}

	src := []rune(name)
	dst := make([]rune, len(src))
	for i, r := range src {
		dst[i] = unicode.ToLower(r)
	}

	var words []string
	start := 0

	for i := 1; i < len(src); i++ {
		if unicode.IsUpper(src[i]) && unicode.IsLower(src[i-1]) {
			words = append(words, string(dst[start:i]))
			start = i
		} else if i > 1 && unicode.IsUpper(src[i-1]) && unicode.IsLower(src[i]) && unicode.IsUpper(src[i-2]) {
			words = append(words, string(dst[start:i-1]))
			start = i - 1
		}
	}
	words = append(words, string(dst[start:]))

	return words
}

func Describe(name string, params []symbol.Param, returns []string, reg *Registry) string {
	if reg == nil {
		reg = NewRegistry()
	}
	if IsConstructor(name) {
		base := name[3:]
		words := SplitCamelCase(base)
		if len(words) > 0 {
			return name + " creates a new " + strings.Join(words, " ")
		}
		return name + " creates a new " + base
	}
	if IsGetter(name) {
		field := name[3:]
		words := SplitCamelCase(field)
		if len(words) > 0 {
			return "Gets the " + strings.Join(words, " ")
		}
		return "Gets the " + field
	}
	if IsSetter(name) {
		field := name[3:]
		words := SplitCamelCase(field)
		if len(words) > 0 {
			return "Sets the " + strings.Join(words, " ")
		}
		return "Sets the " + field
	}
	return ""
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}
