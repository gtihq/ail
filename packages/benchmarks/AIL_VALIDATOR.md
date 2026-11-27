# AIL Validator

The validator lives in `packages/validator`. It is an independent structural and semantic gate and returns machine-readable results. It shares no logic with the linter.

## Checks

| Area | Codes |
|------|-------|
| Syntax validity | Any ERROR diagnostic from lexer or parser is copied into `errors` |
| Required metadata | `VAL_MISSING_META`, `VAL_MISSING_META_ID`, `VAL_EMPTY_FIELD` |
| Supported types | `VAL_UNSUPPORTED_TYPE` |
| Valid operators | `VAL_INVALID_OPERATOR` |
| Reference resolution | `VAL_UNRESOLVED_REFERENCE`, `VAL_UNRESOLVED_INHERIT` |
| Rule dependencies | `VAL_DUPLICATE_RULE`, `VAL_EMPTY_RULE` |
| Priority consistency | `VAL_INVALID_PRIORITY` |
| Scope validity | `VAL_EMPTY_SCOPE`, `VAL_INVALID_SCOPE` |
| Inheritance cycles | `VAL_INHERITANCE_CYCLE` |
| Undefined identifiers | `VAL_UNDEFINED_IDENTIFIER` (warning) |
| Incompatible constructs | `VAL_INVALID_EXECUTION`, `VAL_VERSION_INCOMPATIBLE` |
| IR integrity | `VAL_MISSING_IR`, `VAL_IR_KIND`, `VAL_IR_VERSION`, `VAL_EMPTY_RULE_ID`, `VAL_MISSING_PROVENANCE` |

## Output

```json
{
  "valid": false,
  "errors": [],
  "warnings": [],
  "summary": { "rules": 12, "errors": 1, "warnings": 3 }
}
```

`valid` is true only when `errors` is empty. Warnings do not block.

## Builtin identifier roots

`file`, `env`, `platform`, `agent`, `project`, `scope`, `context`, `language`, `user`, `service`, `tooling` and `context` values declared in the rule are accepted. Single-part uppercase identifiers are treated as constants.
