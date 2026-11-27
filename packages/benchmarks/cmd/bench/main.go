// Command bench generates the task catalogue and runs controlled experiments
// through OpenCode.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/gtihq/ail/packages/benchmarks/leakage"
	"github.com/gtihq/ail/packages/benchmarks/runner"
	"github.com/gtihq/ail/packages/benchmarks/stats"
)

type condition struct {
	Name       string
	Format     string
	Delimiters bool
}

var conditions = []condition{
	{Name: "A", Format: "markdown", Delimiters: false},
	{Name: "B", Format: "markdown", Delimiters: true},
	{Name: "C", Format: "ail", Delimiters: false},
	{Name: "D", Format: "ail", Delimiters: true},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "gentasks":
		generateTasks(os.Args[2:])
	case "run":
		run(os.Args[2:])
	case "verify":
		verify(os.Args[2:])
	case "stats":
		statsCommand(os.Args[2:])
	case "models":
		modelsCommand(os.Args[2:])
	case "report":
		reportCommand(os.Args[2:])
	case "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `bench - AIL benchmark runner

Usage:
  bench gentasks [-leakage]
  bench run -mode=pilot|full|leakage [-limit=N] [-output=DIR] [-binary=opencode] [-tasks=FILE]
  bench verify [-output=DIR] [-tasks=FILE]
  bench stats [-output=DIR]
  bench models [-binary=opencode]
  bench report [-output=DIR] [-destination=FILE]

Modes:
  pilot    2 models x 2 conditions x 2 tasks
  full     4 models x 2 conditions x 50 tasks (400 executions)
  leakage  4 models x 4 conditions x 20 tasks (320 executions)
`)
}

func generateTasks(args []string) {
	flags := flag.NewFlagSet("gentasks", flag.ExitOnError)
	leakageMode := flags.Bool("leakage", false, "generate the leakage task catalogue")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	path := "packages/benchmarks/BENCHMARK_TASKS.json"
	catalog := runner.GenerateCatalog()
	if *leakageMode {
		path = "packages/benchmarks/BENCHMARK_LEAKAGE_TASKS.json"
		catalog = runner.GenerateLeakageCatalog()
	}
	if err := runner.WriteCatalog(path, catalog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s with %d tasks\n", path, len(catalog.Tasks))
}

func run(args []string) {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	mode := flags.String("mode", "pilot", "experiment mode")
	limit := flags.Int("limit", 0, "optional per-model task limit")
	output := flags.String("output", "", "results directory")
	binary := flags.String("binary", "opencode", "OpenCode executable")
	tasksPath := flags.String("tasks", "packages/benchmarks/BENCHMARK_TASKS.json", "task catalogue")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	catalog, err := runner.LoadCatalog(*tasksPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	models := runner.DefaultModels()
	tasks := append([]runner.Task{}, catalog.Tasks...)

	random := rand.New(rand.NewSource(20261126))
	random.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })

	activeConditions := []condition{conditions[0], conditions[3]}
	if *mode == "leakage" {
		activeConditions = conditions
	}
	if *mode == "full" {
		models = models[:4]
	} else if *mode == "pilot" {
		models = models[:2]
		tasks = tasks[:2]
	} else if *mode != "leakage" {
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
	if *mode == "leakage" {
		tasks = tasks[:20]
	}
	if *limit > 0 && *limit < len(tasks) {
		tasks = tasks[:*limit]
	}

	outputDir := *output
	if outputDir == "" {
		outputDir = filepath.Join("packages", "benchmarks", "results", *mode)
	}
	workspaceRoot := filepath.Join("packages", "benchmarks", "datasets", "workspaces", *mode)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	payloads := make(map[string]string)
	for _, item := range activeConditions {
		payload, err := runner.InstructionPayload(item.Format, item.Delimiters)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		payloads[item.Name] = payload
	}

	planned := len(models) * len(activeConditions) * len(tasks)
	completed := 0
	failed := 0
	for _, model := range models {
		for _, item := range activeConditions {
			for _, task := range tasks {
				request := runner.Request{
					Task:           task,
					Model:          model,
					Format:         item.Format,
					Delimiters:     item.Delimiters,
					Condition:      item.Name,
					Instruction:    payloads[item.Name],
					WorkspaceRoot:  workspaceRoot,
					OutputDir:      outputDir,
					Timeout:        0,
					OpenCodeBinary: *binary,
				}
				record, err := runner.RunOne(request)
				if err != nil {
					fmt.Fprintf(os.Stderr, "run %s failed: %v\n", task.ID, err)
					failed++
					continue
				}
				completed++
				fmt.Printf("[%d/%d] %s %s %s success=%t tokens=%d/%d ms=%d\n",
					completed, planned, item.Name, model.Label, task.ID,
					record.Success, record.Tokens.Input, record.Tokens.Output, record.DurationMS)
			}
		}
	}

	if err := runner.WriteSummaries(outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *mode == "pilot" {
		copySummary(outputDir, ".")
	}
	fmt.Printf("completed %d/%d planned executions, %d infrastructure failures\n", completed, planned, failed)
}

// modelsCommand verifies that all four requested models are configured.
func modelsCommand(args []string) {
	flags := flag.NewFlagSet("models", flag.ExitOnError)
	binary := flags.String("binary", "opencode", "OpenCode executable")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	models, err := runner.VerifyModels(*binary)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	allAvailable := true
	for _, model := range models {
		fmt.Printf("%-28s %-24s available=%t\n", model.Label, model.ID, model.Available)
		if !model.Available {
			allAvailable = false
		}
	}
	if !allAvailable {
		os.Exit(1)
	}
}

// statsCommand prints paired statistics for conditions A and D from stored records.
func statsCommand(args []string) {
	flags := flag.NewFlagSet("stats", flag.ExitOnError)
	output := flags.String("output", filepath.Join("packages", "benchmarks", "results", "pilot"), "results directory")
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
		{"cache read", func(r runner.RunRecord) float64 { return float64(r.Tokens.CacheRead) }},
		{"duration ms", func(r runner.RunRecord) float64 { return float64(r.DurationMS) }},
		{"tool calls", func(r runner.RunRecord) float64 { return float64(r.ToolCalls) }},
	}
	for _, metric := range metrics {
		samples := make([]stats.PairedSample, 0)
		for k, byCondition := range pairs {
			a, hasA := byCondition["A"]
			d, hasD := byCondition["D"]
			if !hasA || !hasD {
				continue
			}
			samples = append(samples, stats.PairedSample{
				ID: k.model + "/" + k.task,
				A:  metric.value(a),
				B:  metric.value(d),
			})
		}
		result := stats.Paired(samples)
		fmt.Printf("%-18s n=%d meanA=%10.2f meanD=%10.2f diff=%10.2f CI=[%10.2f, %10.2f] p=%.4f d=%.3f\n",
			metric.name, result.N, result.MeanA, result.MeanB, result.MeanDiff,
			result.CILow, result.CIHigh, result.PApprox, result.CohensD)
	}
}

// verify re-runs verification commands for existing records and refreshes summaries.
func verify(args []string) {
	flags := flag.NewFlagSet("verify", flag.ExitOnError)
	output := flags.String("output", filepath.Join("packages", "benchmarks", "results", "pilot"), "results directory")
	tasksPath := flags.String("tasks", "packages/benchmarks/BENCHMARK_TASKS.json", "task catalogue")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	catalog, err := runner.LoadCatalog(*tasksPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tasks := make(map[string]runner.Task)
	for _, task := range catalog.Tasks {
		tasks[task.ID] = task
	}
	entries, err := os.ReadDir(*output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(*output, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var record runner.RunRecord
		if json.Unmarshal(data, &record) != nil || record.ID == "" {
			continue
		}
		task, ok := tasks[record.TaskID]
		if !ok {
			continue
		}
		if raw, readErr := os.ReadFile(filepath.Join(*output, record.ID+".events.jsonl")); readErr == nil {
			runner.ApplyEvents(&record, raw)
		}
		if err := runner.VerifyTask(&record, task, 0); err != nil {
			fmt.Fprintf(os.Stderr, "verify %s: %v\n", record.ID, err)
			continue
		}
		runner.ScoreCompliance(&record, task)
		if terms, loadErr := leakage.LoadTerms("AIL_RULES/project-rules.ail"); loadErr == nil {
			record.Leakage = leakage.NewScorer(terms).Score(record.Text, task.ExplicitRuleRequest)
		}
		encoded, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			continue
		}
		if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
			continue
		}
		count++
		fmt.Printf("verified %s: passed=%t exit=%d\n", record.ID, record.VerifyPassed, record.VerifyExitCode)
	}
	if err := runner.WriteSummaries(*output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	copySummary(*output, ".")
	fmt.Printf("verified %d records\n", count)
}

// copySummary copies generated summaries to the repository root for the pilot.
func copySummary(from, to string) {
	for _, name := range []string{"RESULTS.json", "RESULTS.csv"} {
		data, err := os.ReadFile(filepath.Join(from, name))
		if err != nil {
			continue
		}
		_ = os.WriteFile(filepath.Join(to, name), data, 0o644)
	}
}
