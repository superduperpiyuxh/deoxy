; Rust tree-sitter query for doc comment extraction
; Captures functions, structs, traits, impls, and enums.

(function_item
  name: (identifier) @name
  parameters: (parameters) @params
  return_type: (_)? @return) @func

; Methods inside impl blocks — function_item is inside declaration_list
(impl_item
  (declaration_list
    (function_item
      name: (identifier) @name
      parameters: (parameters) @params
      return_type: (_)? @return) @method))

(struct_item
  name: (type_identifier) @name) @struct

(trait_item
  name: (type_identifier) @name) @trait

(enum_item
  name: (type_identifier) @name) @enum
