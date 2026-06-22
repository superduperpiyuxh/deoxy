// Package symbol defines the core type system for parsed symbol information.
// These types are the contract between Phase 1 (parsing) and Phase 2 (template engine).
// No tree-sitter imports — pure Go types only.
package symbol

// Kind represents the classification of a parsed symbol.
type Kind int

const (
	KindFunction  Kind = iota // A standalone function
	KindMethod                // A method on a type (has a receiver)
	KindStruct                // A struct type definition
	KindClass                 // A class type definition (Python, C++)
	KindInterface             // An interface or trait definition
	KindEnum                  // An enum type definition
)

// String returns the lowercase name of the Kind.
func (k Kind) String() string {
	switch k {
	case KindFunction:
		return "function"
	case KindMethod:
		return "method"
	case KindStruct:
		return "struct"
	case KindClass:
		return "class"
	case KindInterface:
		return "interface"
	case KindEnum:
		return "enum"
	default:
		return "unknown"
	}
}

// Param represents a function/method parameter with name and type.
type Param struct {
	Name string // Parameter name (e.g., "a", "self", "name")
	Type string // Parameter type (e.g., "int", "string", "*MyStruct")
}

// SymbolInfo holds complete metadata for a parsed symbol.
// Fields are populated by QueryRunner from tree-sitter query captures.
type SymbolInfo struct {
	Name      string    // Symbol name (function name, struct name, etc.)
	Kind      Kind      // Kind of symbol (function, method, struct, etc.)
	Params    []Param   // Function/method parameters
	Returns   []string  // Return types (empty slice if no return)
	TypeParams []Param  // Generic/type parameters (e.g., <T, U>)
	Receiver  *Param    // Method receiver (nil for functions, set for methods)
	StartLine int       // 0-indexed byte line from tree-sitter
	EndLine   int       // 0-indexed byte line from tree-sitter
}
