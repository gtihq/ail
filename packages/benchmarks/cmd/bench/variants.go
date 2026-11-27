package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gtihq/ail/packages/benchmarks/runner"
	"github.com/gtihq/ail/packages/core"
)

// variantHypotheses are the AIL payload shapes tested against the same tasks.
var variantHypotheses = []core.FormatVariant{
	core.VariantFull,
	core.VariantNoInterpretation,
	core.VariantNoDelimiters,
	core.VariantCompactBare,
	core.VariantTinyRules,
}

// variantsCommand runs a small paired experiment over AIL format variants using one model.
func variantsCommand(args []string) {
	flags := flag.NewFlagSet("variants", flag.ExitOnError)
	perVariant := flags.Int("n", 5, "tasks per variant")
	output := flags.String("output", filepath.Join("packages", "benchmarks", "results", "variants"), "results directory")
	modelFilter := flags.String("model", "hy4-flash-free", "model id substring")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	catalog, err := runner.LoadCatalog("packages/benchmarks/BENCHMARK_TASKS.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var model runner.ModelSpec
	for _, candidate := range runner.DefaultModels() {
		if strings.Contains(candidate.ID, *modelFilter) {
			model = candidate
			break
		}
	}
	if model.ID == "" {
		fmt.Fprintf(os.Stderr, "no model matches %q\n", *modelFilter)
		os.Exit(1)
	}

	tasks := catalog.Tasks
	if *perVariant < len(tasks) {
		tasks = tasks[:*perVariant]
	}
	workspaceRoot := filepath.Join("packages", "benchmarks", "datasets", "workspaces", "variants")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, variant := range variantHypotheses {
		instruction, err := runner.InstructionPayloadVariant(variant)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		chars := len(instruction)
		words := len(strings.Fields(instruction))
		fmt.Printf("VARIANT %s chars=%d words=%d\n", variant, chars, words)
		for _, task := range tasks {
			request := runner.Request{
				Task:           task,
				Model:          model,
				Format:         "ail",
				Delimiters:     true,
				Condition:      string(variant),
				Instruction:    instruction,
				WorkspaceRoot:  workspaceRoot,
				OutputDir:      *output,
				OpenCodeBinary: "opencode",
				Variant:        variant,
			}
			record, err := runner.RunOne(request)
			if err != nil {
				fmt.Fprintf(os.Stderr, "run %s failed: %v\n", task.ID, err)
				continue
			}
			fmt.Printf("[%s] %s success=%t input=%d output=%d tools=%d ms=%d\n",
				variant, task.ID, record.Success, record.Tokens.Input,
				record.Tokens.Output, record.ToolCalls, record.DurationMS)
		}
	}
	if err := runner.WriteSummaries(*output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	writeVariantReport(*output)
}

// writeVariantReport produces a compact JSON summary by variant.
func writeVariantReport(output string) {
	data, err := os.ReadFile(filepath.Join(output, "RESULTS.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var records []runner.RunRecord
	if err := json.Unmarshal(data, &records); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	summary := make(map[string]map[string]float64)
	for _, record := range records {
		key := record.Condition
		if summary[key] == nil {
			summary[key] = map[string]float64{"n": 0, "success": 0, "input": 0, "output": 0, "tools": 0, "duration": 0}
		}
		entry := summary[key]
		entry["n"]++
		if record.Success {
			entry["success"]++
		}
		entry["input"] += float64(record.Tokens.Input)
		entry["output"] += float64(record.Tokens.Output)
		entry["tools"] += float64(record.ToolCalls)
		entry["duration"] += float64(record.DurationMS)
	}
	for key, entry := range summary {
		n := entry["n"]
		if n == 0 {
			continue
		}
		fmt.Printf("SUMMARY %s n=%.0f success=%.0f input=%.1f output=%.1f tools=%.2f duration=%.1f\n",
			key, n, entry["success"], entry["input"]/n, entry["output"]/n,
			entry["tools"]/n, entry["duration"]/n)
	}
}
