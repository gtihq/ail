# Benchmark results

Executions recorded: 8

## Coverage by condition

| Condition | Runs | Success | Verify passed | Mean input | Mean output | Mean duration ms |
|-----------|------|---------|---------------|------------|-------------|------------------|
| A | 4 | 100.0% | 100.0% | 1226.25 | 127.00 | 82161.75 |
| D | 4 | 100.0% | 100.0% | 375.50 | 182.75 | 93848.00 |

## Paired comparison A versus D

| Metric | n | Mean A | Mean D | Difference | CI low | CI high | p | Cohen d |
|--------|---|--------|--------|------------|--------|---------|---|---------|
| input tokens | 4 | 1226.25 | 375.50 | -850.75 | -3758.64 | 2057.14 | 0.3519 | -0.465 |
| output tokens | 4 | 127.00 | 182.75 | 55.75 | -271.16 | 382.66 | 0.5874 | 0.271 |
| reasoning tokens | 4 | 91.25 | 187.00 | 95.75 | -319.21 | 510.71 | 0.4628 | 0.367 |
| duration ms | 4 | 82161.75 | 93848.00 | 11686.25 | -152373.26 | 175745.76 | 0.8207 | 0.113 |
| tool calls | 4 | 10.50 | 9.75 | -0.75 | -7.91 | 6.41 | 0.7389 | -0.167 |

## Rule leakage by condition

| Condition | n | Rule text | Unsolicited | Correct behavior | Mean score |
|-----------|---|-----------|-------------|------------------|------------|
| A | 4 | 0.0% | 0.0% | 100.0% | 1.000 |
| D | 4 | 0.0% | 0.0% | 100.0% | 1.000 |

## Compliance

Machine-checkable rules followed: 4 of 4 (100.0%)
