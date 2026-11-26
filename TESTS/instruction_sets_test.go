package tests_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/parser"
	"github.com/gtihq/ail/packages/validator"
)

func TestGeneratedInstructionSetsAreEquivalent(t *testing.T) {
	source, err := os.ReadFile("../AIL_RULES/project-rules.ail")
	if err != nil {
		t.Fatalf("AIL rules: %v", err)
	}
	baseline, err := os.ReadFile("../MARKDOWN_BASELINE/project-rules.md")
	if err != nil {
		t.Fatalf("Markdown baseline: %v", err)
	}
	parsed := parser.Parse(string(source), "project-rules.ail")
	interpreted := interpreter.Interpret(parsed)
	validation := validator.Validate(parsed, interpreted.IR)
	if !validation.Valid {
		t.Fatalf("AIL rules are invalid: %+v", validation.Errors)
	}

	markdown := string(baseline)
	terms := 0
	for _, rule := range interpreted.IR.Rules {
		for _, group := range [][]string{rule.Do, rule.Never, rule.Verify, rule.Fallback} {
			for _, term := range group {
				terms++
				if !strings.Contains(markdown, term) {
					t.Fatalf("Markdown baseline is missing term %q from rule %s", term, rule.ID)
				}
			}
		}
		if !strings.Contains(markdown, "## "+rule.ID) {
			t.Fatalf("Markdown baseline is missing section for %s", rule.ID)
		}
	}
	if terms == 0 {
		t.Fatal("no instruction terms were compared")
	}
}

func TestGeneratedAILUsesSupportedVersion(t *testing.T) {
	source, err := os.ReadFile("../AIL_RULES/project-rules.ail")
	if err != nil {
		t.Fatalf("AIL rules: %v", err)
	}
	parsed := parser.Parse(string(source), "project-rules.ail")
	interpreted := interpreter.Interpret(parsed)
	if interpreted.IR.Version != validator.SupportedVersion {
		t.Fatalf("version = %q, want %q", interpreted.IR.Version, validator.SupportedVersion)
	}
	if len(interpreted.IR.Rules) != 10 {
		t.Fatalf("rules = %d, want 10", len(interpreted.IR.Rules))
	}
}

func TestNoTypographicDashInGeneratedFiles(t *testing.T) {
	paths := []string{
		"../AIL_RULES/project-rules.ail",
		"../MARKDOWN_BASELINE/project-rules.md",
		"../README.md",
	}
	forbidden := string([]rune{0x2014})
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if strings.Contains(string(content), forbidden) {
			t.Fatalf("%s contains a typographic dash", path)
		}
	}
}
