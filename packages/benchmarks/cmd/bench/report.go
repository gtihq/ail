package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gtihq/ail/packages/benchmarks/leakage"
	"github.com/gtihq/ail/packages/benchmarks/runner"
	"github.com/gtihq/ail/packages/benchmarks/stats"
)

// reportCommand generates a data-driven report from stored execution records.
func reportCommand(args []string) {
	flags := flag.NewFlagSet("report", flag.ExitOnError)
	output := flags.String("output", filepath.Join("packages", "benchmarks", "results", "full"), "results directory")
	destination := flags.String("destination", "", "report destination, defaults to REPORT.md in results directory")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	data, err := os.ReadFile(filepath.Join(*output, "RESULTS.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var records []runner.RunRecord
	if err := json.Unmarshal(data, &records); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dest := *destination
	if dest == "" {
		dest = filepath.Join(*output, "REPORT.md")
	}
	var out strings.Builder
	out.WriteString("# Benchmark results\n\n")
	fmt.Fprintf(&out, "Executions recorded: %d\n\n", len(records))

	out.WriteString("## Coverage by condition\n\n")
	out.WriteString("| Condition | Runs | Success | Verify passed | Mean input | Mean output | Mean duration ms |\n")
	out.WriteString("|-----------|------|---------|---------------|------------|-------------|------------------|\n")
	conditions := []string{"A", "B", "C", "D"}
	for _, condition := range conditions {
		group := filterRecords(records, func(record runner.RunRecord) bool { return record.Condition == condition })
		if len(group) == 0 {
			continue
		}
		var input, outputTokens, duration float64
		success, verified := 0, 0
		for _, record := range group {
			input += float64(record.Tokens.Input)
			outputTokens += float64(record.Tokens.Output)
			duration += float64(record.DurationMS)
			if record.Success {
				success++
			}
			if record.VerifyPassed {
				verified++
			}
		}
		n := float64(len(group))
		fmt.Fprintf(&out, "| %s | %d | %.1f%% | %.1f%% | %.2f | %.2f | %.2f |\n",
			condition, len(group), float64(success)/n*100, float64(verified)/n*100,
			input/n, outputTokens/n, duration/n)
	}
	out.WriteString("\n")

	out.WriteString("## Paired comparison A versus D\n\n")
	type key struct{ model, task string }
	pairs := make(map[key]map[string]runner.RunRecord)
	for _, record := range records {
		k := key{model: record.Model, task: record.TaskID}
		if pairs[k] == nil {
			pairs[k] = make(map[string]runner.RunRecord)
		}
		pairs[k][record.Condition] = record
	}
	metrics := []struct {
		name  string
		value func(runner.RunRecord) float64
	}{
		{"input tokens", func(r runner.RunRecord) float64 { return float64(r.Tokens.Input) }},
		{"output tokens", func(r runner.RunRecord) float64 { return float64(r.Tokens.Output) }},
		{"reasoning tokens", func(r runner.RunRecord) float64 { return float64(r.Tokens.Reasoning) }},
		{"duration ms", func(r runner.RunRecord) float64 { return float64(r.DurationMS) }},
		{"tool calls", func(r runner.RunRecord) float64 { return float64(r.ToolCalls) }},
	}
	out.WriteString("| Metric | n | Mean A | Mean D | Difference | CI low | CI high | p | Cohen d |\n")
	out.WriteString("|--------|---|--------|--------|------------|--------|---------|---|---------|\n")
	for _, metric := range metrics {
		samples := make([]stats.PairedSample, 0)
		for k, byCondition := range pairs {
			a, hasA := byCondition["A"]
			d, hasD := byCondition["D"]
			if !hasA || !hasD {
				continue
			}
			samples = append(samples, stats.PairedSample{ID: k.model + "/" + k.task, A: metric.value(a), B: metric.value(d)})
		}
		result := stats.Paired(samples)
		fmt.Fprintf(&out, "| %s | %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.4f | %.3f |\n",
			metric.name, result.N, result.MeanA, result.MeanB, result.MeanDiff,
			result.CILow, result.CIHigh, result.PApprox, result.CohensD)
	}
	out.WriteString("\n")

	out.WriteString("## Rule leakage by condition\n\n")
	out.WriteString("| Condition | n | Rule text | Unsolicited | Correct behavior | Mean score |\n")
	out.WriteString("|-----------|---|-----------|-------------|------------------|------------|\n")
	for _, condition := range conditions {
		group := filterRecords(records, func(record runner.RunRecord) bool { return record.Condition == condition })
		if len(group) == 0 {
			continue
		}
		verdicts := make([]leakage.Verdict, 0, len(group))
		for _, record := range group {
			verdicts = append(verdicts, record.Leakage)
		}
		result := leakage.Aggregate(verdicts)
		fmt.Fprintf(&out, "| %s | %d | %.1f%% | %.1f%% | %.1f%% | %.3f |\n",
			condition, result.Count, result.RuleTextRate, result.UnsolicitedRate,
			result.CorrectBehaviorRate, result.MeanScore)
	}
	out.WriteString("\n")

	out.WriteString("## Compliance\n\n")
	var checked, followed int
	for _, record := range records {
		checked += record.Compliance.Checked
		followed += record.Compliance.Followed
	}
	if checked > 0 {
		fmt.Fprintf(&out, "Machine-checkable rules followed: %d of %d (%.1f%%)\n", followed, checked, float64(followed)/float64(checked)*100)
	} else {
		out.WriteString("No machine-checkable rule applied to this result set.\n")
	}

	if err := os.WriteFile(dest, []byte(out.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", dest)
}

func filterRecords(records []runner.RunRecord, predicate func(runner.RunRecord) bool) []runner.RunRecord {
	result := make([]runner.RunRecord, 0)
	for _, record := range records {
		if predicate(record) {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
