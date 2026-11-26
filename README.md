# AIL - Agent Instruction Language

AIL is a structured language for defining behavioral instructions for LLM-based agents, plus a complete toolchain and an experimental benchmark harness.

Version: 0.3

## Layout

This repository is a Go monorepo. All library code lives in `packages/`:

```text
packages/core         shared data model, diagnostics, expression normalization
packages/parser       hand-written lexer and recursive-descent parser
packages/interpreter  AST to deterministic IR resolution
packages/linter       independent advisory static analysis
packages/validator    independent structural and semantic validation
packages/renderer     host-specific rendering with mandatory delimiters
packages/cli          command line wiring
cmd/ail               executable entry point
TESTS                 integration, golden, regression, property and fuzz tests
BENCHMARK             tasks, runner, stats, leakage scorer, datasets, results and reports
AIL_RULES             generated AIL instruction set for this project
MARKDOWN_BASELINE     equivalent conventional Markdown instruction set
```

## Build

Go 1.25 or newer is required.

```powershell
go build ./...
go build -o dist/ail-windows-amd64.exe ./cmd/ail
```

Cross-platform targets are handled by `build.ps1` (Windows) and `build.sh` (macOS and Linux). Both produce binaries for Windows, macOS and Linux under `dist/`.

## Test

```powershell
go test ./...
go vet ./...
```

## CLI

```powershell
go run ./cmd/ail check AIL_RULES/project-rules.ail
go run ./cmd/ail render AIL_RULES/project-rules.ail --host=opencode --mode=VERBOSE
go run ./cmd/bench models
go run ./cmd/bench gentasks -leakage
go run ./cmd/bench stats -output=BENCHMARK/results/pilot
```

Commands: `parse`, `interpret`, `lint`, `validate`, `render`, `check`.

Options: `--host=cursor|opencode|agents-md|markdown`, `--mode=INLINE|FILE|COMPACT|VERBOSE`, `--no-delimiters`, `--provenance`.

## Safety model

AIL files are declarative data. The parser and interpreter never execute code embedded in a rule file. Unsupported constructs produce explicit diagnostics. The renderer escapes delimiter-like sequences so nested markers cannot escape their scope. Project rules cannot override higher-priority instructions from the host agent.

## Documentation

See `AIL_SPEC.md`, `AIL_GRAMMAR.md`, `AIL_INTERPRETER.md`, `AIL_LINTER.md`, `AIL_VALIDATOR.md`, `AIL_RENDERER.md`, `AIL_INJECTION.md`, `AIL_CHANGELOG.md` and `REPRODUCIBILITY.md`.
