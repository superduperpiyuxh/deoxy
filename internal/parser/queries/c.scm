; C tree-sitter query for doc comment extraction
; Captures function definitions and forward declarations.

(function_declarator
  declarator: (identifier) @name
  parameters: (parameter_list) @params) @func

(function_definition
  declarator: (function_declarator
    declarator: (identifier) @name
    parameters: (parameter_list) @params) @func
  type: (_) @return_type)
