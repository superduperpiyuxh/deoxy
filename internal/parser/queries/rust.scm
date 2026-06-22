; Rust tree-sitter query for doc comment extraction
; Captures functions, structs, traits, impls, and enums.

(function_item
  name: (identifier) @name
  parameters: (parameters) @params
  return_type: (_)? @return) @func

; Methods inside impl blocks
(impl_item
  (function_item
    name: (identifier) @name
    parameters: (parameters) @params
    return_type: (_)? @return) @method)

(struct_item
  name: (type_identifier) @name) @struct

(trait_item
  name: (type_identifier) @name) @trait

(impl_item
  trait: (_)? @impl_trait) @impl

(enum_item
  name: (type_identifier) @name) @enum
