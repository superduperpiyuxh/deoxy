; C tree-sitter query for doc comment extraction
; Captures function definitions and forward declarations.

(function_definition
  type: (_) @return_type
  declarator: (function_declarator
    declarator: (identifier) @name
    parameters: (parameter_list) @params) @func)

(declaration
  type: (_) @return_type
  declarator: (function_declarator
    declarator: (identifier) @name
    parameters: (parameter_list) @params) @func)
