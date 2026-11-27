# AIL Grammar - version 0.3

## 1. Lexical rules

Characters are scanned one byte at a time with line, column and offset tracking.

| Category | Rule |
|----------|------|
| Whitespace | space, tab, carriage return, line feed |
| Line comment | `//` until end of line |
| Block comment | `/* ... */`, may span lines |
| Identifier | `[A-Za-z_][A-Za-z0-9_]*` |
| Number | digits with optional fractional part |
| String | double quotes, escapes: `\\`, `\"`, `\n`, `\t`, `\r` |
| Punctuation | `{}` `()` `:` `,` `.` |
| Operators | `&&` `\|\|` `!` `!=` `==` `>` `<` `>=` `<=` |

Keywords are contextual identifiers resolved by the parser, case-insensitively. A single `=`, `&` or `|` is a lexical error. Unknown characters are reported and skipped so scanning always terminates.

## 2. File grammar

```text
file           := meta? (rule | field)*
meta           := META LBRACE field* RBRACE
rule           := RULE IDENT? LBRACE field* RBRACE
field          := priority | context | when | assert | instruction
                | refs | scope | execution
priority       := PRIORITY COLON? PRIORITY_VALUE
scope          := SCOPE COLON? STRING
context        := CONTEXT COLON? LBRACE field* RBRACE
when           := WHEN COLON? expression_list
assert         := ASSERT COLON? expression_list
instruction    := (DO | NEVER | VERIFY | FALLBACK) COLON? LBRACE term* RBRACE
refs           := (REQUIRES | INHERIT) COLON? IDENT*
execution      := EXECUTION COLON? IDENT
expression_list:= LBRACE expression* RBRACE | expression
term           := IDENT | STRING | NUMBER
```

A file is aggregated when `RULE` appears at brace depth zero. Otherwise it is standalone and all top-level blocks become fields of one implicit rule.

## 3. Expression grammar

Precedence from lowest to highest:

```text
expression     := or
or             := and ( (OR | "or") and )*
and            := comparison ( (AND | "and") comparison )*
comparison     := unary ( (EQ | NEQ | GT | LT | GTE | LTE | "in") unary )?
unary          := (NOT | "not") unary | primary
primary        := literal | path | LPAREN expression RPAREN
literal        := STRING | NUMBER | "true" | "false"
path           := IDENT (DOT IDENT)*
```

Operators: `&&`, `||`, `==`, `!=`, `>`, `<`, `>=`, `<=`, `in`, `!`.

## 4. Data types

| Type | Example | Notes |
|------|---------|-------|
| string | `"src/**/*.ts"` | Supports standard escapes |
| number | `700` | Integer or decimal |
| boolean | `true`, `false` | Lowercase only |
| path | `file.language` | Dotted identifiers in expressions |
| priority | `CRITICAL` | Enum-like identifier |
| instruction term | `use_strict_types` | Opaque declarative token |

## 5. Scope resolution

A `SCOPE` value is a glob string. META scope becomes the default for rules without an explicit scope. Exact `*`, `**` and `**/*` are flagged as overly broad by the linter. Empty scopes and scopes containing `..` or whitespace are validator errors.

## 6. Priority resolution

META priority is the default. A rule `PRIORITY` field overrides it. Unrecognized priorities are parser errors and resolve to `MEDIUM`.

## 7. Inheritance resolution

`INHERIT` names parent rules. Parent fields are copied into the child only when the child does not declare that field. Child declarations always win. Cycles produce `INT_INHERITANCE_CYCLE` and `VAL_INHERITANCE_CYCLE`.

## 8. Determinism

Rules are sorted by ID. Expression text is normalized without whitespace variance. Context maps are emitted with sorted keys. Repeated parsing and rendering of the same input produces byte-identical output.
