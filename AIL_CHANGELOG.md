# Changelog

All notable changes to the AIL toolchain are documented here. Versions follow semantic versioning.

## 0.3.0

### Added

- Go monorepo layout with library code under `packages/` and the `ail` executable under `cmd/ail`.
- Hand-written lexer with source locations and recoverable diagnostics.
- Recursive-descent parser supporting standalone and aggregated modes, case-insensitive keywords and optional colons.
- Interpreter producing deterministic IR with priority, scope, inheritance, dependency and provenance resolution.
- Independent linter covering thirteen diagnostic categories with severity overrides.
- Independent validator returning machine-readable results.
- Renderer for Cursor, OpenCode, AGENTS.md and generic Markdown with four injection modes.
- Automatic mandatory delimiters and nested-delimiter escaping.
- Generated AIL and Markdown instruction sets for this project with cross-validation tests.
- Cross-platform build scripts for Windows, macOS and Linux.
- Benchmark task catalogue and OpenCode runner scaffold.
- Paired statistical analysis with confidence intervals, McNemar test, effect sizes and Holm correction.
- Rule-leakage scoring rubric that separates legitimate disclosure from unsolicited reproduction.
- Pre-execution verification of the four requested model identifiers through `opencode models`.
- Leakage task catalogue with six scenarios and explicit-rule-request flags.
- Data-driven report generator with coverage, paired statistics, leakage rates and compliance.
- Benchmark subsystem reorganized under `packages/benchmarks`, with reports, task catalogues and result summaries at the package root and code split into runner, stats, leakage, report and CLI subfolders.
