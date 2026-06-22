package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/superduperpiyuxh/deoxy/internal/symbol"
)

// RunQuery executes a tree-sitter query against a parsed tree and returns
// a slice of SymbolInfo extracted from the query captures.
//
// The queryContent is the .scm query string for the given language.
// src is the original source bytes used when parsing the tree.
// The caller is responsible for closing the Tree. QueryRunner does NOT close the Tree.
// All Query and QueryCursor objects are explicitly defer-closed within this function.
func RunQuery(tree *sitter.Tree, queryContent string, lang string, src []byte) ([]symbol.SymbolInfo, error) {
	langCfg, ok := GetLanguageConfig(lang)
	if !ok {
		return nil, fmt.Errorf("queryrunner: unsupported language %q", lang)
	}

	q, qErr := sitter.NewQuery(langCfg.GrammarAsLanguage(), queryContent)
	if qErr != nil {
		return nil, fmt.Errorf("queryrunner: invalid query for %s: %w", lang, qErr)
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	root := tree.RootNode()
	matches := cursor.Matches(q, root, src)

	var results []symbol.SymbolInfo
	for {
		match := matches.Next()
		if match == nil {
			break
		}

		info, err := extractSymbolInfo(match, q, lang, src)
		if err != nil {
			// Skip matches we can't interpret (e.g., impl blocks)
			continue
		}
		results = append(results, info)
	}

	return results, nil
}

// extractSymbolInfo maps a single query match's captures to a SymbolInfo struct.
func extractSymbolInfo(match *sitter.QueryMatch, q *sitter.Query, lang string, src []byte) (symbol.SymbolInfo, error) {
	captureNames := q.CaptureNames()

	// Build a map of capture name → node for this match
	captures := make(map[string]*sitter.Node)
	for _, cap := range match.Captures {
		if int(cap.Index) < len(captureNames) {
			name := captureNames[cap.Index]
			captures[name] = &cap.Node
		}
	}

	info := symbol.SymbolInfo{}
	hasTopCapture := false

	// Check for impl block top capture — skip, methods are captured separately
	if _, ok := captures["impl"]; ok {
		return info, fmt.Errorf("impl block — skipping, methods captured separately")
	}

	// @func capture: function-like nodes
	if _, ok := captures["func"]; ok {
		hasTopCapture = true
		if receiverNode, ok := captures["receiver"]; ok {
			info.Kind = symbol.KindMethod
			info.Receiver = parseReceiver(receiverNode, lang, src)
		} else {
			info.Kind = symbol.KindFunction
		}
		populateFromCaptures(&info, captures, lang, src)
	}

	// @method capture: methods (Rust, or explicit method capture)
	if _, ok := captures["method"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindMethod
		populateFromCaptures(&info, captures, lang, src)

		// Detect method receiver (self param) — strip from params
		if len(info.Params) > 0 {
			first := info.Params[0]
			if first.Name == "self" || first.Name == "&self" || first.Name == "&mut self" || first.Name == "self:" {
				info.Receiver = &symbol.Param{Name: first.Name, Type: "Self"}
				info.Params = info.Params[1:]
			}
		}
	}

	// @struct capture
	if _, ok := captures["struct"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindStruct
		if nameNode, ok := captures["name"]; ok {
			info.Name = nameNode.Utf8Text(src)
		}
	}

	// @class capture
	if _, ok := captures["class"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindClass
		if nameNode, ok := captures["name"]; ok {
			info.Name = nameNode.Utf8Text(src)
		}
	}

	// @interface capture (Go)
	if _, ok := captures["interface"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindInterface
		if nameNode, ok := captures["name"]; ok {
			info.Name = nameNode.Utf8Text(src)
		}
	}

	// @trait capture (Rust)
	if _, ok := captures["trait"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindInterface
		if nameNode, ok := captures["name"]; ok {
			info.Name = nameNode.Utf8Text(src)
		}
	}

	// @enum capture (Rust)
	if _, ok := captures["enum"]; ok {
		hasTopCapture = true
		info.Kind = symbol.KindEnum
		if nameNode, ok := captures["name"]; ok {
			info.Name = nameNode.Utf8Text(src)
		}
	}

	if !hasTopCapture {
		return info, fmt.Errorf("no recognized capture pattern in match")
	}

	// Set start/end lines from the first top-level capture node
	topNames := map[string]bool{"func": true, "method": true, "struct": true, "class": true, "interface": true, "trait": true, "enum": true}
	for _, cap := range match.Captures {
		if int(cap.Index) < len(captureNames) && topNames[captureNames[cap.Index]] {
			info.StartLine = int(cap.Node.StartPosition().Row)
			info.EndLine = int(cap.Node.EndPosition().Row)
			break
		}
	}

	return info, nil
}

// populateFromCaptures fills SymbolInfo.Name, Params, Returns from capture nodes.
func populateFromCaptures(info *symbol.SymbolInfo, captures map[string]*sitter.Node, lang string, src []byte) {
	if nameNode, ok := captures["name"]; ok {
		info.Name = nameNode.Utf8Text(src)
	}
	if paramsNode, ok := captures["params"]; ok {
		info.Params = parseParams(paramsNode, lang, src)
	}
	if returnNode, ok := captures["return"]; ok {
		info.Returns = parseReturnTypes(returnNode, lang, src)
	}
}

// parseParams extracts parameter names and types from a parameter_list or parameters node.
func parseParams(node *sitter.Node, lang string, src []byte) []symbol.Param {
	var params []symbol.Param

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}

		switch lang {
		case "go":
			param := extractGoParam(child, src)
			if param.Name != "" || param.Type != "" {
				params = append(params, param)
			}

		case "python":
			childKind := child.Kind()
			switch childKind {
			case "identifier":
				params = append(params, symbol.Param{Name: child.Utf8Text(src)})
			case "typed_parameter":
				params = append(params, extractPythonTypedParam(child, src))
			case "default_parameter":
				params = append(params, extractPythonDefaultParam(child, src))
			case "list_splat":
				if nn := child.NamedChild(0); nn != nil {
					params = append(params, symbol.Param{Name: "*" + nn.Utf8Text(src), Type: "..."})
				}
			case "dictionary_splat":
				if nn := child.NamedChild(0); nn != nil {
					params = append(params, symbol.Param{Name: "**" + nn.Utf8Text(src), Type: "..."})
				}
			}

		case "c", "cpp":
			childKind := child.Kind()
			if childKind == "parameter_declaration" {
				param := extractCParam(child, src)
				params = append(params, param)
			}

		case "rust":
			childKind := child.Kind()
			switch childKind {
			case "parameter":
				params = append(params, extractRustParam(child, src))
			case "self", "mut_self", "ampersand_self":
				params = append(params, symbol.Param{Name: child.Utf8Text(src), Type: "Self"})
			}
		}
	}

	// Deduplicate Go params where bare types appear (e.g., "a, b int" produces only one
	// parameter_declaration with name="a" but the "b" identifier is a sibling)
	// This happens when Go parser packs multiple params with same type into one node.
	// Actually, tree-sitter-go parses "a, b int" as TWO parameter_declaration nodes.
	// The second one has no name child and its text is just "int" (the type).
	// We need to handle this in extractGoParam.

	return params
}

// extractGoParam extracts a single Go parameter from a parameter_declaration node.
func extractGoParam(node *sitter.Node, src []byte) symbol.Param {
	param := symbol.Param{}

	// Look for name child (identifier)
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		param.Name = nameNode.Utf8Text(src)
	}

	// Look for type child
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		param.Type = typeNode.Utf8Text(src)
	}

	return param
}

// extractPythonTypedParam extracts a typed parameter (name: Type).
func extractPythonTypedParam(node *sitter.Node, src []byte) symbol.Param {
	param := symbol.Param{}
	if nn := node.ChildByFieldName("name"); nn != nil {
		param.Name = nn.Utf8Text(src)
	}
	if tn := node.ChildByFieldName("type"); tn != nil {
		param.Type = tn.Utf8Text(src)
	}
	return param
}

// extractPythonDefaultParam extracts a default parameter (name=value).
func extractPythonDefaultParam(node *sitter.Node, src []byte) symbol.Param {
	param := symbol.Param{}
	if nn := node.ChildByFieldName("name"); nn != nil {
		param.Name = nn.Utf8Text(src)
	}
	if tn := node.ChildByFieldName("type"); tn != nil {
		param.Type = tn.Utf8Text(src)
	}
	return param
}

// extractCParam extracts a single C/C++ parameter.
func extractCParam(node *sitter.Node, src []byte) symbol.Param {
	param := symbol.Param{}

	if tn := node.ChildByFieldName("type"); tn != nil {
		param.Type = tn.Utf8Text(src)
	}
	if dn := node.ChildByFieldName("declarator"); dn != nil {
		param.Name = dn.Utf8Text(src)
	}

	// Handle void-only parameter
	if param.Type == "" && param.Name == "" {
		text := node.Utf8Text(src)
		if text != "void" {
			param.Name = text
		}
	}

	return param
}

// extractRustParam extracts a single Rust parameter.
func extractRustParam(node *sitter.Node, src []byte) symbol.Param {
	param := symbol.Param{}
	if nn := node.ChildByFieldName("name"); nn != nil {
		param.Name = nn.Utf8Text(src)
	}
	if tn := node.ChildByFieldName("type"); tn != nil {
		param.Type = tn.Utf8Text(src)
	}
	return param
}

// parseReturnTypes extracts return type strings from a return/result node.
func parseReturnTypes(node *sitter.Node, lang string, src []byte) []string {
	text := strings.TrimSpace(node.Utf8Text(src))
	if text == "" {
		return nil
	}

	switch lang {
	case "go":
		// Go return types may be multiple (func() (int, error))
		// For simplicity, return the full text
		return []string{text}
	case "python":
		return []string{text}
	case "rust":
		return []string{text}
	default:
		return []string{text}
	}
}

// parseReceiver extracts the receiver parameter (for Go methods).
func parseReceiver(node *sitter.Node, lang string, src []byte) *symbol.Param {
	if lang != "go" {
		return nil
	}
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Kind() == "parameter_declaration" {
			param := extractGoParam(child, src)
			return &param
		}
	}
	return nil
}
