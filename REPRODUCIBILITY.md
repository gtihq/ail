# Reproducibility

## Environment

- Go 1.25.5 or newer.
- OpenCode 2.3.2 for benchmark executions.
- No third-party Go dependencies. The module has no `require` entries.

## Build

```powershell
go build ./...
```

Cross-platform binaries:

```powershell
powershell -File build.ps1
```

```bash
bash build.sh
```

Outputs are written to `dist/` for Windows amd64, macOS arm64 and amd64, and Linux amd64.

## Test

```powershell
go test ./...
go vet ./...
```

Tests are deterministic. Fuzz and property tests use fixed random seeds. Golden files live in `TESTS/fixtures/`.

## Cross-validating instruction sets

```powershell
go test ./TESTS -run TestGeneratedInstructionSetsAreEquivalent -v
```

This verifies that every instruction term in `AIL_RULES/project-rules.ail` appears in `MARKDOWN_BASELINE/project-rules.md` and that the AIL file validates without errors.

## Regenerating the golden render file

```powershell
$text = & go run ./cmd/ail render TESTS/fixtures/valid-standalone.ail --host=markdown --mode=INLINE
[System.IO.File]::WriteAllText("TESTS/fixtures/golden-render.md", ($text -join "`n") + "`n", (New-Object System.Text.UTF8Encoding($false)))
```

## Benchmark reproducibility

See `packages/benchmarks/README.md` and `packages/benchmarks/BENCHMARK_REPORT.md`. Raw results are stored unmodified under `packages/benchmarks/results/`. The runner supports resume so interrupted experiments can continue without re-running completed executions.
