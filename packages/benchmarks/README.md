# Benchmarks

This package contains the AIL versus Markdown experiment. Run all commands from the repository root.

## Root files

Kept at the root of this package for visibility:

| File | Purpose |
|------|---------|
| `BENCHMARK_REPORT.md` | Final report with coverage, paired statistics and answers to the research questions |
| `BENCHMARK_TASKS.json` | 50 main tasks across seven categories |
| `BENCHMARK_LEAKAGE_TASKS.json` | 20 leakage tasks across six scenarios |
| `RESULTS.json` | Pilot records, machine-readable |
| `RESULTS.csv` | Pilot records, tabular |

## Subfolders

| Folder | Contents |
|--------|----------|
| `cmd/bench` | CLI entry point with `gentasks`, `run`, `verify`, `stats`, `models` and `report` |
| `runner` | OpenCode invocation, JSONL event parsing, isolated workspaces, resume, verification and compliance |
| `stats` | Paired statistics, confidence intervals, McNemar test, effect sizes and Holm correction |
| `leakage` | Scoring rubric that separates legitimate disclosure from unsolicited rule reproduction |
| `report` | Data-driven report generator (called by `cmd/bench report`) |
| `datasets/workspaces` | Regenerated isolated workspaces, ignored by git |
| `results/pilot`, `results/full`, `results/leakage` | Per-execution records and raw OpenCode event streams |

## Commands

```powershell
go run ./packages/benchmarks/cmd/bench models
go run ./packages/benchmarks/cmd/bench gentasks
go run ./packages/benchmarks/cmd/bench gentasks -leakage
go run ./packages/benchmarks/cmd/bench run -mode=pilot
go run ./packages/benchmarks/cmd/bench run -mode=full
go run ./packages/benchmarks/cmd/bench run -mode=leakage -tasks packages/benchmarks/BENCHMARK_LEAKAGE_TASKS.json
go run ./packages/benchmarks/cmd/bench verify -output packages/benchmarks/results/full
go run ./packages/benchmarks/cmd/bench stats -output=packages/benchmarks/results/full
go run ./packages/benchmarks/cmd/bench report -output packages/benchmarks/results/full
```

## Notes

Token values are recomputed from raw `*.events.jsonl` streams. Parser fixes never require re-running model calls; run `verify` to refresh records and summaries.
