# AIL Renderer

The renderer lives in `packages/renderer`. It converts IR into instructions for a host agent. Host agents are not expected to understand AIL natively, so output is plain Markdown or Cursor frontmatter.

## Host formats

| Host | Output shape |
|------|--------------|
| `cursor` | `.mdc` frontmatter with `description`, `globs`, `alwaysApply` |
| `opencode` | Markdown with an `# AIL Operational Instructions` heading |
| `agents-md` | Aggregated Markdown with a `# Project Agent Instructions` heading |
| `markdown` | Generic Markdown with an `# AIL Instructions` heading |

## Injection modes

| Mode | Behavior |
|------|----------|
| `INLINE` | Structured bullet form with all fields |
| `FILE` | Structured form plus a note that the host loads it through native file discovery |
| `COMPACT` | One line per rule containing every field |
| `VERBOSE` | Structured form plus explanations, context values, provenance and inheritance |

COMPACT and VERBOSE preserve the same semantic intent: every DO, NEVER, VERIFY, FALLBACK, WHEN and ASSERT item appears in both.

## Options

```go
result := renderer.Render(ir, core.RenderOptions{
    Host:              core.HostOpenCode,
    Mode:              core.ModeVerbose,
    IncludeDelimiters: &enabled,
    IncludeProvenance: true,
})
```

Delimiters are generated automatically unless `IncludeDelimiters` is explicitly set to false. Unknown host or mode values produce diagnostics and fall back to `markdown` and `INLINE`.

## Determinism

Rules render in sorted IR order. Context keys are sorted. Rendered text for identical IR is byte-identical across runs.
