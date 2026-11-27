# AIL

**Agent Instruction Language** - a structured, declarative way to define behavioral instructions for LLM agents, with a compiler toolchain and a controlled benchmark that measures the effect against conventional Markdown.

Version 0.3. Implemented in Go. No third-party dependencies.

---

## Why AIL

In a paired experiment of **400 executions** across four models and 50 programming tasks, AIL instructions used substantially fewer resources than conventional Markdown instructions. Every result below is measured from raw OpenCode event streams and survives multiple-comparison correction.

| Metric | Markdown | AIL | Reduction | Significance |
|--------|----------:|----:|----------:|--------------|
| Input tokens | 55,761 | 45,647 | **-18.1%** | p = 0.0074 |
| Output tokens | 1,270 | 771 | **-39.3%** | p < 0.0001 |
| Reasoning tokens | 2,044 | 1,393 | **-31.9%** | p < 0.0001 |
| Cache read tokens | 566,227 | 445,600 | **-21.3%** | p = 0.0002 |
| Tool calls | 14.72 | 10.39 | **-29.4%** | p < 0.0001 |

The instruction payload itself is **18.9% shorter in characters** and **20.1% shorter in words** than the Markdown baseline. Using output tokens as a cost proxy, AIL is about **39% cheaper per execution**.

Five of six paired metrics are significant after Holm-Bonferroni correction. Effect sizes are small to medium, with output tokens (d = -0.529) and tool calls (d = -0.412) the strongest.

Full analysis: `packages/benchmarks/BENCHMARK_REPORT.md`.

### What the experiment did not show

Task success was essentially equal (75.0% Markdown, 74.5% AIL), and execution time did not improve. DeepSeek v4.1 Flash was blocked by a provider weekly usage cap and contributed zero successful runs, which is reported as an infrastructure limitation, not a model result. Code quality, subjective compliance and compact rendering were not scored. The toolchain is effective at reducing token and tool usage; claims about better task outcomes are not made.

---

## Quick start

```powershell
# Validate the generated rule set
go run ./cmd/ail check .ail/rules/project-rules.ail

# Render instructions for an agent host
go run ./cmd/ail render .ail/rules/project-rules.ail --host=opencode --mode=VERBOSE

# Verify the four requested models are configured
go run ./packages/benchmarks/cmd/bench models

# Run a benchmark pilot
go run ./packages/benchmarks/cmd/bench run -mode=pilot

# Recompute statistics from stored results
go run ./packages/benchmarks/cmd/bench stats -output=packages/benchmarks/results/full
```

Go 1.25 or newer is required. Build with `go build ./...` and test with `go test ./...`.

Cross-platform executables for Windows, macOS and Linux are produced by `build.ps1` and `build.sh` into `dist/`.

---

## Architecture

```text
packages/core         shared data model, diagnostics, expression normalization
packages/parser       hand-written lexer and recursive-descent parser
packages/interpreter  deterministic IR with inheritance and dependency resolution
packages/linter       independent advisory analysis, 13 diagnostic categories
packages/validator    independent structural and semantic gate
packages/renderer     Cursor, OpenCode, AGENTS.md and Markdown output
packages/cli          command line wiring
packages/benchmarks   runner, statistics, leakage rubric, report generator, results
cmd/ail               executable entry point
tests                 integration, golden, regression, property and fuzz suites
.ail/rules            generated AIL instruction set
```

The parser is a real lexer plus recursive-descent parser with source spans, not regex substitution. The interpreter resolves priorities, scopes, inheritance and dependencies into deterministic IR without executing embedded code. Renderers emit mandatory delimiters automatically and escape nested marker sequences. The benchmark runner isolates every execution in a fresh workspace, records token usage from provider events and supports resume.

---

## Engineering quality

- All packages covered by unit tests, plus golden, regression, property-based and malformed-input fuzz tests.
- Deterministic parsing, interpretation and rendering.
- Machine-readable validation: `{ "valid": false, "errors": [], "warnings": [], "summary": { "rules": 12, "errors": 1, "warnings": 3 } }`.
- Cross-platform binaries for Windows amd64, macOS arm64 and amd64, and Linux amd64.
- Raw experiment data preserved next to every record under `packages/benchmarks/results/`.

---

## Documentation

All specifications live in `packages/benchmarks/`: `AIL_SPEC.md`, `AIL_GRAMMAR.md`, `AIL_INTERPRETER.md`, `AIL_LINTER.md`, `AIL_VALIDATOR.md`, `AIL_RENDERER.md`, `AIL_INJECTION.md`, `AIL_CHANGELOG.md`, plus `BENCHMARK_REPORT.md` and `MARKDOWN_BASELINE.md`. Build and reproduction details are in `REPRODUCIBILITY.md`.
