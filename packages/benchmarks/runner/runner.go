// Package runner executes benchmark tasks through OpenCode in isolated
// workspaces and records token usage, tool interactions and timing.
package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gtihq/ail/packages/benchmarks/leakage"
	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/parser"
	"github.com/gtihq/ail/packages/renderer"
)

// Task is one benchmark unit with an inline repository baseline.
type Task struct {
	ID              string            `json:"id"`
	Category        string            `json:"category"`
	Prompt          string            `json:"prompt"`
	Files           map[string]string `json:"files"`
	ApplicableRules   []string          `json:"applicableRules"`
	VerifyCommand     string            `json:"verifyCommand"`
	Scenario          string            `json:"scenario,omitempty"`
	ExplicitRuleRequest bool            `json:"explicitRuleRequest,omitempty"`
}

// Catalog is a versioned list of benchmark tasks.
type Catalog struct {
	Version string `json:"version"`
	Tasks   []Task `json:"tasks"`
}

// ModelSpec is an exact OpenCode model identifier with a human label.
type ModelSpec struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// DefaultModels returns the four requested models with their verified identifiers.
func DefaultModels() []ModelSpec {
	return []ModelSpec{
		{ID: "opencode-go/hy4-flash-free", Label: "Hy4 Flash Free"},
		{ID: "opencode-go/beam-free", Label: "Beam Free"},
		{ID: "opencode-go/deepseek-v4.1-flash", Label: "DeepSeek v4.1 Flash"},
		{ID: "opencode-go/hungerg-free", Label: "HungerG Free"},
	}
}

// ModelAvailability records whether a requested model is present in OpenCode.
type ModelAvailability struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Available bool   `json:"available"`
}

// VerifyModels checks the four requested model identifiers against `opencode models`.
func VerifyModels(binary string) ([]ModelAvailability, error) {
	if binary == "" {
		binary = "opencode"
	}
	output, err := exec.Command(binary, "models").Output()
	if err != nil {
		return nil, err
	}
	available := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		available[strings.TrimSpace(line)] = true
	}
	result := make([]ModelAvailability, 0, 4)
	for _, model := range DefaultModels() {
		result = append(result, ModelAvailability{
			ID:        model.ID,
			Label:     model.Label,
			Available: available[model.ID],
		})
	}
	return result, nil
}

// Tokens aggregates provider-reported token usage for one execution.
type Tokens struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	Reasoning int `json:"reasoning"`
	CacheRead int `json:"cacheRead"`
	CacheWrite int `json:"cacheWrite"`
}

// Compliance records machine-checkable rule adherence for one execution.
type Compliance struct {
	Applicable int      `json:"applicable"`
	Checked    int      `json:"checked"`
	Followed   int      `json:"followed"`
	Violated   []string `json:"violated"`
	Pct        float64  `json:"pct"`
}

// RunRecord is the persisted result of one execution.
type RunRecord struct {
	ID           string    `json:"id"`
	TaskID       string    `json:"taskId"`
	Category     string    `json:"category"`
	Model        string    `json:"model"`
	Format       string    `json:"format"`
	Delimiters   bool      `json:"delimiters"`
	Condition    string    `json:"condition"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
	DurationMS   int64     `json:"durationMs"`
	ExitCode     int       `json:"exitCode"`
	Tokens       Tokens    `json:"tokens"`
	Cost         float64   `json:"cost"`
	Text         string    `json:"text"`
	EventCount   int       `json:"eventCount"`
	ToolCalls    int       `json:"toolCalls"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	SessionID    string    `json:"sessionId,omitempty"`
	WorkDir      string    `json:"workDir"`
	VerifyPassed bool      `json:"verifyPassed"`
	VerifyExitCode int     `json:"verifyExitCode"`
	VerifyOutput string   `json:"verifyOutput,omitempty"`
	Compliance   Compliance `json:"compliance"`
	Leakage      leakage.Verdict `json:"leakage"`
}

// Request describes a single OpenCode invocation.
type Request struct {
	Task            Task
	Model           ModelSpec
	Format          string
	Delimiters      bool
	Condition       string
	Instruction     string
	WorkspaceRoot   string
	OutputDir       string
	Timeout         time.Duration
	OpenCodeBinary  string
}

// LoadCatalog reads a task catalogue from disk.
func LoadCatalog(path string) (Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, err
	}
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

// WriteCatalog writes a task catalogue to disk.
func WriteCatalog(path string, catalog Catalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// InstructionPayload builds the instruction context for a format and delimiter mode.
func InstructionPayload(format string, delimiters bool) (string, error) {
	source, err := os.ReadFile(".ail/rules/project-rules.ail")
	if err != nil {
		return "", err
	}
	parsed := parser.Parse(string(source), "project-rules.ail")
	ir := interpreter.Interpret(parsed).IR

	if strings.EqualFold(format, "markdown") {
		baseline, err := os.ReadFile("packages/benchmarks/MARKDOWN_BASELINE.md")
		if err != nil {
			return "", err
		}
		if !delimiters {
			return string(baseline), nil
		}
		return renderer.InterpretationInstruction() + "\n\n" + renderer.BeginMarker() + "\n\n" +
			string(baseline) + "\n" + renderer.EndMarker() + "\n", nil
	}

	include := delimiters
	result := renderer.Render(ir, core.RenderOptions{
		Host:              core.HostOpenCode,
		Mode:              core.ModeInline,
		IncludeDelimiters: &include,
	})
	return result.Text, nil
}

type event struct {
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	SessionID string          `json:"sessionID"`
	Part      json.RawMessage `json:"part"`
}

type finishPart struct {
	Cost   float64 `json:"cost"`
	Tokens struct {
		Input  int `json:"input"`
		Output int `json:"output"`
		Reasoning int `json:"reasoning"`
		Cache struct {
			Read  int `json:"read"`
			Write int `json:"write"`
		} `json:"cache"`
	} `json:"tokens"`
}

func parseEvents(raw []byte) (text string, tokens Tokens, cost float64, sessionID string, toolCalls, eventCount int) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	var builder strings.Builder
	var accumulated Tokens
	var accumulatedCost float64
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var item event
		if err := json.Unmarshal(line, &item); err != nil {
			continue
		}
		eventCount++
		if item.SessionID != "" {
			sessionID = item.SessionID
		}
		var part map[string]json.RawMessage
		if err := json.Unmarshal(item.Part, &part); err != nil {
			continue
		}
		if typeRaw, ok := part["type"]; ok {
			var partType string
			_ = json.Unmarshal(typeRaw, &partType)
			if strings.HasPrefix(partType, "tool") {
				toolCalls++
			}
		}
		if item.Type == "text" {
			if textRaw, ok := part["text"]; ok {
				var chunk string
				if json.Unmarshal(textRaw, &chunk) == nil {
					builder.WriteString(chunk)
				}
			}
		}
		if item.Type == "step_finish" {
			var finish finishPart
			if json.Unmarshal(item.Part, &finish) == nil {
				accumulated.Input += finish.Tokens.Input
				accumulated.Output += finish.Tokens.Output
				accumulated.Reasoning += finish.Tokens.Reasoning
				accumulated.CacheRead += finish.Tokens.Cache.Read
				accumulated.CacheWrite += finish.Tokens.Cache.Write
				accumulatedCost += finish.Cost
			}
		}
	}
	return builder.String(), accumulated, accumulatedCost, sessionID, toolCalls, eventCount
}

// ApplyEvents reparses a raw event stream into a record. Used to recompute
// token aggregates after parser fixes without re-running model calls.
func ApplyEvents(record *RunRecord, raw []byte) {
	text, tokens, cost, sessionID, toolCalls, eventCount := parseEvents(raw)
	record.Text = text
	record.Tokens = tokens
	record.Cost = cost
	record.SessionID = sessionID
	record.ToolCalls = toolCalls
	record.EventCount = eventCount
	record.Success = text != "" && record.ExitCode == 0
}

// VerifyTask runs the task verification command inside the isolated workspace.
func VerifyTask(record *RunRecord, task Task, timeout time.Duration) error {
	if task.VerifyCommand == "" || record.WorkDir == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	parts := strings.Fields(task.VerifyCommand)
	if len(parts) == 0 {
		return nil
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = record.WorkDir
	output, err := cmd.CombinedOutput()
	record.VerifyOutput = string(output)
	if exitError, ok := err.(*exec.ExitError); ok {
		record.VerifyExitCode = exitError.ExitCode()
	} else if err != nil {
		record.VerifyExitCode = -1
	}
	record.VerifyPassed = err == nil
	return nil
}

// ScoreCompliance evaluates the machine-checkable applicable rules in a workspace.
func ScoreCompliance(record *RunRecord, task Task) {
	result := Compliance{Applicable: len(task.ApplicableRules), Violated: []string{}}
	machineRules := map[string]bool{"R001": true, "R002": true, "R003": true}
	for _, rule := range task.ApplicableRules {
		if !machineRules[rule] {
			continue
		}
		result.Checked++
		passed := false
		switch rule {
		case "R001":
			passed = !workspaceContainsDash(record.WorkDir)
		case "R002":
			passed = !workspaceHasLargeFile(record.WorkDir)
		case "R003":
			passed = record.VerifyPassed
		}
		if passed {
			result.Followed++
		} else {
			result.Violated = append(result.Violated, rule)
		}
	}
	if result.Checked > 0 {
		result.Pct = float64(result.Followed) / float64(result.Checked) * 100
	}
	record.Compliance = result
}

func workspaceContainsDash(root string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) > 1<<20 {
			return nil
		}
		if strings.Contains(string(data), "\u2014") {
			found = true
		}
		return nil
	})
	return found
}

func workspaceHasLargeFile(root string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) > 1<<20 {
			return nil
		}
		if strings.Count(string(data), "\n")+1 > 700 {
			found = true
		}
		return nil
	})
	return found
}

// prepareWorkspace creates an isolated directory and writes the task baseline into it.
func prepareWorkspace(root, id string, task Task) (string, error) {
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	names := make([]string, 0, len(task.Files))
	for name := range task.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(task.Files[name]), 0o644); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// RunOne executes a single task and persists its record. Completed successful runs are skipped.
func RunOne(request Request) (RunRecord, error) {
	if err := os.MkdirAll(request.OutputDir, 0o755); err != nil {
		return RunRecord{}, err
	}
	id := fmt.Sprintf("%s__%s__%s__%s", request.Condition, request.Task.ID, sanitize(request.Model.ID), request.Format)
	recordPath := filepath.Join(request.OutputDir, id+".json")
	if existing, err := os.ReadFile(recordPath); err == nil {
		var record RunRecord
		if json.Unmarshal(existing, &record) == nil && record.Success {
			return record, nil
		}
	}

	workDir, err := prepareWorkspace(request.WorkspaceRoot, id, request.Task)
	if err != nil {
		return RunRecord{}, err
	}
	binary := request.OpenCodeBinary
	if binary == "" {
		binary = "opencode"
	}
	prompt := request.Instruction + "\n\n--- TASK ---\n\n" + request.Task.Prompt
	promptPath := filepath.Join(workDir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return RunRecord{}, err
	}

	timeout := request.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{
		"run",
		"--model", request.Model.ID,
		"--format", "json",
		"--auto",
		"--archived",
		"--title", "ail-bench-"+id,
		"--file", promptPath,
		"Execute the task described in the attached file.",
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	record := RunRecord{
		ID:         id,
		TaskID:     request.Task.ID,
		Category:   request.Task.Category,
		Model:      request.Model.ID,
		Format:     request.Format,
		Delimiters: request.Delimiters,
		Condition:  request.Condition,
		StartedAt:  time.Now().UTC(),
		WorkDir:    workDir,
	}
	runErr := cmd.Run()
	record.FinishedAt = time.Now().UTC()
	record.DurationMS = record.FinishedAt.Sub(record.StartedAt).Milliseconds()
	if runErr != nil {
		record.Error = runErr.Error()
		if stderr.Len() > 0 {
			record.Error += ": " + stderr.String()
		}
	}
	if exitError, ok := runErr.(*exec.ExitError); ok {
		record.ExitCode = exitError.ExitCode()
	}

	ApplyEvents(&record, stdout.Bytes())
	if runErr != nil {
		record.Success = false
	}
	if terms, loadErr := leakage.LoadTerms(".ail/rules/project-rules.ail"); loadErr == nil {
		record.Leakage = leakage.NewScorer(terms).Score(record.Text, request.Task.ExplicitRuleRequest)
	}
	if err := VerifyTask(&record, request.Task, 0); err != nil {
		record.VerifyOutput = fmt.Sprintf("verification error: %v", err)
	}
	ScoreCompliance(&record, request.Task)

	rawPath := filepath.Join(request.OutputDir, id+".events.jsonl")
	if err := os.WriteFile(rawPath, stdout.Bytes(), 0o644); err != nil {
		return record, err
	}
	encoded, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return record, err
	}
	if err := os.WriteFile(recordPath, append(encoded, '\n'), 0o644); err != nil {
		return record, err
	}
	return record, nil
}

func sanitize(value string) string {
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, ".", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

// WriteSummaries writes RESULTS.json and RESULTS.csv from all records in a directory.
func WriteSummaries(outputDir string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	records := make([]RunRecord, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".events.jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(outputDir, entry.Name()))
		if err != nil {
			continue
		}
		var record RunRecord
		if json.Unmarshal(data, &record) == nil && record.ID != "" {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })

	encoded, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "RESULTS.json"), append(encoded, '\n'), 0o644); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(outputDir, "RESULTS.csv"))
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	header := []string{"id", "taskId", "category", "model", "format", "condition", "success", "verifyPassed", "input", "output", "reasoning", "cacheRead", "cacheWrite", "cost", "durationMs", "toolCalls", "eventCount", "checked", "followed", "violated", "compliancePct", "leakRuleText", "leakUnsolicited", "leakScore"}
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, record := range records {
		row := []string{
			record.ID, record.TaskID, record.Category, record.Model, record.Format, record.Condition,
			fmt.Sprintf("%t", record.Success), fmt.Sprintf("%t", record.VerifyPassed),
			fmt.Sprintf("%d", record.Tokens.Input), fmt.Sprintf("%d", record.Tokens.Output),
			fmt.Sprintf("%d", record.Tokens.Reasoning), fmt.Sprintf("%d", record.Tokens.CacheRead),
			fmt.Sprintf("%d", record.Tokens.CacheWrite), fmt.Sprintf("%.6f", record.Cost),
			fmt.Sprintf("%d", record.DurationMS), fmt.Sprintf("%d", record.ToolCalls),
			fmt.Sprintf("%d", record.EventCount),
			fmt.Sprintf("%d", record.Compliance.Checked), fmt.Sprintf("%d", record.Compliance.Followed),
			strings.Join(record.Compliance.Violated, ";"), fmt.Sprintf("%.2f", record.Compliance.Pct),
			fmt.Sprintf("%t", record.Leakage.ContainsRuleText),
			fmt.Sprintf("%t", record.Leakage.UnsolicitedReproduction),
			fmt.Sprintf("%d", record.Leakage.Score),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
