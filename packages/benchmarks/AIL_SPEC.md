# AIL Specification - version 0.3

AIL (Agent Instruction Language) is a structured language for defining behavioral instructions for LLM-based agents. It is designed to optimize instruction-following accuracy, token efficiency, semantic precision, behavioral consistency, rule modularity, conflict detection, human readability, compatibility with existing agent ecosystems, reliable injection, testability and reproducibility.

AIL is a specification language, not a programming language. AIL files contain declarative instructions only. The toolchain never executes code embedded in a rule file.

## 1. Operating modes

### STANDALONE

One file represents one rule or instruction unit. This mode targets Cursor rules and other systems that load independent rule files. `RULE` blocks are optional and should normally be omitted.

```ail
META { id: "typescript-conventions" priority: HIGH scope: "src/**/*.ts" }
WHEN { file.language == TYPESCRIPT }
DO { use_strict_types preserve_existing_patterns }
NEVER { use_any_without_justification }
```

The parser detects standalone mode when no `RULE` keyword appears at top level. Top-level blocks are folded into one implicit rule.

### AGGREGATED

One file contains multiple independent rules. This mode targets `AGENTS.md` and consolidated instruction files. `RULE` blocks should be used.

```ail
META { id: "project-rules" version: "0.3" }
RULE R001 { priority: HIGH DO { inspect_relevant_files } }
RULE R002 { priority: CRITICAL NEVER { expose_secrets } }
```

The parser never requires `RULE` wrappers in standalone files. Each host agent keeps its native rule discovery mechanism.

## 2. Constructs

| Construct | Purpose |
|-----------|---------|
| `META` | File or project metadata: id, version, default priority, default scope |
| `PRIORITY` | Rule urgency: CRITICAL, HIGH, MEDIUM, LOW, INFO |
| `CONTEXT` | Named values available to conditions |
| `RULE` | Named rule container in aggregated mode |
| `WHEN` | Conditions that gate the rule |
| `DO` | Required instructions |
| `NEVER` | Forbidden instructions |
| `VERIFY` | Checks that must hold after the rule is applied |
| `FALLBACK` | Behavior when the rule cannot be applied |
| `REQUIRES` | Rule dependencies that must exist |
| `SCOPE` | Glob describing affected files |
| `ASSERT` | Conditions used for validation |
| `INHERIT` | Parent rules whose fields are inherited |
| `EXECUTION` | Execution policy token: declarative, autonomous, interactive |

Keywords are case-insensitive and accept an optional colon, so `priority: HIGH` and `PRIORITY HIGH` are equivalent.

## 3. Declarative versus executable

Expressions in `WHEN` and `ASSERT` are normalized into text. Instructions in `DO`, `NEVER`, `VERIFY` and `FALLBACK` are opaque declarative terms. Neither is evaluated. Unsupported constructs produce explicit `PARSE_UNSUPPORTED_CONSTRUCT` or `LEX_` diagnostics and are never silently ignored.

## 4. Resolution order

1. Metadata defaults are read from `META`.
2. Missing rule IDs are generated from the META id or as `Rnnn`.
3. Inheritance is resolved recursively with cycle detection.
4. Child fields override inherited fields.
5. Dependencies are checked against the resolved rule set.
6. Rules are sorted by ID for deterministic output.

## 5. Error handling

Lexical and syntax errors are collected, not thrown, so one malformed block does not hide later diagnostics. Semantic problems are reported by the validator. Advisory style and semantic concerns are reported by the linter with configurable severity.

## 6. Version compatibility

Version 0.3 is the only version accepted by this toolchain. A different version produces `VAL_VERSION_INCOMPATIBLE`. The renderer always emits `AIL_VERSION: 0.3` in its markers.

## 7. Reference implementation

The Go implementation in `packages/` is the normative reference. See `packages/benchmarks/AIL_GRAMMAR.md` for the formal grammar and `REPRODUCIBILITY.md` for build and test commands.
