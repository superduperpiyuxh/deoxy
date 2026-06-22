; C++ tree-sitter query for doc comment extraction
; Captures functions, classes, and structs.

(function_declarator
  declarator: (identifier) @name
  parameters: (parameter_list) @params) @func

(function_definition
  declarator: (function_declarator
    declarator: (identifier) @name
    parameters: (parameter_list) @params) @func)

(class_specifier
  name: (type_identifier) @name) @class

(struct_specifier
  name: (type_identifier) @name) @struct
