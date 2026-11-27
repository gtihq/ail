# Benchmark results

Executions recorded: 400

## Coverage by condition

| Condition | Runs | Success | Verify passed | Mean input | Mean output | Mean duration ms |
|-----------|------|---------|---------------|------------|-------------|------------------|
| A | 200 | 75.0% | 94.0% | 55761.21 | 1269.92 | 62417.68 |
| D | 200 | 74.5% | 92.0% | 45646.57 | 770.68 | 67094.74 |

## Paired comparison A versus D

| Metric | n | Mean A | Mean D | Difference | CI low | CI high | p | Cohen d |
|--------|---|--------|--------|------------|--------|---------|---|---------|
| input tokens | 200 | 55761.21 | 45646.57 | -10114.65 | -17519.40 | -2709.89 | 0.0074 | -0.189 |
| output tokens | 200 | 1269.92 | 770.68 | -499.24 | -629.92 | -368.56 | 0.0000 | -0.529 |
| reasoning tokens | 200 | 2044.03 | 1392.96 | -651.07 | -899.25 | -402.89 | 0.0000 | -0.364 |
| duration ms | 200 | 62417.68 | 67094.74 | 4677.06 | -43465.83 | 52819.95 | 0.8490 | 0.013 |
| tool calls | 200 | 14.72 | 10.39 | -4.33 | -5.79 | -2.87 | 0.0000 | -0.412 |

## Rule leakage by condition

| Condition | n | Rule text | Unsolicited | Correct behavior | Mean score |
|-----------|---|-----------|-------------|------------------|------------|
| A | 200 | 0.0% | 0.0% | 75.0% | 0.750 |
| D | 200 | 0.0% | 0.0% | 75.0% | 0.750 |

## Compliance

Machine-checkable rules followed: 531 of 560 (94.8%)
