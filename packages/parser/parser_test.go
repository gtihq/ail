package parser_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/parser"
)

func hasCode(diagnostics []core.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestValidStandalone(t *testing.T) {
	source := `META { id: "standalone" version: "0.3" priority: HIGH scope: "src/**/*.ts" }
WHEN { file.language == TYPESCRIPT }
DO { use_strict_types preserve_existing_patterns }
NEVER { use_any_without_justification }`
	result := parser.Parse(source, "standalone.ail")
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if result.AST.Mode != "standalone" {
		t.Fatalf("mode = %q, want standalone", result.AST.Mode)
	}
	if len(result.AST.Rules) != 1 {
		t.Fatalf("rules = %d, want 1", len(result.AST.Rules))
	}
	if result.AST.Meta == nil || len(result.AST.Meta.Fields) != 4 {
		t.Fatalf("meta fields = %+v", result.AST.Meta)
	}
	terms := 0
	for _, field := range result.AST.Rules[0].Fields {
		if instruction, ok := field.(*core.InstructionField); ok && instruction.Kind == "DO" {
			terms = len(instruction.Terms)
		}
	}
	if terms != 2 {
		t.Fatalf("DO terms = %d, want 2", terms)
	}
}

func TestValidAggregated(t *testing.T) {
	source := `META { id: "project-rules" version: "0.3" }
RULE R001 { priority: HIGH DO { inspect_relevant_files } }
RULE R002 { priority: CRITICAL NEVER { expose_secrets } }`
	result := parser.Parse(source, "aggregated.ail")
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if result.AST.Mode != "aggregated" {
		t.Fatalf("mode = %q, want aggregated", result.AST.Mode)
	}
	if len(result.AST.Rules) != 2 {
		t.Fatalf("rules = %d, want 2", len(result.AST.Rules))
	}
	if result.AST.Rules[0].ID != "R001" || result.AST.Rules[1].ID != "R002" {
		t.Fatalf("ids = %q, %q", result.AST.Rules[0].ID, result.AST.Rules[1].ID)
	}
}

func TestNestedExpressions(t *testing.T) {
	source := `META { id: "expr" }
WHEN { (file.language == TYPESCRIPT && env.mode == "dev") || !file.generated }
DO { act }`
	result := parser.Parse(source, "expr.ail")
	if result.AST.Mode != "standalone" {
		t.Fatalf("mode = %q", result.AST.Mode)
	}
	found := false
	for _, rule := range result.AST.Rules {
		for _, field := range rule.Fields {
			if when, ok := field.(*core.WhenField); ok {
				for _, expression := range when.Expressions {
					if _, ok := expression.(*core.BinaryExpr); ok {
						found = true
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("expected a binary expression in WHEN")
	}
}

func TestMissingDelimiters(t *testing.T) {
	result := parser.Parse(`META { id: "broken"`, "broken.ail")
	if !hasCode(result.Diagnostics, "PARSE_EXPECTED_RBRACE") {
		t.Fatalf("missing expected RBRACE diagnostic: %+v", result.Diagnostics)
	}
}

func TestMissingFields(t *testing.T) {
	result := parser.Parse(`META { id "x" }`, "missing.ail")
	if !hasCode(result.Diagnostics, "PARSE_EXPECTED_COLON") {
		t.Fatalf("missing colon diagnostic: %+v", result.Diagnostics)
	}
}

func TestInvalidOperators(t *testing.T) {
	result := parser.Parse(`META { id: "x" } WHEN { file.language = TYPESCRIPT }`, "operator.ail")
	if !hasCode(result.Diagnostics, "LEX_UNEXPECTED_OPERATOR") {
		t.Fatalf("missing operator diagnostic: %+v", result.Diagnostics)
	}
}

func TestSourceLocations(t *testing.T) {
	result := parser.Parse(`META { id: "loc" }
DO { first_term }`, "loc.ail")
	for _, rule := range result.AST.Rules {
		for _, field := range rule.Fields {
			if instruction, ok := field.(*core.InstructionField); ok && len(instruction.Terms) > 0 {
				term := instruction.Terms[0]
				if term.Loc.Line != 2 {
					t.Fatalf("term line = %d, want 2", term.Loc.Line)
				}
				return
			}
		}
	}
	t.Fatal("DO instruction not found")
}

func TestFuzzMalformedInputNeverPanics(t *testing.T) {
	random := rand.New(rand.NewSource(42))
	alphabet := []byte("{}()\":,.!&|=<> \n\r\t/*abcXYZ019_")
	for iteration := 0; iteration < 500; iteration++ {
		length := random.Intn(120)
		var builder strings.Builder
		for i := 0; i < length; i++ {
			builder.WriteByte(alphabet[random.Intn(len(alphabet))])
		}
		result := parser.Parse(builder.String(), "fuzz.ail")
		if result.AST == nil {
			t.Fatalf("iteration %d produced nil AST", iteration)
		}
	}
}
