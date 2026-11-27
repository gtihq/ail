package tests_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/linter"
	"github.com/gtihq/ail/packages/parser"
	"github.com/gtihq/ail/packages/renderer"
	"github.com/gtihq/ail/packages/validator"
)

func pipeline(source string) (core.ParseResult, interpreter.InterpretResult, core.AnalyzerResult, core.ValidationResult, core.RenderResult) {
	parsed := parser.Parse(source, "integration.ail")
	interpreted := interpreter.Interpret(parsed)
	linted := linter.Lint(parsed, interpreted.IR, core.LintConfig{})
	validated := validator.Validate(parsed, interpreted.IR)
	rendered := renderer.Render(interpreted.IR, core.RenderOptions{Host: core.HostMarkdown, Mode: core.ModeInline})
	return parsed, interpreted, linted, validated, rendered
}

func TestFullPipeline(t *testing.T) {
	source, err := os.ReadFile("fixtures/valid-standalone.ail")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	parsed, interpreted, linted, validated, rendered := pipeline(string(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse diagnostics: %+v", parsed.Diagnostics)
	}
	if len(interpreted.IR.Rules) != 1 {
		t.Fatalf("rules = %d", len(interpreted.IR.Rules))
	}
	if linted.Summary.Errors != 0 {
		t.Fatalf("lint errors: %+v", linted.Diagnostics)
	}
	if !validated.Valid {
		t.Fatalf("validation failed: %+v", validated.Errors)
	}
	if !strings.Contains(rendered.Text, renderer.BeginMarker()) {
		t.Fatal("rendered output missing markers")
	}
}

func TestGoldenRender(t *testing.T) {
	source, err := os.ReadFile("fixtures/valid-standalone.ail")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	_, _, _, _, rendered := pipeline(string(source))
	golden, err := os.ReadFile("fixtures/golden-render.md")
	if err != nil {
		t.Fatalf("golden: %v", err)
	}
	if rendered.Text != string(golden) {
		t.Fatalf("render output diverged from golden file\ngot:\n%s\nwant:\n%s", rendered.Text, string(golden))
	}
}

func TestRegressionFalsePositives(t *testing.T) {
	source, err := os.ReadFile("fixtures/valid-standalone.ail")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	parsed := parser.Parse(string(source), "integration.ail")
	interpreted := interpreter.Interpret(parsed)
	linted := linter.Lint(parsed, interpreted.IR, core.LintConfig{})
	validated := validator.Validate(parsed, interpreted.IR)
	for _, diagnostic := range append(linted.Diagnostics, validated.Warning...) {
		if diagnostic.Code == "LINT_BROAD_SCOPE" {
			t.Fatalf("prefixed glob flagged as broad: %+v", diagnostic)
		}
		if diagnostic.Code == "VAL_UNDEFINED_IDENTIFIER" {
			t.Fatalf("uppercase constant flagged as identifier: %+v", diagnostic)
		}
	}
}

func TestPropertyTermPreservation(t *testing.T) {
	random := rand.New(rand.NewSource(7))
	for iteration := 0; iteration < 50; iteration++ {
		count := 1 + random.Intn(10)
		var builder strings.Builder
		builder.WriteString("META { id: \"property\" }\nDO {")
		terms := make([]string, 0, count)
		for i := 0; i < count; i++ {
			term := fmt.Sprintf("term_%d", i)
			terms = append(terms, term)
			builder.WriteString(" " + term)
		}
		builder.WriteString(" }")
		source := builder.String()
		parsed := parser.Parse(source, "property.ail")
		interpreted := interpreter.Interpret(parsed)
		if len(interpreted.IR.Rules) != 1 {
			t.Fatalf("iteration %d: rules = %d", iteration, len(interpreted.IR.Rules))
		}
		actual := interpreted.IR.Rules[0].Do
		if len(actual) != len(terms) {
			t.Fatalf("iteration %d: terms = %d, want %d", iteration, len(actual), len(terms))
		}
		for i := range terms {
			if actual[i] != terms[i] {
				t.Fatalf("iteration %d: term %d = %q, want %q", iteration, i, actual[i], terms[i])
			}
		}
		compact := renderer.Render(interpreted.IR, core.RenderOptions{Host: core.HostMarkdown, Mode: core.ModeCompact}).Text
		verbose := renderer.Render(interpreted.IR, core.RenderOptions{Host: core.HostMarkdown, Mode: core.ModeVerbose}).Text
		for _, term := range terms {
			if !strings.Contains(compact, term) || !strings.Contains(verbose, term) {
				t.Fatalf("iteration %d: rendering lost term %q", iteration, term)
			}
		}
	}
}

func TestPropertyDeterministicRendering(t *testing.T) {
	source := `META { id: "deterministic" } DO { b a c } NEVER { d }`
	_, firstInterpreted, _, _, firstRendered := pipeline(source)
	_, secondInterpreted, _, _, secondRendered := pipeline(source)
	firstIR, _ := json.Marshal(firstInterpreted.IR)
	secondIR, _ := json.Marshal(secondInterpreted.IR)
	if !strings.EqualFold(string(firstIR), string(secondIR)) {
		t.Fatalf("IR diverged:\n%s\n%s", firstIR, secondIR)
	}
	if firstRendered.Text != secondRendered.Text {
		t.Fatal("rendered text diverged")
	}
}
