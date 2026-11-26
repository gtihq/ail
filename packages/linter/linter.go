// Package linter is an independent advisory static-analysis engine. It separates
// semantic problems from style preferences and supports severity overrides.
package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gtihq/ail/packages/core"
)

var (
	ruleIDPattern = regexp.MustCompile(`^R[0-9]+$`)
	snakePattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

var softWords = map[string]bool{
	"consider": true,
	"maybe":    true,
	"try":      true,
	"should":   true,
	"might":    true,
	"perhaps":  true,
	"ideally":  true,
}

type linter struct {
	file        string
	config      core.LintConfig
	diagnostics []core.Diagnostic
}

func (l *linter) report(category core.LintCategory, code string, fallback core.Severity, kind core.DiagnosticKind, message string, loc core.SourceLocation) {
	severity := fallback
	if configured, ok := l.config[category]; ok {
		severity = configured
	}
	l.diagnostics = append(l.diagnostics, core.MakeDiagnostic(code, severity, kind, message, l.file, loc))
}

func walkExpression(expression core.Expression, visit func(core.Expression)) {
	visit(expression)
	switch node := expression.(type) {
	case *core.UnaryExpr:
		walkExpression(node.Operand, visit)
	case *core.BinaryExpr:
		walkExpression(node.Left, visit)
		walkExpression(node.Right, visit)
	}
}

// ambiguousExpression reports a structural contradiction or tautology.
func ambiguousExpression(expression core.Expression) bool {
	binary, ok := expression.(*core.BinaryExpr)
	if !ok {
		return false
	}
	if binary.Operator != core.OpAnd && binary.Operator != core.OpOr {
		return false
	}
	leftText := core.ExpressionText(binary.Left)
	rightText := core.ExpressionText(binary.Right)
	leftNegated, leftIsUnary := binary.Left.(*core.UnaryExpr)
	rightNegated, rightIsUnary := binary.Right.(*core.UnaryExpr)
	if leftIsUnary && rightText == core.ExpressionText(leftNegated.Operand) {
		return true
	}
	if rightIsUnary && leftText == core.ExpressionText(rightNegated.Operand) {
		return true
	}
	return false
}

func expressionDepth(expression core.Expression) int {
	switch node := expression.(type) {
	case *core.BinaryExpr:
		left, right := expressionDepth(node.Left), expressionDepth(node.Right)
		if left > right {
			return left + 1
		}
		return right + 1
	case *core.UnaryExpr:
		return expressionDepth(node.Operand) + 1
	default:
		return 0
	}
}

// Lint runs all advisory checks over the AST and IR.
func Lint(parsed core.ParseResult, ir *core.IR, config core.LintConfig) core.AnalyzerResult {
	file := ""
	if parsed.AST != nil {
		file = parsed.AST.File
	}
	l := &linter{file: file, config: config, diagnostics: []core.Diagnostic{}}

	// #region AST-only checks
	if parsed.AST != nil {
		for _, rule := range parsed.AST.Rules {
			declared := make(map[string]bool)
			referenced := make(map[string]bool)
			for _, field := range rule.Fields {
				switch value := field.(type) {
				case *core.ContextField:
					for _, item := range value.Fields {
						declared[item.Key] = true
					}
				case *core.WhenField:
					for _, expression := range value.Expressions {
						walkExpression(expression, func(node core.Expression) {
							if path, ok := node.(*core.PathExpr); ok && len(path.Parts) > 0 {
								referenced[path.Parts[0]] = true
							}
							if ambiguousExpression(node) {
								l.report(core.LintAmbiguousCondition, "LINT_AMBIGUOUS_CONDITION", core.SeverityWarning, core.KindSemantic,
									"Condition combines a term with its negation", node.NodeLocation())
							}
							if expressionDepth(node) > 6 {
								l.report(core.LintUnnecessaryNesting, "LINT_NESTING_DEPTH", core.SeverityHint, core.KindStyle,
									"Expression nesting exceeds a readable depth", node.NodeLocation())
							}
						})
					}
				case *core.AssertField:
					for _, expression := range value.Expressions {
						walkExpression(expression, func(node core.Expression) {
							if path, ok := node.(*core.PathExpr); ok && len(path.Parts) > 0 {
								referenced[path.Parts[0]] = true
							}
						})
					}
				}
			}
			for key := range declared {
				if !referenced[key] {
					l.report(core.LintUnusedDeclaration, "LINT_UNUSED_CONTEXT", core.SeverityInfo, core.KindStyle,
						fmt.Sprintf("Context field %q is never referenced", key), rule.Loc)
				}
			}
		}
	}
	// #endregion

	if ir == nil {
		return core.AnalyzerResult{Diagnostics: l.diagnostics, Summary: core.Summarize(l.diagnostics)}
	}

	seenIDs := make(map[string]bool)
	for _, rule := range ir.Rules {
		loc := rule.Source.ToLocation()

		// #region Semantic checks
		if seenIDs[rule.ID] {
			l.report(core.LintDuplicateRule, "LINT_DUPLICATE_RULE", core.SeverityError, core.KindSemantic,
				fmt.Sprintf("Duplicate rule id %q", rule.ID), loc)
		}
		seenIDs[rule.ID] = true

		doSet := makeSet(rule.Do)
		neverSet := makeSet(rule.Never)
		verifySet := makeSet(rule.Verify)
		for term := range doSet {
			if neverSet[term] {
				l.report(core.LintContradictory, "LINT_CONTRADICTION", core.SeverityError, core.KindSemantic,
					fmt.Sprintf("Term %q is both required and forbidden", term), loc)
			}
		}
		for _, condition := range rule.When {
			if condition == "false" {
				l.report(core.LintUnreachableRule, "LINT_UNREACHABLE", core.SeverityWarning, core.KindSemantic,
					"Rule condition is permanently false", loc)
			}
		}
		// #endregion

		// #region Style checks
		checkRedundant(rule.Do, loc, func(message string, location core.SourceLocation) {
			l.report(core.LintRedundantInstruction, "LINT_REDUNDANT_DO", core.SeverityWarning, core.KindStyle, message, location)
		})
		checkRedundant(rule.Never, loc, func(message string, location core.SourceLocation) {
			l.report(core.LintRedundantInstruction, "LINT_REDUNDANT_NEVER", core.SeverityWarning, core.KindStyle, message, location)
		})
		checkRedundant(rule.Verify, loc, func(message string, location core.SourceLocation) {
			l.report(core.LintRedundantInstruction, "LINT_REDUNDANT_VERIFY", core.SeverityWarning, core.KindStyle, message, location)
		})

		for term := range doSet {
			if verifySet[term] {
				l.report(core.LintUnnecessaryVerbosity, "LINT_VERIFY_ECHO", core.SeverityHint, core.KindStyle,
					fmt.Sprintf("Term %q appears in both DO and VERIFY", term), loc)
			}
		}

		if (len(rule.Do) > 0 || len(rule.Never) > 0) && len(rule.Fallback) == 0 {
			l.report(core.LintMissingFallback, "LINT_MISSING_FALLBACK", core.SeverityWarning, core.KindStyle,
				"Rule has instructions but no FALLBACK behavior", loc)
		}

		if rule.Scope == "*" || rule.Scope == "**" || rule.Scope == "**/*" {
			l.report(core.LintOverlyBroadScope, "LINT_BROAD_SCOPE", core.SeverityWarning, core.KindStyle,
				fmt.Sprintf("Scope %q matches every file", rule.Scope), loc)
		}

		complexity := len(rule.Do) + len(rule.Never) + len(rule.Verify) + len(rule.Fallback) + len(rule.When) + len(rule.Assert)
		if complexity > 20 {
			l.report(core.LintExcessiveComplexity, "LINT_COMPLEXITY", core.SeverityInfo, core.KindStyle,
				fmt.Sprintf("Rule has %d instruction units", complexity), loc)
		}

		if rule.ID != "" && !ruleIDPattern.MatchString(rule.ID) {
			l.report(core.LintInconsistentNaming, "LINT_RULE_ID_NAMING", core.SeverityHint, core.KindStyle,
				fmt.Sprintf("Rule id %q does not follow Rnnn naming", rule.ID), loc)
		}
		for _, term := range append(append([]string{}, rule.Do...), rule.Never...) {
			if !snakePattern.MatchString(term) {
				l.report(core.LintInconsistentNaming, "LINT_TERM_NAMING", core.SeverityHint, core.KindStyle,
					fmt.Sprintf("Instruction %q is not snake_case", term), loc)
			}
			if softWords[strings.ToLower(term)] {
				l.report(core.LintIneffectiveInstruction, "LINT_SOFT_TERM", core.SeverityInfo, core.KindStyle,
					fmt.Sprintf("Instruction %q is vague and may be ineffective", term), loc)
			}
		}
		// #endregion
	}

	return core.AnalyzerResult{Diagnostics: l.diagnostics, Summary: core.Summarize(l.diagnostics)}
}

func makeSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func checkRedundant(values []string, fallbackLoc core.SourceLocation, report func(string, core.SourceLocation)) {
	seen := make(map[string]bool)
	for _, value := range values {
		if seen[value] {
			report(fmt.Sprintf("Instruction %q is repeated", value), fallbackLoc)
		}
		seen[value] = true
	}
}
