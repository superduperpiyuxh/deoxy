; Python tree-sitter query for doc comment extraction
; Captures functions and class definitions.

(function_definition
  name: (identifier) @name
  parameters: (parameters) @params
  return_type: (_)? @return) @func

(class_definition
  name: (identifier) @name) @class
