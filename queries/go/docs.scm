; Go tree-sitter query for doc comment extraction
; Captures functions, methods, structs, and interfaces.

(function_declaration
  name: (identifier) @name
  parameters: (parameter_list) @params
  result: (_)? @return) @func

(method_declaration
  receiver: (parameter_list) @receiver
  name: (field_identifier) @name
  parameters: (parameter_list) @params
  result: (_)? @return) @method

(type_declaration
  (type_spec
    name: (type_identifier) @name
    (struct_type)) @struct)

(type_declaration
  (type_spec
    name: (type_identifier) @name
    (interface_type)) @interface)
