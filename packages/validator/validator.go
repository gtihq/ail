// Package validator is an independent structural and semantic validator. It
// gates correctness and returns machine-readable results; it shares no logic
// with the advisory linter.
package validator

import (
	"fmt"
	"strings"

	"github.com/gtihq/ail/packages/core"
)

// SupportedVersion is the only AIL version this toolchain accepts.
const SupportedVersion = "0.3"

var prioritySet = map[core.Priority]bool{
	core.PriorityCritical: true,
	core.PriorityHigh:     true,
	core.PriorityMedium:   true,
	core.PriorityLow:      true,
	core.PriorityInfo:     true,
}

var executionSet = map[string]bool{
	"declarative": true,
	"autonomous":  true,
	"interactive": true,
}

var builtinRoots = map[string]bool{
	"file": true, "env": true, "platform": true, "agent": true,
	"project": true, "scope": true, "context": true, "language": true,
	"user": true, "service": true, "tooling": true,
}

type validator struct {
	file    string
	errors  []core.Diagnostic
	warning []core.Diagnostic
}

func (v *validator) fail(code, message string, loc core.SourceLocation) {
	v.errors = append(v.errors, core.MakeDiagnostic(code, core.SeverityError, core.KindSemantic, message, v.file, loc))
}

func (v *validator) warn(code, message string, loc core.SourceLocation) {
	v.warning = append(v.warning, core.MakeDiagnostic(code, core.SeverityWarning, core.KindSemantic, message, v.file, loc))
}

func walk(expression core.Expression, visit func(core.Expression)) {
	visit(expression)
	switch node := expression.(type) {
	case *core.UnaryExpr:
		walk(node.Operand, visit)
	case *core.BinaryExpr:
		walk(node.Left, visit)
		walk(node.Right, visit)
	}
}

// Validate checks a parse result and its normalized IR for structural and semantic errors.
func Validate(parsed core.ParseResult, ir *core.IR) core.ValidationResult {
	file := ""
	if parsed.AST != nil {
		file = parsed.AST.File
	}
	v := &validator{file: file, errors: []core.Diagnostic{}, warning: []core.Diagnostic{}}

	// #region Syntax validity
	for _, diagnostic := range parsed.Diagnostics {
		if diagnostic.Severity == core.SeverityError {
			v.errors = append(v.errors, diagnostic)
		} else {
			v.warning = append(v.warning, diagnostic)
		}
	}
	// #endregion

	if parsed.AST == nil {
		v.fail("VAL_MISSING_AST", "No abstract syntax tree was produced", core.SourceLocation{Line: 1, Column: 1})
		return buildResult(v, 0)
	}

	// #region Metadata and version
	meta := parsed.AST.Meta
	if meta == nil {
		v.fail("VAL_MISSING_META", "META block is required", parsed.AST.Loc)
	} else {
		hasID := false
		for _, field := range meta.Fields {
			if field.Key == "" {
				v.fail("VAL_EMPTY_FIELD", "META field name is empty", field.Loc)
			}
			if field.Key == "id" && field.Value.String() != "" {
				hasID = true
			}
			if field.Key == "version" && field.Value.String() != SupportedVersion {
				v.fail("VAL_VERSION_INCOMPATIBLE", fmt.Sprintf("Unsupported AIL version %q", field.Value.String()), field.Loc)
			}
			if field.Key == "priority" && !prioritySet[core.Priority(field.Value.String())] {
				v.fail("VAL_INVALID_PRIORITY", fmt.Sprintf("Invalid priority %q in META", field.Value.String()), field.Loc)
			}
		}
		if !hasID {
			v.fail("VAL_MISSING_META_ID", "META must declare a non-empty id", meta.Loc)
		}
	}
	// #endregion

	// #region AST declarations and expressions
	declaredContexts := make(map[string]map[string]bool)
	for _, rule := range parsed.AST.Rules {
		if len(rule.Fields) == 0 {
			v.fail("VAL_EMPTY_RULE", "Rule has no declarations", rule.Loc)
		}
		context := make(map[string]bool)
		declaredContexts[rule.ID] = context
		for _, field := range rule.Fields {
			switch value := field.(type) {
			case *core.ContextField:
				for _, item := range value.Fields {
					if item.Key == "" {
						v.fail("VAL_EMPTY_FIELD", "CONTEXT field name is empty", item.Loc)
					}
					if item.Value.Kind != core.LiteralString && item.Value.Kind != core.LiteralNumber && item.Value.Kind != core.LiteralBool {
						v.fail("VAL_UNSUPPORTED_TYPE", fmt.Sprintf("Unsupported literal type %q", item.Value.Kind), item.Loc)
					}
					context[item.Key] = true
				}
			case *core.WhenField:
				validateExpressions(v, value.Expressions, context)
			case *core.AssertField:
				validateExpressions(v, value.Expressions, context)
			case *core.ScopeField:
				if value.Value == "" {
					v.fail("VAL_EMPTY_SCOPE", "SCOPE must not be empty", value.Loc)
				}
				if strings.ContainsAny(value.Value, " \t\r\n") || strings.Contains(value.Value, "..") {
					v.fail("VAL_INVALID_SCOPE", fmt.Sprintf("Invalid scope %q", value.Value), value.Loc)
				}
			case *core.ExecutionField:
				if !executionSet[strings.ToLower(value.Value)] {
					v.fail("VAL_INVALID_EXECUTION", fmt.Sprintf("Unsupported execution policy %q", value.Value), value.Loc)
				}
			}
		}
	}
	// #endregion

	if ir == nil {
		v.fail("VAL_MISSING_IR", "No intermediate representation was produced", parsed.AST.Loc)
		return buildResult(v, len(parsed.AST.Rules))
	}

	// #region IR integrity
	if ir.Kind != "IR" {
		v.fail("VAL_IR_KIND", "IR kind must be \"IR\"", parsed.AST.Loc)
	}
	if ir.Version != SupportedVersion {
		v.fail("VAL_IR_VERSION", fmt.Sprintf("IR version %q is not supported", ir.Version), parsed.AST.Loc)
	}
	ids := make(map[string]bool)
	for _, rule := range ir.Rules {
		loc := rule.Source.ToLocation()
		if rule.ID == "" {
			v.fail("VAL_EMPTY_RULE_ID", "IR rule has an empty id", loc)
		}
		if ids[rule.ID] {
			v.fail("VAL_DUPLICATE_RULE", fmt.Sprintf("Duplicate rule id %q", rule.ID), loc)
		}
		ids[rule.ID] = true
		if !prioritySet[rule.Priority] {
			v.fail("VAL_INVALID_PRIORITY", fmt.Sprintf("Invalid resolved priority %q", rule.Priority), loc)
		}
	}
	if len(ir.Provenance) == 0 {
		v.fail("VAL_MISSING_PROVENANCE", "IR has no provenance entries", parsed.AST.Loc)
	}
	// #endregion

	// #region References, dependencies and cycles
	for _, rule := range ir.Rules {
		loc := rule.Source.ToLocation()
		for _, dependency := range rule.Requires {
			if !ids[dependency] {
				v.fail("VAL_UNRESOLVED_REFERENCE", fmt.Sprintf("Undefined required rule %q", dependency), loc)
			}
		}
		for _, parent := range rule.Inherit {
			if !ids[parent] {
				v.fail("VAL_UNRESOLVED_INHERIT", fmt.Sprintf("Undefined inherited rule %q", parent), loc)
			}
		}
	}
	if cycle := findInheritanceCycle(ir.Rules); cycle != "" {
		v.fail("VAL_INHERITANCE_CYCLE", fmt.Sprintf("Inheritance cycle involving %q", cycle), parsed.AST.Loc)
	}
	// #endregion

	return buildResult(v, len(parsed.AST.Rules))
}

func validateExpressions(v *validator, expressions []core.Expression, context map[string]bool) {
	for _, expression := range expressions {
		walk(expression, func(node core.Expression) {
			switch value := node.(type) {
			case *core.BinaryExpr:
				if !isSupportedOperator(value.Operator) {
					v.fail("VAL_INVALID_OPERATOR", fmt.Sprintf("Unsupported operator %q", value.Operator), value.Loc)
				}
			case *core.PathExpr:
				if len(value.Parts) == 0 {
					v.fail("VAL_UNDEFINED_IDENTIFIER", "Empty identifier path", value.Loc)
					return
				}
				root := value.Parts[0]
				if len(value.Parts) == 1 && strings.ToUpper(root) == root {
					return
				}
				if !context[root] && !builtinRoots[root] {
					v.warn("VAL_UNDEFINED_IDENTIFIER", fmt.Sprintf("Identifier %q is not declared in context", root), value.Loc)
				}
			}
		})
	}
}

func isSupportedOperator(operator core.BinaryOperator) bool {
	switch operator {
	case core.OpAnd, core.OpOr, core.OpEq, core.OpNe, core.OpGt, core.OpLt, core.OpGte, core.OpLte, core.OpIn:
		return true
	default:
		return false
	}
}

func findInheritanceCycle(rules []core.IRRule) string {
	graph := make(map[string][]string)
	for _, rule := range rules {
		graph[rule.ID] = rule.Inherit
	}
	state := make(map[string]int)
	var visit func(string) string
	visit = func(id string) string {
		if state[id] == 1 {
			return id
		}
		if state[id] == 2 {
			return ""
		}
		state[id] = 1
		for _, parent := range graph[id] {
			if found := visit(parent); found != "" {
				return found
			}
		}
		state[id] = 2
		return ""
	}
	for _, rule := range rules {
		if found := visit(rule.ID); found != "" {
			return found
		}
	}
	return ""
}

func buildResult(v *validator, ruleCount int) core.ValidationResult {
	result := core.ValidationResult{
		Valid:   len(v.errors) == 0,
		Errors:  v.errors,
		Warning: v.warning,
	}
	result.Summary.Rules = ruleCount
	result.Summary.Errors = len(v.errors)
	result.Summary.Warnings = len(v.warning)
	return result
}
