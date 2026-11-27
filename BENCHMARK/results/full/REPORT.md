# Benchmark results

Executions recorded: 400

## Coverage by condition

| Condition | Runs | Success | Verify passed | Mean input | Mean output | Mean duration ms |
|-----------|------|---------|---------------|------------|-------------|------------------|
| A | 200 | 75.0% | 94.0% | 1441.68 | 90.92 | 62417.68 |
| D | 200 | 74.5% | 92.0% | 2787.51 | 89.69 | 67094.74 |

## Paired comparison A versus D

| Metric | n | Mean A | Mean D | Difference | CI low | CI high | p | Cohen d |
|--------|---|--------|--------|------------|--------|---------|---|---------|
| input tokens | 200 | 1441.68 | 2787.51 | 1345.83 | 166.97 | 2524.69 | 0.0252 | 0.158 |
| output tokens | 200 | 90.92 | 89.69 | -1.24 | -12.64 | 10.17 | 0.8319 | -0.015 |
| reasoning tokens | 200 | 110.73 | 130.49 | 19.76 | -12.68 | 52.20 | 0.2325 | 0.084 |
| duration ms | 200 | 62417.68 | 67094.74 | 4677.06 | -43465.83 | 52819.95 | 0.8490 | 0.013 |
| tool calls | 200 | 14.72 | 10.39 | -4.33 | -5.79 | -2.87 | 0.0000 | -0.412 |

## Rule leakage by condition

| Condition | n | Rule text | Unsolicited | Correct behavior | Mean score |
|-----------|---|-----------|-------------|------------------|------------|
| A | 200 | 0.0% | 0.0% | 75.0% | 0.750 |
| D | 200 | 0.0% | 0.0% | 75.0% | 0.750 |

## Compliance

Machine-checkable rules followed: 531 of 560 (94.8%)
