# AIL Linter

The linter lives in `packages/linter` and is an independent advisory engine. It never gates correctness; that is the validator's job.

## Categories

| Category | Default severity | Kind | Detects |
|----------|------------------|------|---------|
| `duplicate-rule` | ERROR | semantic | Repeated rule IDs |
| `redundant-instruction` | WARNING | style | A repeated term in one block |
| `contradictory-instruction` | ERROR | semantic | A term in both DO and NEVER |
| `ambiguous-condition` | WARNING | semantic | A term combined with its negation |
| `unnecessary-nesting` | HINT | style | Deeply nested expression trees |
| `unused-declaration` | INFO | style | Context field never referenced |
| `unreachable-rule` | WARNING | semantic | Condition permanently false |
| `missing-fallback` | WARNING | style | Instructions without fallback |
| `overly-broad-scope` | WARNING | style | Exact `*`, `**` or `**/*` scope |
| `excessive-complexity` | INFO | style | More than 20 instruction units |
| `inconsistent-naming` | HINT | style | Rule ID not `Rnnn` or term not snake_case |
| `unnecessary-verbosity` | HINT | style | Same term in DO and VERIFY |
| `ineffective-instruction` | INFO | style | Vague words such as `consider` |

## Severity configuration

`Lint` accepts a `core.LintConfig` map from category to severity. A configured value overrides the default. Unspecified categories keep their defaults.

```go
result := linter.Lint(parsed, ir, core.LintConfig{
    core.LintMissingFallback: core.SeverityError,
})
```

## Semantic versus style

Every diagnostic carries a `Kind` of `semantic` or `style`. Semantic findings describe correctness problems. Style findings are preferences and should not block a build unless explicitly promoted through configuration.

## False-positive discipline

Prefixed globs such as `src/**/*.ts` are not considered broad. Uppercase single-part identifiers are treated as enum constants, not undefined identifiers. Valid rule sets with complete DO, NEVER, VERIFY and FALLBACK blocks produce zero ERROR findings.
