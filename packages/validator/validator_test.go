package validator_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/parser"
	"github.com/gtihq/ail/packages/validator"
)

func validateSource(source string) core.ValidationResult {
	parsed := parser.Parse(source, "validate.ail")
	return validator.Validate(parsed, interpreter.Interpret(parsed).IR)
}

func hasCode(diagnostics []core.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestValidDocument(t *testing.T) {
	result := validateSource(`META { id: "clean" version: "0.3" }
WHEN { file.language == TYPESCRIPT }
DO { act }
FALLBACK { ask }`)
	if !result.Valid {
		t.Fatalf("expected valid, got %+v", result)
	}
	if result.Summary.Rules != 1 {
		t.Fatalf("rules = %d", result.Summary.Rules)
	}
}

func TestMissingMetaID(t *testing.T) {
	result := validateSource(`META { version: "0.3" } DO { act }`)
	if !hasCode(result.Errors, "VAL_MISSING_META_ID") {
		t.Fatalf("missing id diagnostic: %+v", result.Errors)
	}
}

func TestVersionCompatibility(t *testing.T) {
	result := validateSource(`META { id: "p" version: "9.9" } DO { act }`)
	if !hasCode(result.Errors, "VAL_VERSION_INCOMPATIBLE") {
		t.Fatalf("version diagnostic: %+v", result.Errors)
	}
}

func TestUndefinedReference(t *testing.T) {
	result := validateSource(`META { id: "p" } RULE R001 { requires R999 DO { act } }`)
	if !hasCode(result.Errors, "VAL_UNRESOLVED_REFERENCE") {
		t.Fatalf("reference diagnostic: %+v", result.Errors)
	}
}

func TestInheritanceCycle(t *testing.T) {
	result := validateSource(`META { id: "p" }
RULE R001 { inherit R002 DO { a } }
RULE R002 { inherit R001 DO { b } }`)
	if !hasCode(result.Errors, "VAL_INHERITANCE_CYCLE") {
		t.Fatalf("cycle diagnostic: %+v", result.Errors)
	}
}

func TestInvalidScope(t *testing.T) {
	result := validateSource(`META { id: "p" } SCOPE "../secret/*" DO { act }`)
	if !hasCode(result.Errors, "VAL_INVALID_SCOPE") {
		t.Fatalf("scope diagnostic: %+v", result.Errors)
	}
}

func TestInvalidExecution(t *testing.T) {
	result := validateSource(`META { id: "p" } EXECUTION run_arbitrary_code DO { act }`)
	if !hasCode(result.Errors, "VAL_INVALID_EXECUTION") {
		t.Fatalf("execution diagnostic: %+v", result.Errors)
	}
}

func TestDuplicateRule(t *testing.T) {
	result := validateSource(`META { id: "p" } RULE R001 { DO { a } } RULE R001 { DO { b } }`)
	if !hasCode(result.Errors, "VAL_DUPLICATE_RULE") {
		t.Fatalf("duplicate diagnostic: %+v", result.Errors)
	}
}

func TestSyntaxErrorsAreIncluded(t *testing.T) {
	result := validateSource(`META { id: "p" `)
	if result.Valid || result.Summary.Errors == 0 {
		t.Fatalf("syntax errors should be reported: %+v", result)
	}
}

func TestMissingIR(t *testing.T) {
	parsed := parser.Parse(`META { id: "p" } DO { act }`, "validate.ail")
	result := validator.Validate(parsed, nil)
	if !hasCode(result.Errors, "VAL_MISSING_IR") {
		t.Fatalf("missing IR diagnostic: %+v", result.Errors)
	}
}

func TestMachineReadableJSON(t *testing.T) {
	result := validateSource(`META { id: "p" } DO { act }`)
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json error: %v", err)
	}
	for _, key := range []string{`"valid"`, `"errors"`, `"warnings"`, `"summary"`, `"rules"`} {
		if !strings.Contains(string(encoded), key) {
			t.Fatalf("missing key %s in %s", key, encoded)
		}
	}
}
