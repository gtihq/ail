package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain lets the compiled test binary act as a fake OpenCode executable.
func TestMain(m *testing.M) {
	if os.Getenv("AIL_FAKE_OPENCODE") == "1" && len(os.Args) > 1 && os.Args[1] == "run" {
		fmt.Println(`{"type":"step_start","timestamp":1,"sessionID":"ses_fake","part":{"type":"step-start"}}`)
		fmt.Println(`{"type":"text","timestamp":2,"sessionID":"ses_fake","part":{"type":"text","text":"fake answer"}}`)
		fmt.Println(`{"type":"tool_use","timestamp":3,"sessionID":"ses_fake","part":{"type":"tool-use"}}`)
		fmt.Println(`{"type":"step_finish","timestamp":4,"sessionID":"ses_fake","part":{"type":"step-finish","cost":0.0123,"tokens":{"input":100,"output":20,"reasoning":5,"cache":{"read":7,"write":3}}}}`)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestParseEvents(t *testing.T) {
	raw := []byte(strings.Join([]string{
		`{"type":"step_start","sessionID":"ses_1","part":{"type":"step-start"}}`,
		`{"type":"text","part":{"type":"text","text":"hello "}}`,
		`{"type":"text","part":{"type":"text","text":"world"}}`,
		`{"type":"tool_use","part":{"type":"tool-use"}}`,
		`{"type":"step_finish","part":{"type":"step-finish","cost":0.5,"tokens":{"input":10,"output":20,"reasoning":3,"cache":{"read":4,"write":5}}}}`,
	}, "\n"))
	text, tokens, cost, session, tools, events := parseEvents(raw)
	if text != "hello world" {
		t.Fatalf("text = %q", text)
	}
	if tokens.Input != 10 || tokens.Output != 20 || tokens.Reasoning != 3 || tokens.CacheRead != 4 || tokens.CacheWrite != 5 {
		t.Fatalf("tokens = %+v", tokens)
	}
	if cost != 0.5 || session != "ses_1" || tools != 1 || events != 5 {
		t.Fatalf("cost=%f session=%q tools=%d events=%d", cost, session, tools, events)
	}
}

func TestPrepareWorkspaceIsolation(t *testing.T) {
	root := t.TempDir()
	task := Task{ID: "ISO-01", Files: map[string]string{"go.mod": "module x\n", "src/main.go": "package main\n"}}
	first, err := prepareWorkspace(root, "first", task)
	if err != nil {
		t.Fatalf("first workspace: %v", err)
	}
	second, err := prepareWorkspace(root, "second", task)
	if err != nil {
		t.Fatalf("second workspace: %v", err)
	}
	if first == second {
		t.Fatal("workspaces are not isolated")
	}
	for _, dir := range []string{first, second} {
		if _, err := os.Stat(filepath.Join(dir, "src", "main.go")); err != nil {
			t.Fatalf("baseline file missing in %s: %v", dir, err)
		}
	}
}

func TestRunOneWithFakeBinary(t *testing.T) {
	t.Setenv("AIL_FAKE_OPENCODE", "1")
	output := t.TempDir()
	workspaces := t.TempDir()
	request := Request{
		Task: Task{
			ID:       "FAKE-01",
			Category: "implementation",
			Prompt:   "Do the thing",
			Files:    map[string]string{"go.mod": "module x\n"},
		},
		Model:          ModelSpec{ID: "opencode-go/hy4-flash-free", Label: "Hy4 Flash Free"},
		Format:         "markdown",
		Delimiters:     false,
		Condition:      "A",
		Instruction:    "rules here",
		WorkspaceRoot:  workspaces,
		OutputDir:      output,
		OpenCodeBinary: os.Args[0],
	}
	record, err := RunOne(request)
	if err != nil {
		t.Fatalf("RunOne: %v", err)
	}
	if !record.Success {
		t.Fatalf("record not successful: %+v", record)
	}
	if record.Text != "fake answer" {
		t.Fatalf("text = %q", record.Text)
	}
	if record.Tokens.Input != 100 || record.Cost != 0.0123 {
		t.Fatalf("usage not captured: %+v", record)
	}
	if record.ToolCalls != 1 || record.SessionID != "ses_fake" {
		t.Fatalf("tool/session not captured: %+v", record)
	}
	if _, err := os.Stat(filepath.Join(output, record.ID+".events.jsonl")); err != nil {
		t.Fatalf("raw events missing: %v", err)
	}
}

func TestRunOneResume(t *testing.T) {
	t.Setenv("AIL_FAKE_OPENCODE", "1")
	output := t.TempDir()
	request := Request{
		Task:           Task{ID: "RESUME-01", Category: "debugging", Prompt: "Fix", Files: map[string]string{"go.mod": "module x\n"}},
		Model:          ModelSpec{ID: "opencode-go/beam-free", Label: "Beam Free"},
		Format:         "ail",
		Delimiters:     true,
		Condition:      "D",
		Instruction:    "rules",
		WorkspaceRoot:  t.TempDir(),
		OutputDir:      output,
		OpenCodeBinary: os.Args[0],
	}
	first, err := RunOne(request)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	request.OpenCodeBinary = "definitely-not-a-real-binary"
	second, err := RunOne(request)
	if err != nil {
		t.Fatalf("resumed run: %v", err)
	}
	if second.ID != first.ID || !second.Success {
		t.Fatalf("resume did not reuse record: %+v", second)
	}
}

func TestWriteSummaries(t *testing.T) {
	output := t.TempDir()
	record := RunRecord{
		ID: "SUM-01", TaskID: "T-01", Category: "implementation",
		Model: "opencode-go/hy4-flash-free", Format: "markdown", Condition: "A", Success: true,
		Tokens: Tokens{Input: 10, Output: 5},
	}
	data, _ := json.Marshal(record)
	if err := os.WriteFile(filepath.Join(output, "SUM-01.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteSummaries(output); err != nil {
		t.Fatalf("WriteSummaries: %v", err)
	}
	jsonData, err := os.ReadFile(filepath.Join(output, "RESULTS.json"))
	if err != nil {
		t.Fatal(err)
	}
	var records []RunRecord
	if err := json.Unmarshal(jsonData, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].ID != "SUM-01" {
		t.Fatalf("unexpected records: %+v", records)
	}
	csvData, err := os.ReadFile(filepath.Join(output, "RESULTS.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(csvData), "SUM-01") {
		t.Fatalf("CSV missing record: %s", csvData)
	}
}

func TestGenerateCatalogShape(t *testing.T) {
	catalog := GenerateCatalog()
	if len(catalog.Tasks) != 50 {
		t.Fatalf("tasks = %d", len(catalog.Tasks))
	}
	counts := make(map[string]int)
	for _, task := range catalog.Tasks {
		counts[task.Category]++
		if task.Prompt == "" || task.VerifyCommand == "" || len(task.Files) == 0 {
			t.Fatalf("incomplete task %s", task.ID)
		}
	}
	want := map[string]int{
		"implementation": 10, "debugging": 10, "refactoring": 10,
		"testing": 5, "architecture": 5, "code-review": 5, "ambiguous-requirement": 5,
	}
	for category, expected := range want {
		if counts[category] != expected {
			t.Fatalf("%s = %d, want %d", category, counts[category], expected)
		}
	}
}
