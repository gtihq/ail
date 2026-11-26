package interpreter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/parser"
)

func interpretSource(t *testing.T, source string) interpreter.InterpretResult {
	t.Helper()
	return interpreter.Interpret(parser.Parse(source, "test.ail"))
}

func TestIRGeneration(t *testing.T) {
	result := interpretSource(t, `META { id: "basic" priority: HIGH scope: "src/**/*.ts" }
DO { first_rule second_rule }
NEVER { forbidden_rule }`)
	if len(result.IR.Rules) != 1 {
		t.Fatalf("rules = %d, want 1", len(result.IR.Rules))
	}
	rule := result.IR.Rules[0]
	if rule.ID != "basic" {
		t.Fatalf("id = %q", rule.ID)
	}
	if rule.Priority != core.PriorityHigh {
		t.Fatalf("priority = %q", rule.Priority)
	}
	if rule.Scope != "src/**/*.ts" {
		t.Fatalf("scope = %q", rule.Scope)
	}
	if len(rule.Do) != 2 || len(rule.Never) != 1 {
		t.Fatalf("do/never = %d/%d", len(rule.Do), len(rule.Never))
	}
	if rule.Source.File != "test.ail" || rule.Source.Line < 1 {
		t.Fatalf("source = %+v", rule.Source)
	}
}

func TestPriorityAndScopeFromMeta(t *testing.T) {
	result := interpretSource(t, `META { id: "meta" priority: CRITICAL scope: "platform/**/*.ts" }
DO { act }`)
	if result.IR.Rules[0].Priority != core.PriorityCritical {
		t.Fatalf("priority = %q", result.IR.Rules[0].Priority)
	}
	if result.IR.Rules[0].Scope != "platform/**/*.ts" {
		t.Fatalf("scope = %q", result.IR.Rules[0].Scope)
	}
}

func TestInheritance(t *testing.T) {
	result := interpretSource(t, `META { id: "project" }
RULE R001 { priority: HIGH DO { inherited } NEVER { base_forbidden } }
RULE R002 { inherit R001 priority: CRITICAL DO { own } }`)
	if len(result.IR.Rules) != 2 {
		t.Fatalf("rules = %d", len(result.IR.Rules))
	}
	byID := make(map[string]core.IRRule)
	for _, rule := range result.IR.Rules {
		byID[rule.ID] = rule
	}
	child := byID["R002"]
	if len(child.Do) != 1 || child.Do[0] != "own" {
		t.Fatalf("child DO = %+v", child.Do)
	}
	if len(child.Never) != 1 || child.Never[0] != "base_forbidden" {
		t.Fatalf("child NEVER = %+v", child.Never)
	}
	if child.Priority != core.PriorityCritical {
		t.Fatalf("child priority = %q", child.Priority)
	}
	if len(child.InheritedFrom) != 1 || child.InheritedFrom[0] != "R001" {
		t.Fatalf("inheritedFrom = %+v", child.InheritedFrom)
	}
}

func TestInheritanceCycle(t *testing.T) {
	result := interpretSource(t, `META { id: "cycle" }
RULE R001 { inherit R002 DO { a } }
RULE R002 { inherit R001 DO { b } }`)
	found := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == "INT_INHERITANCE_CYCLE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("cycle diagnostic missing: %+v", result.Diagnostics)
	}
}

func TestUnresolvedDependency(t *testing.T) {
	result := interpretSource(t, `META { id: "deps" }
RULE R001 { requires R999 DO { act } }`)
	found := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == "INT_UNRESOLVED_REFERENCE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("dependency diagnostic missing: %+v", result.Diagnostics)
	}
}

func TestDeterministicOutput(t *testing.T) {
	source := `META { id: "deterministic" }
RULE R001 { DO { b } }
RULE R002 { INHERIT R001 DO { a } }`
	first := interpretSource(t, source)
	second := interpretSource(t, source)
	firstJSON, _ := json.Marshal(first.IR)
	secondJSON, _ := json.Marshal(second.IR)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("IR is not deterministic:\n%s\n%s", firstJSON, secondJSON)
	}
	if !strings.Contains(string(firstJSON), `"kind":"IR"`) {
		t.Fatalf("IR kind missing: %s", firstJSON)
	}
}
