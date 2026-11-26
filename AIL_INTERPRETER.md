# AIL Interpreter

The interpreter lives in `packages/interpreter` and converts a `ParseResult` into deterministic IR. It never executes instructions.

## Responsibilities

- Resolve metadata: META id, version, default priority, default scope.
- Generate missing rule IDs from the META id or as `Rnnn`.
- Resolve scopes and priorities with META as the default layer.
- Resolve inheritance recursively, with child fields overriding parent fields.
- Resolve dependencies declared through `REQUIRES`.
- Normalize equivalent constructs into flat string lists.
- Detect unresolved references and inheritance cycles.
- Preserve source provenance for every rule.
- Produce deterministic, sortable output.

## IR shape

```json
{
  "kind": "IR",
  "version": "0.3",
  "rules": [
    {
      "id": "R001",
      "priority": "HIGH",
      "scope": "src/**/*.ts",
      "context": {},
      "when": ["file.language == TYPESCRIPT"],
      "do": ["use_strict_types"],
      "never": ["use_any"],
      "verify": [],
      "fallback": [],
      "requires": [],
      "inherit": [],
      "assert": [],
      "execution": "",
      "source": { "file": "rules.ail", "line": 1, "column": 1 },
      "inheritedFrom": []
    }
  ],
  "provenance": [{ "file": "rules.ail", "rules": ["R001"] }]
}
```

## Diagnostics

| Code | Severity | Meaning |
|------|----------|---------|
| `INT_DUPLICATE_RULE` | ERROR | Two rules share an ID |
| `INT_INHERITANCE_CYCLE` | ERROR | Inheritance graph contains a cycle |
| `INT_UNRESOLVED_INHERIT` | WARNING | Parent rule is undefined |
| `INT_UNRESOLVED_REFERENCE` | WARNING | Required rule is undefined |

## Non-goals

The interpreter does not evaluate expressions, does not run shell commands, does not read files referenced by scopes and does not modify the host agent configuration.
