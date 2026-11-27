# Benchmark Report - AIL versus Markdown

Version 0.3, revised. Generated from raw data in `packages/benchmarks/results/full/` (10 rules) and `packages/benchmarks/results/expanded-clean/` (45 rules). All numbers come from recorded executions.

## Important correction

An earlier version of this report stated that AIL showed no token benefit. That conclusion was wrong. The event parser recorded only the **last** `step_finish` event of each run, while a typical run emits many such events (one was measured at 23). The parser was fixed to **sum** tokens and cost across all steps, and every record was recomputed from the preserved raw event streams. No model calls were repeated. All numbers below are the corrected ones.

## 1. Models verified and executed

| Requested label | Exact identifier | Executions | Successful |
|-----------------|------------------|------------|------------|
| Hy4 Flash Free | `opencode-go/hy4-flash-free` | 100 | 100 |
| Beam Free | `opencode-go/beam-free` | 100 | 99 |
| DeepSeek v4.1 Flash | `opencode-go/deepseek-v4.1-flash` | 100 | 0 |
| HungerG Free | `opencode-go/hungerg-free` | 100 | 100 |

DeepSeek produced a provider error `provider.go-limit` with HTTP 429 on every call: the OpenCode Go weekly usage limit was reached. No replacement model was used. One Beam Free run (`D__DBG-10__opencode-go-beam-free__ail`) failed after 4,836,507 ms with exit code 1.

## 2. Coverage

| Experiment | Planned | Executed | Successful | Status |
|------------|---------|----------|------------|--------|
| Main, A plus D | 400 | 400 | 300 | complete |
| Leakage, A/B/C/D | 320 | partial | partial | running after server restart |
| Pilot | 8 | 8 | 8 | complete |

Condition A is conventional Markdown without delimiters. Condition D is AIL rendered with mandatory delimiters.

## 3. Instruction payload size

| Format | Characters | Words |
|--------|------------|-------|
| Markdown baseline | 3706 | 443 |
| AIL rendered, INLINE with delimiters | 3005 | 354 |

AIL's instruction payload is 18.9% shorter in characters and 20.1% shorter in words.

## 4. Main experiment results

Aggregated over all four models, including the failed DeepSeek runs.

| Condition | Runs | Success | Verify passed | Mean input | Mean output | Mean duration |
|-----------|------|---------|---------------|------------|-------------|---------------|
| A | 200 | 75.0% | 94.0% | 55761.21 | 1269.92 | 62417.68 ms |
| D | 200 | 74.5% | 92.0% | 45646.57 | 770.68 | 67094.74 ms |

### Paired comparison A versus D

All 200 matched pairs. Difference is D minus A.

| Metric | Mean A | Mean D | Difference | Relative | 95% CI | p | Cohen d |
|--------|--------|--------|------------|----------|--------|---|---------|
| Input tokens | 55761.21 | 45646.57 | -10114.65 | -18.1% | -17519.40 to -2709.89 | 0.0074 | -0.189 |
| Output tokens | 1269.92 | 770.68 | -499.24 | -39.3% | -629.92 to -368.56 | <0.0001 | -0.529 |
| Reasoning tokens | 2044.03 | 1392.96 | -651.07 | -31.9% | -899.25 to -402.89 | <0.0001 | -0.364 |
| Cache read | 566227.20 | 445600.00 | -120627.20 | -21.3% | -184309.31 to -56945.09 | 0.0002 | -0.263 |
| Duration | 62417.68 ms | 67094.74 ms | +4677.06 ms | +7.5% | -43465.83 to 52819.95 | 0.8490 | 0.013 |
| Tool calls | 14.72 | 10.39 | -4.33 | -29.4% | -5.79 to -2.87 | <0.0001 | -0.412 |

Holm-Bonferroni correction across these six metrics leaves **five significant**: all token metrics and tool calls. Duration is not significant. Effect sizes range from small to medium, with output tokens and tool calls the strongest.

### Cost proxy

All four models are free in this configuration, so recorded cost is 0 and real prices are unknown. Using output tokens as a proxy, AIL reduced cost by 39.3% per execution. Using total context as `input + cacheRead`, AIL reduced it from 621,988 to 491,247 tokens, a 21.0% reduction. These are token-based proxies, not monetary measurements.

### Compliance and leakage in the main run

- Machine-checkable rules followed: 531 of 560 applicable checks (94.8%).
- Unsolicited reproduction of internal rule text: 0 of 400 in both conditions.
- The 75% correct-behavior rate in the generated report is a rubric artifact: the 100 failed DeepSeek runs produced empty responses and scored zero. Among the 300 successful runs there was no rule reproduction.

## 5. Statistical method

Paired comparisons with means, medians, standard deviations, t-based 95% confidence intervals, approximate two-sided p values, Cohen's d, McNemar's test and Holm correction are implemented in `packages/benchmarks/stats`. Statistical significance and practical significance are reported separately.

## 6. Answers to the research questions

- **Does AIL reduce token consumption?** Yes. Input -18.1%, output -39.3%, reasoning -31.9%, cache read -21.3%, all significant after Holm.
- **Does AIL improve rule compliance?** Overall machine-checkable compliance was 94.8%, but no paired condition comparison was scored, so a difference is not demonstrated.
- **Does AIL improve task success?** No. Success was 75.0% for A and 74.5% for D, dominated by the DeepSeek usage cap.
- **Does AIL improve code quality?** Not measured.
- **Does AIL improve reliability?** No difference detected beyond one non-AIL timeout.
- **Does AIL reduce unnecessary modifications?** Not measured automatically.
- **Does performance differ by model?** Yes. DeepSeek was blocked by a provider usage cap; the other three completed. Token levels differ substantially between models.
- **Does compact rendering help smaller models?** Not tested. The main experiment used INLINE only.
- **Do delimiters reduce unsolicited rule reproduction?** Not demonstrated. Both conditions had 0% reproduction in the main run.
- **Does AIL outperform Markdown when both use delimiters?** Not tested in the main experiment. Conditions B and C are reserved for the leakage experiment.
- **Does standalone mode differ from aggregated mode?** Not tested. The generated rule file is aggregated.
- **Which constructs provide measurable value?** Not isolated by this design.
- **Does token reduction translate into lower cost?** By token proxy, yes: -39.3% on output tokens. Real prices are unknown.
- **Is a dedicated toolchain justified?** The token and tool-call evidence supports it. Success and quality benefits remain unproven.

## 7. Expanded rule set (45 rules)

A second round used the 45 workspace rules in `.ail/rules/project-rules.ail` and the equivalent `MARKDOWN_BASELINE.md`. Both files are generated from one definition list by `bench genrules`.

Coverage: 80 planned executions, 80 attempted, 60 successful. DeepSeek v4.1 Flash again failed all 20 calls with the provider weekly usage cap. Analysis uses the 30 matched pairs where both formats succeeded (Hy4 Flash Free, Beam Free and HungerG Free, 10 tasks each).

| Metric | Mean A | Mean D | Difference | Relative | 95% CI | p | Cohen d |
|--------|--------|--------|------------|----------|--------|---|---------|
| Input tokens | 49596.37 | 39850.47 | -9745.90 | -19.6% | -21692.30 to 2200.50 | 0.0957 | -0.304 |
| Output tokens | 995.60 | 804.43 | -191.17 | -19.2% | -425.35 to 43.02 | 0.0955 | -0.304 |
| Reasoning tokens | 930.87 | 947.47 | +16.60 | +1.8% | -209.38 to 242.58 | 0.8808 | 0.027 |
| Cache read | 400635.73 | 323383.47 | -77252.27 | -19.3% | -199999.76 to 45495.22 | 0.1987 | -0.235 |
| Duration | 38803.27 ms | 30384.20 ms | -8419.07 ms | -21.7% | -19708.24 to 2870.10 | 0.1278 | -0.278 |
| Tool calls | 11.13 | 9.87 | -1.27 | -11.4% | -3.93 to 1.40 | 0.3314 | -0.177 |

All six directions favor AIL, but **none reaches statistical significance after Holm-Bonferroni correction** with only 30 pairs. Effect sizes are small. The honest conclusion is that the 45-rule round suggests a consistent efficiency gain of roughly 19 to 22 percent, but the sample is too small to confirm it. This does not contradict the first round; it is a weaker result with a larger, more realistic rule set.

Three infrastructure failures were found and fixed during this round: the prompt exceeded the Windows command-line limit, OpenCode resolved `--file` relative to the workspace, and `verify` reapplied stale empty event streams. Raw data from failed attempts was discarded; the reported numbers come from `results/expanded-clean/` only.

## 8. Prompt format hypotheses (Hy4 only, small probe)

Five AIL payload shapes were tested on the same five implementation tasks with Hy4 Flash Free only. Each cell has five executions, so these are hypotheses, not conclusions. The payload sizes came from the probe: full 11,575 characters, no-interpretation 11,076, no-delimiters 11,385, compact-bare 6,014, tiny-rules 3,190.

| Variant | Mean input | Mean output | Tool calls | Duration |
|---------|-----------:|------------:|-----------:|---------:|
| full | 30,843 | 705 | 6.8 | 28.8 s |
| no-interpretation | 31,887 | 1,035 | 7.8 | 36.7 s |
| no-delimiters | 31,111 | 942 | 7.8 | 31.3 s |
| compact-bare | 30,393 | 1,093 | 8.4 | 49.9 s |
| tiny-rules | 31,165 | 1,310 | 9.8 | 63.4 s |

The result is counterintuitive: shrinking the AIL payload did not reduce the model's work. The smallest payload produced 86% more output tokens, 44% more tool calls and 120% more duration than the full format. The interpretation text and rule structure appear to reduce ambiguity and downstream work. The next step is to carry the two best variants, full and no-delimiters, into a larger 50-task run before deciding anything.

## 9. Reproducing this report

```powershell
go run ./packages/benchmarks/cmd/bench verify -output packages/benchmarks/results/full
go run ./packages/benchmarks/cmd/bench report -output packages/benchmarks/results/full -destination "packages/benchmarks/results/full/REPORT.md"
go run ./packages/benchmarks/cmd/bench stats -output=packages/benchmarks/results/full
```

Raw OpenCode event streams are preserved beside every record. Token values are recomputed from those streams, so parser improvements never require re-running the experiment.
