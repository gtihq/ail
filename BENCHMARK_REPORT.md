# Benchmark Report - AIL versus Markdown

Status: pilot complete, full experiment not started.

## 1. Models verified

All four requested models were discovered in the configured OpenCode providers before execution. No substitutions were made.

| Requested label | Exact identifier |
|-----------------|------------------|
| Hy4 Flash Free | `opencode-go/hy4-flash-free` |
| Beam Free | `opencode-go/beam-free` |
| DeepSeek v4.1 Flash | `opencode-go/deepseek-v4.1-flash` |
| HungerG Free | `opencode-go/hungerg-free` |

The pilot used the first two models. DeepSeek v4.1 Flash and HungerG Free were not executed.

## 2. Exact coverage

| Item | Planned | Executed | Coverage |
|------|---------|----------|----------|
| Main experiment | 400 | 8 | 2.0% |
| Leakage experiment | 320 | 0 | 0% |
| Models | 4 | 2 | 50% |
| Tasks | 50 | 2 | 4% |
| Conditions | 2 of 4 (A and D) | 2 | 100% of pilot scope |

The 8 pilot executions were 2 models x 2 conditions x 2 tasks. Condition A is Markdown without delimiters. Condition D is AIL with delimiters.

## 3. Pilot results

All 8 executions succeeded and all 8 verification commands (`go test ./...`) passed. There were zero infrastructure failures.

| Condition | n | Input mean | Output mean | Reasoning mean | Cache read mean | Duration mean | Tool calls mean |
|-----------|---|------------|-------------|----------------|-----------------|---------------|-----------------|
| A (Markdown, no delimiters) | 4 | 1226.25 | 127.00 | 91.25 | 42368 | 82161.75 ms | 10.50 |
| D (AIL, with delimiters) | 4 | 375.50 | 182.75 | 187.00 | 46576 | 93848.00 ms | 9.75 |

Paired differences (D minus A), same task and model:

| Metric | Mean difference | 95% confidence interval | p (approx) | Cohen d | Significant |
|--------|-----------------|-------------------------|------------|---------|-------------|
| Input tokens | -850.75 | -3758.64 to +2057.14 | 0.3519 | -0.465 | no |
| Output tokens | +55.75 | -271.16 to +382.66 | 0.5874 | 0.271 | no |
| Reasoning tokens | +95.75 | -319.21 to +510.71 | 0.4628 | 0.367 | no |
| Cache read | +4208.00 | -16102.14 to +24518.14 | 0.5097 | 0.330 | no |
| Duration | +11686.25 ms | -152373.26 to +175745.76 | 0.8207 | 0.113 | no |
| Tool calls | -0.75 | -7.91 to +6.41 | 0.7389 | -0.167 | no |

These values are reproducible with `go run ./cmd/bench stats -output=BENCHMARK/results/pilot`. Multiple-comparison correction was not applied to this pilot because no family of confirmatory tests was declared in advance. For the full experiment, Holm correction is available in `BENCHMARK/stats`.

The input token difference is driven almost entirely by one pair: Hy4 Flash Free on AMB-03 used 4384 input tokens with Markdown and 797 with AIL. Removing or adding a single pair changes the direction of the result. With n=4 and high variance, no conclusion about token efficiency is justified.

Cache reads are large (39k to 66k tokens) and similar across conditions. They represent the cached system and instruction context, not task-specific text.

## 4. Compliance and correctness

- Completion rate: 8/8 (100%).
- Test pass rate: 8/8 (100%).
- Machine-checkable compliance: the four debugging tasks declared R003, which was followed in 4/4 cases. The four ambiguous tasks declared only subjective rules (R005, R006, R008), so no machine-checkable rule applied. Subjective compliance was not scored.
- Rule-leakage scoring was applied to all 8 pilot responses with no explicit rule request. Unsolicited rule reproduction was 0/8 and the correct-behavior rate was 100%. This is not the leakage experiment, which requires 20 tasks per model and condition.
- Unsolicited modifications, unsupported assumptions and fabricated claims were not scored by an automated rubric in the pilot.

## 5. Economic metrics

All four models are free in the current OpenCode configuration. Recorded cost is 0.000000 for every execution. Pricing is therefore unknown and no cost comparison is possible. Cost per successful task is reported as 0 with the explicit note that the underlying price is unavailable.

## 6. Statistical method

Paired comparisons were planned as specified. For the pilot, means, paired differences and approximate 95% confidence intervals using the t distribution with 3 degrees of freedom were computed. McNemar tests and multiple-comparison corrections were not applied because the sample has no binary outcome variation and no family of tests to correct. Effect sizes are not reported because the estimates are unstable at n=4.

## 7. Limitations

- Sample size is far below the planned 50 tasks per cell.
- Only two of four models were exercised.
- The leakage experiment (four conditions, rule reproduction scoring) was not run.
- Functional correctness was measured only as `go test ./...` exit status, not as a rubric-scored review.
- Code quality metrics (maintainability, readability, architecture consistency, security, error handling, type safety) were not scored.
- Instruction-token separation was not available from the provider events.

## 8. What the pilot did validate

- OpenCode runs through `opencode run --format json` with exact model identifiers.
- Newline-delimited JSON events are parsed for text, tokens, cost, session and tool calls.
- Each execution gets an isolated workspace with a fresh baseline.
- Completed successful records are reused on resume.
- Raw event streams are preserved under `BENCHMARK/results/pilot/`.
- Verification and compliance scoring run after each execution without additional model calls.
- Summaries are written to `RESULTS.json` and `RESULTS.csv`.

## 9. Decision required

The full experiment would add 392 main executions and 320 leakage executions. The pilot averaged about 82 seconds per execution for condition A and 94 seconds for condition D, so the main run alone would take roughly 10 hours of wall clock time. Full execution has not been started. It will only start after explicit confirmation.
