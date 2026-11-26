package linter_test

import (
	"testing"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/linter"
	"github.com/gtihq/ail/packages/parser"
)

func lintSource(source string, config core.LintConfig) core.AnalyzerResult {
	parsed := parser.Parse(source, "lint.ail")
	interpreted := interpreter.Interpret(parsed)
	return linter.Lint(parsed, interpreted.IR, config)
}

func hasCode(diagnostics []core.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestEveryDiagnosticCategory(t *testing.T) {
	cases := []struct {
		name   string
		source string
		code   string
	}{
		{"duplicate-rule", `META { id: "p" } RULE R001 { DO { a } } RULE R001 { DO { b } }`, "LINT_DUPLICATE_RULE"},
		{"redundant", `META { id: "p" } DO { a a }`, "LINT_REDUNDANT_DO"},
		{"contradictory", `META { id: "p" } DO { a } NEVER { a }`, "LINT_CONTRADICTION"},
		{"ambiguous", `META { id: "p" } WHEN { a && !a } DO { act }`, "LINT_AMBIGUOUS_CONDITION"},
		{"nesting", `META { id: "p" } WHEN { !!!!!!!file.ready } DO { act }`, "LINT_NESTING_DEPTH"},
		{"unused", `META { id: "p" } CONTEXT { x: "1" } DO { act }`, "LINT_UNUSED_CONTEXT"},
		{"unreachable", `META { id: "p" } WHEN { false } DO { act }`, "LINT_UNREACHABLE"},
		{"fallback", `META { id: "p" } DO { act }`, "LINT_MISSING_FALLBACK"},
		{"scope", `META { id: "p" } SCOPE "**" DO { act }`, "LINT_BROAD_SCOPE"},
		{"complexity", `META { id: "p" } DO { a b c d e f g h i j k l m n o p q r s t u }`, "LINT_COMPLEXITY"},
		{"naming", `META { id: "p" } RULE weird { DO { CamelCase } }`, "LINT_RULE_ID_NAMING"},
		{"verbosity", `META { id: "p" } DO { a } VERIFY { a }`, "LINT_VERIFY_ECHO"},
		{"ineffective", `META { id: "p" } DO { consider }`, "LINT_SOFT_TERM"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			result := lintSource(item.source, core.LintConfig{})
			if !hasCode(result.Diagnostics, item.code) {
				t.Fatalf("expected %s, got %+v", item.code, result.Diagnostics)
			}
		})
	}
}

func TestValidRulesDoNotProduceErrors(t *testing.T) {
	result := lintSource(`META { id: "clean" version: "0.3" }
WHEN { file.language == TYPESCRIPT }
DO { use_strict_types }
NEVER { use_any }
VERIFY { types_compile }
FALLBACK { ask_user }`, core.LintConfig{})
	if result.Summary.Errors != 0 {
		t.Fatalf("unexpected errors: %+v", result.Diagnostics)
	}
}

func TestSeverityOverride(t *testing.T) {
	source := `META { id: "p" } DO { act }`
	defaultResult := lintSource(source, core.LintConfig{})
	overridden := lintSource(source, core.LintConfig{
		core.LintMissingFallback: core.SeverityError,
	})
	if defaultResult.Summary.Errors != 0 {
		t.Fatalf("default should have no errors: %+v", defaultResult.Diagnostics)
	}
	if overridden.Summary.Errors == 0 {
		t.Fatalf("override should promote to error: %+v", overridden.Diagnostics)
	}
}

func TestSemanticAndStyleAreDistinguished(t *testing.T) {
	result := lintSource(`META { id: "p" } DO { a } NEVER { a } FALLBACK { b }`, core.LintConfig{})
	semantic := false
	style := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Kind == core.KindSemantic {
			semantic = true
		}
		if diagnostic.Kind == core.KindStyle {
			style = true
		}
	}
	if !semantic || !style {
		t.Fatalf("expected both semantic and style findings: %+v", result.Diagnostics)
	}
}
