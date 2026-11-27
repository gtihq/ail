# Benchmark Report - AIL versus Markdown

Version 0.3. Generated from raw data in `BENCHMARK/results/full/` and `BENCHMARK/results/pilot/`. All numbers below come from recorded executions. No result was estimated or fabricated.

## 1. Models verified and actually executed

Identifiers were confirmed with `opencode models` before execution.

| Requested label | Exact identifier | Executions | Successful |
|-----------------|------------------|------------|------------|
| Hy4 Flash Free | `opencode-go/hy4-flash-free` | 100 | 100 |
| Beam Free | `opencode-go/beam-free` | 100 | 99 |
| DeepSeek v4.1 Flash | `opencode-go/deepseek-v4.1-flash` | 100 | 0 |
| HungerG Free | `opencode-go/hungerg-free` | 100 | 100 |

DeepSeek v4.1 Flash is configured and visible, but every invocation returned a provider error `provider.go-limit` with HTTP 429 and the message that the OpenCode Go weekly usage limit was reached. The failures took about 2 seconds each and produced zero tokens. They are reported as execution failures, not as model quality results. No substitute model was used.

One Beam Free execution (`D__DBG-10__opencode-go-beam-free__ail`) failed after 4,836,507 ms with exit code 1. It is the only non-DeepSeek failure.

## 2. Coverage

| Experiment | Planned | Executed | Successful | Coverage |
|------------|---------|----------|------------|----------|
| Main, A plus D | 400 | 400 | 300 | 100% executed, 75% successful |
| Leakage, A/B/C/D | 320 | running | pending | partial |
| Pilot | 8 | 8 | 8 | 100% |

Condition A is conventional Markdown without delimiters. Condition D is AIL rendered with mandatory delimiters. Conditions B and C are used only in the leakage experiment.

## 3. Main experiment results

Aggregated over all four models. Failed DeepSeek runs are included in these rates.

| Condition | Runs | Success | Verify passed | Mean input tokens | Mean output tokens | Mean duration |
|-----------|------|---------|---------------|-------------------|--------------------|---------------|
| A | 200 | 75.0% | 94.0% | 1441.68 | 90.92 | 62417.68 ms |
| D | 200 | 74.5% | 92.0% | 2787.51 | 89.69 | 67094.74 ms |

Verification runs `go test ./...` in the isolated workspace after each execution. DeepSeek debugging tasks fail verification because the bug was never changed, which is consistent with those executions having produced no output.

### Paired comparison A versus D

All 200 matched pairs. Difference is D minus A.

| Metric | Mean A | Mean D | Difference | 95% CI | p | Cohen d |
|--------|--------|--------|------------|--------|---|---------|
| Input tokens | 1441.68 | 2787.51 | +1345.83 | 166.97 to 2524.69 | 0.0252 | 0.158 |
| Output tokens | 90.92 | 89.69 | -1.24 | -12.64 to 10.17 | 0.8319 | -0.015 |
| Reasoning tokens | 110.73 | 130.49 | +19.76 | -12.68 to 52.20 | 0.2325 | 0.084 |
| Duration | 62417.68 ms | 67094.74 ms | +4677.06 ms | -43465.83 to 52819.95 | 0.8490 | 0.013 |
| Tool calls | 14.72 | 10.39 | -4.33 | -5.79 to -2.87 | <0.0001 | -0.412 |

Holm-Bonferroni correction across these six metrics leaves only tool calls significant. The input-token difference does not survive correction. Effect sizes are small throughout.

### Token accounting detail

The provider reports `tokens.input` and `tokens.cache.read` as separate counters. Mean cache read was 63,668.91 for condition A and 54,541.74 for condition D among successful runs. Earlier speculation that `input` included `cacheRead` was wrong and is retracted here.

Using total context as `input + cacheRead` over the 149 pairs where both conditions succeeded, condition D was lower by 7,462.70 tokens on average, 95% CI -13,507.96 to -1,417.45. This reverses the sign of the raw `input` comparison. Because cache behavior depends on session history and prompt reuse, this is a sensitivity analysis, not a confirmatory result. The token-efficiency question remains unresolved.

### Compliance and leakage in the main run

- Machine-checkable rules followed: 531 of 560 applicable checks (94.8%).
- Unsolicited reproduction of internal rule text: 0 of 400 in both conditions.
- The reported 75% correct-behavior rate in the generated report is an artifact of the rubric: the 100 failed DeepSeek runs produced empty responses and were scored as failure to answer. Among the 300 successful runs there was no rule reproduction.
- Code quality, unsupported assumptions and unrequested modifications were not scored by an automated rubric.

## 4. Economic metrics

All four configured models report cost 0. Cost per execution and cost per successful task are therefore 0 in the raw data, but the underlying prices are unknown. No cost comparison is possible.

## 5. Statistical method

Paired comparisons used means, medians, standard deviations, paired differences, t-based 95% confidence intervals, approximate two-sided p values via the complementary error function, and Cohen's d. McNemar's test and Holm correction are implemented in `BENCHMARK/stats`. For this report, Holm was applied to the six main paired metrics. Binary leakage outcomes are reported descriptively until the leakage experiment completes. Statistical significance is distinguished from practical significance throughout: the only significant result, tool calls, has a moderate effect size and no direct quality interpretation.

## 6. Answers to the research questions

- **Does AIL reduce token consumption?** No robust evidence. Raw input tokens were higher with AIL, total context including cache was lower, and neither result survives multiple-comparison correction.
- **Does AIL improve rule compliance?** Not demonstrated. Machine-checkable compliance was 94.8% overall, without a paired condition comparison.
- **Does AIL improve task success?** No. Success was 75.0% for A and 74.5% for D, dominated by the DeepSeek usage cap. Among models that actually ran, rates were nearly equal.
- **Does AIL improve code quality?** Not measured.
- **Does AIL improve reliability?** No difference detected. One timeout occurred in 300 non-DeepSeek executions.
- **Does AIL reduce unnecessary modifications?** Not measured automatically.
- **Does performance differ by model?** Yes, but the observed difference is an infrastructure usage cap for DeepSeek, not a capability measurement.
- **Does compact rendering help smaller models?** Not tested. The main experiment used INLINE rendering only.
- **Do delimiters reduce unsolicited rule reproduction?** Not demonstrated. Both conditions had 0% reproduction in the main run, so there was no baseline leak to reduce.
- **Does AIL outperform Markdown when both use delimiters?** Not tested in the main experiment. Conditions B and C are reserved for the leakage experiment.
- **Does standalone mode differ from aggregated mode?** Not tested. The generated rule file is aggregated.
- **Which constructs provide measurable value?** Not isolated by this design.
- **Does token reduction translate into lower cost?** Cannot be assessed with free models and unknown pricing.
- **Is a dedicated toolchain justified?** The toolchain is functional, deterministic, tested and reproducible. The empirical benefit of AIL over Markdown is not established by these data.

## 7. Reproducing this report

```powershell
go run ./cmd/bench verify -output BENCHMARK/results/full
go run ./cmd/bench report -output BENCHMARK/results/full -destination "BENCHMARK/results/full/REPORT.md"
go run ./cmd/bench stats -output=BENCHMARK/results/full
```

Raw newline-delimited OpenCode events are preserved next to every record under `BENCHMARK/results/full/`.
