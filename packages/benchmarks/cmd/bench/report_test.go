package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/benchmarks/leakage"
	"github.com/gtihq/ail/packages/benchmarks/runner"
)

func TestFilterRecords(t *testing.T) {
	records := []runner.RunRecord{
		{ID: "A-1", Condition: "A"},
		{ID: "D-1", Condition: "D"},
		{ID: "A-2", Condition: "A"},
	}
	filtered := filterRecords(records, func(record runner.RunRecord) bool { return record.Condition == "A" })
	if len(filtered) != 2 {
		t.Fatalf("filtered = %d, want 2", len(filtered))
	}
	if filtered[0].ID != "A-1" || filtered[1].ID != "A-2" {
		t.Fatalf("unexpected order: %+v", filtered)
	}
}

func TestReportCommand(t *testing.T) {
	dir := t.TempDir()
	records := []runner.RunRecord{
		{
			ID: "A__T-1", TaskID: "T-1", Model: "opencode-go/hy4-flash-free",
			Condition: "A", Success: true, VerifyPassed: true,
			Tokens: runner.Tokens{Input: 100, Output: 20},
			DurationMS: 1000, ToolCalls: 3,
			Compliance: runner.Compliance{Checked: 1, Followed: 1},
			Leakage: leakage.Verdict{Score: 1},
		},
		{
			ID: "D__T-1", TaskID: "T-1", Model: "opencode-go/hy4-flash-free",
			Condition: "D", Success: true, VerifyPassed: true,
			Tokens: runner.Tokens{Input: 80, Output: 25},
			DurationMS: 900, ToolCalls: 2,
			Compliance: runner.Compliance{Checked: 1, Followed: 0},
			Leakage: leakage.Verdict{Score: 0, ContainsRuleText: true, UnsolicitedReproduction: true},
		},
	}
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "RESULTS.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(dir, "REPORT.md")
	reportCommand([]string{"-output", dir, "-destination", destination})

	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("report not written: %v", err)
	}
	for _, expected := range []string{
		"Executions recorded: 2",
		"## Coverage by condition",
		"## Paired comparison A versus D",
		"## Rule leakage by condition",
		"## Compliance",
		"Machine-checkable rules followed: 1 of 2 (50.0%)",
	} {
		if !strings.Contains(string(content), expected) {
			t.Fatalf("report missing %q", expected)
		}
	}
}
