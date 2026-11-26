package core

import "strings"

// ExpressionText renders an expression tree into canonical deterministic text.
// It is a normalization aid only, never an evaluation of the expression.
func ExpressionText(expression Expression) string {
	switch node := expression.(type) {
	case *LiteralExpr:
		return node.Value.String()
	case *PathExpr:
		return strings.Join(node.Parts, ".")
	case *UnaryExpr:
		operand := ExpressionText(node.Operand)
		if _, isBinary := node.Operand.(*BinaryExpr); isBinary {
			return "!(" + operand + ")"
		}
		return "!" + operand
	case *BinaryExpr:
		return ExpressionText(node.Left) + " " + string(node.Operator) + " " + ExpressionText(node.Right)
	default:
		return ""
	}
}

// ExpressionTexts renders a list of expressions into canonical text.
func ExpressionTexts(expressions []Expression) []string {
	texts := make([]string, 0, len(expressions))
	for _, expression := range expressions {
		texts = append(texts, ExpressionText(expression))
	}
	return texts
}
