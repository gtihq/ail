package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gtihq/ail/packages/core"
)

// keyword normalizes a contextual keyword for case-insensitive matching.
func keyword(value string) string {
	return strings.ToUpper(value)
}

var fieldKeywords = map[string]bool{
	"PRIORITY": true, "CONTEXT": true, "WHEN": true, "DO": true,
	"NEVER": true, "VERIFY": true, "FALLBACK": true, "REQUIRES": true,
	"INHERIT": true, "ASSERT": true, "EXECUTION": true, "SCOPE": true,
}

var priorities = map[string]core.Priority{
	"CRITICAL": core.PriorityCritical,
	"HIGH":     core.PriorityHigh,
	"MEDIUM":   core.PriorityMedium,
	"LOW":      core.PriorityLow,
	"INFO":     core.PriorityInfo,
}

type parser struct {
	tokens      []core.Token
	diagnostics []core.Diagnostic
	file        string
	index       int
}

func (p *parser) current() core.Token {
	if p.index < len(p.tokens) {
		return p.tokens[p.index]
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *parser) advance() core.Token {
	token := p.current()
	if p.index < len(p.tokens)-1 {
		p.index++
	}
	return token
}

func (p *parser) check(value string) bool {
	token := p.current()
	return token.Type == core.TokenIdent && keyword(token.Value) == keyword(value)
}

func (p *parser) consumeOptionalColon() {
	if p.current().Type == core.TokenColon {
		p.advance()
	}
}

func (p *parser) isEnd() bool {
	return p.current().Type == core.TokenEOF
}

func (p *parser) span(start core.Token) core.SourceLocation {
	endIndex := p.index - 1
	if endIndex < 0 {
		endIndex = 0
	}
	end := p.tokens[endIndex]
	return core.SourceLocation{
		Line:      start.Loc.Line,
		Column:    start.Loc.Column,
		Offset:    start.Loc.Offset,
		EndOffset: end.Loc.EndOffset,
	}
}

func (p *parser) report(code, message string, token core.Token) {
	p.diagnostics = append(p.diagnostics, core.MakeDiagnostic(
		code,
		core.SeverityError,
		core.KindSyntax,
		message,
		p.file,
		token.Loc,
	))
}

func (p *parser) expect(tokenType core.TokenType, code, message string) core.Token {
	if p.current().Type == tokenType {
		return p.advance()
	}
	p.report(code, message, p.current())
	return p.current()
}

func (p *parser) skipUnexpected(code, message string) {
	p.report(code, message, p.current())
	p.advance()
}

func (p *parser) detectMode() string {
	depth := 0
	for _, token := range p.tokens {
		switch token.Type {
		case core.TokenLBrace:
			depth++
		case core.TokenRBrace:
			depth--
		default:
			if depth == 0 && token.Type == core.TokenIdent && keyword(token.Value) == "RULE" {
				return "aggregated"
			}
		}
	}
	return "standalone"
}

func (p *parser) parseFile() *core.SourceFile {
	start := p.current()
	mode := p.detectMode()
	var meta *core.MetaNode
	rules := make([]*core.RuleNode, 0)
	implicitFields := make([]core.RuleField, 0)
	var implicitStart *core.Token

	for !p.isEnd() {
		if p.check("META") {
			if meta != nil {
				p.report("PARSE_DUPLICATE_META", "Only one META block is allowed per file", p.current())
				p.advance()
				p.skipBalancedBlock()
				continue
			}
			meta = p.parseMeta()
			continue
		}

		if p.check("RULE") {
			if mode == "standalone" {
				p.report("PARSE_RULE_IN_STANDALONE", "RULE blocks are not allowed in standalone mode", p.current())
			}
			rules = append(rules, p.parseRule())
			continue
		}

		if fieldKeywords[keyword(p.current().Value)] {
			token := p.current()
			if implicitStart == nil {
				implicitStart = &token
			}
			implicitFields = append(implicitFields, p.parseRuleFields()...)
			continue
		}

		p.skipUnexpected("PARSE_UNEXPECTED_TOP_LEVEL", fmt.Sprintf("Unexpected token %q at top level", p.current().Value))
	}

	if mode == "standalone" || (implicitStart != nil && len(implicitFields) > 0) {
		syntheticStart := start
		if implicitStart != nil {
			syntheticStart = *implicitStart
		}
		rules = append(rules, &core.RuleNode{
			Loc:    p.span(syntheticStart),
			ID:     "",
			Fields: implicitFields,
		})
	}

	return &core.SourceFile{
		File:  p.file,
		Loc:   p.span(start),
		Mode:  mode,
		Meta:  meta,
		Rules: rules,
	}
}

func (p *parser) skipBalancedBlock() {
	if p.current().Type != core.TokenLBrace {
		return
	}
	depth := 0
	for !p.isEnd() {
		if p.current().Type == core.TokenLBrace {
			depth++
		}
		if p.current().Type == core.TokenRBrace {
			depth--
			p.advance()
			if depth == 0 {
				return
			}
			continue
		}
		p.advance()
	}
}

func (p *parser) parseMeta() *core.MetaNode {
	start := p.advance()
	p.expect(core.TokenLBrace, "PARSE_EXPECTED_LBRACE", "Expected \"{\" after META")
	fields := make([]core.FieldNode, 0)
	for !p.isEnd() && p.current().Type != core.TokenRBrace {
		if p.current().Type == core.TokenComma {
			p.advance()
			continue
		}
		fields = append(fields, p.parseField())
	}
	p.expect(core.TokenRBrace, "PARSE_EXPECTED_RBRACE", "Expected \"}\" to close META block")
	return &core.MetaNode{Loc: p.span(start), Fields: fields}
}

func (p *parser) parseRule() *core.RuleNode {
	start := p.advance()
	id := ""
	token := p.current()
	if token.Type == core.TokenIdent && !fieldKeywords[keyword(token.Value)] {
		id = p.advance().Value
	}
	if p.current().Type != core.TokenLBrace {
		p.report("PARSE_EXPECTED_LBRACE", "Expected \"{\" after RULE declaration", p.current())
	} else {
		p.advance()
	}
	fields := p.parseRuleFields()
	p.expect(core.TokenRBrace, "PARSE_EXPECTED_RBRACE", "Expected \"}\" to close RULE block")
	return &core.RuleNode{Loc: p.span(start), ID: id, Fields: fields}
}

func (p *parser) parseRuleFields() []core.RuleField {
	fields := make([]core.RuleField, 0)
	for !p.isEnd() && p.current().Type != core.TokenRBrace {
		if p.current().Type == core.TokenComma {
			p.advance()
			continue
		}
		if !fieldKeywords[keyword(p.current().Value)] {
			p.skipUnexpected("PARSE_UNEXPECTED_FIELD", fmt.Sprintf("Unexpected token %q in rule", p.current().Value))
			continue
		}
		fields = append(fields, p.parseRuleField())
	}
	return fields
}

func (p *parser) parseRuleField() core.RuleField {
	start := p.current()
	word := keyword(start.Value)
	p.advance()
	p.consumeOptionalColon()

	switch word {
	case "PRIORITY":
		token := p.current()
		priority, ok := priorities[keyword(token.Value)]
		if token.Type != core.TokenIdent || !ok {
			p.report("PARSE_INVALID_PRIORITY", fmt.Sprintf("Unknown priority %q", token.Value), token)
			p.advance()
			return &core.PriorityField{Value: core.PriorityMedium, Loc: p.span(start)}
		}
		p.advance()
		return &core.PriorityField{Value: priority, Loc: p.span(start)}

	case "SCOPE":
		token := p.current()
		if token.Type != core.TokenString {
			p.report("PARSE_EXPECTED_STRING", "SCOPE requires a quoted string", token)
			p.advance()
			return &core.ScopeField{Value: "", Loc: p.span(start)}
		}
		p.advance()
		return &core.ScopeField{Value: token.Value, Loc: p.span(start)}

	case "CONTEXT":
		p.expect(core.TokenLBrace, "PARSE_EXPECTED_LBRACE", "Expected \"{\" after CONTEXT")
		fields := make([]core.FieldNode, 0)
		for !p.isEnd() && p.current().Type != core.TokenRBrace {
			if p.current().Type == core.TokenComma {
				p.advance()
				continue
			}
			fields = append(fields, p.parseField())
		}
		p.expect(core.TokenRBrace, "PARSE_EXPECTED_RBRACE", "Expected \"}\" to close CONTEXT block")
		return &core.ContextField{Fields: fields, Loc: p.span(start)}

	case "WHEN", "ASSERT":
		expressions := p.parseExpressionList()
		if word == "WHEN" {
			return &core.WhenField{Expressions: expressions, Loc: p.span(start)}
		}
		return &core.AssertField{Expressions: expressions, Loc: p.span(start)}

	case "DO", "NEVER", "VERIFY", "FALLBACK":
		p.expect(core.TokenLBrace, "PARSE_EXPECTED_LBRACE", fmt.Sprintf("Expected \"{\" after %s", word))
		terms := make([]core.InstructionTerm, 0)
		for !p.isEnd() && p.current().Type != core.TokenRBrace {
			if p.current().Type == core.TokenComma {
				p.advance()
				continue
			}
			token := p.current()
			if token.Type == core.TokenIdent || token.Type == core.TokenString || token.Type == core.TokenNumber {
				p.advance()
				kind := "identifier"
				if token.Type == core.TokenString {
					kind = "string"
				} else if token.Type == core.TokenNumber {
					kind = "number"
				}
				terms = append(terms, core.InstructionTerm{Kind: kind, Text: token.Value, Loc: token.Loc})
			} else {
				p.skipUnexpected("PARSE_UNEXPECTED_TERM", fmt.Sprintf("Unexpected term %q in %s block", token.Value, word))
			}
		}
		p.expect(core.TokenRBrace, "PARSE_EXPECTED_RBRACE", fmt.Sprintf("Expected \"}\" to close %s block", word))
		return &core.InstructionField{Kind: word, Terms: terms, Loc: p.span(start)}

	case "REQUIRES", "INHERIT":
		ids := make([]string, 0)
		for !p.isEnd() && p.current().Type != core.TokenRBrace {
			if p.current().Type == core.TokenComma {
				p.advance()
				continue
			}
			token := p.current()
			if token.Type != core.TokenIdent || fieldKeywords[keyword(token.Value)] {
				break
			}
			ids = append(ids, p.advance().Value)
		}
		return &core.RefsField{Kind: word, IDs: ids, Loc: p.span(start)}

	case "EXECUTION":
		token := p.current()
		if token.Type != core.TokenIdent && token.Type != core.TokenString {
			p.report("PARSE_EXPECTED_IDENTIFIER", "EXECUTION requires a single identifier", token)
			p.advance()
			return &core.ExecutionField{Value: "", Loc: p.span(start)}
		}
		p.advance()
		return &core.ExecutionField{Value: token.Value, Loc: p.span(start)}

	default:
		p.report("PARSE_UNSUPPORTED_CONSTRUCT", fmt.Sprintf("Unsupported construct %q", start.Value), start)
		return &core.ContextField{Fields: []core.FieldNode{}, Loc: p.span(start)}
	}
}

func (p *parser) parseField() core.FieldNode {
	start := p.current()
	if start.Type != core.TokenIdent {
		p.skipUnexpected("PARSE_EXPECTED_KEY", "Expected a field name")
		return core.FieldNode{Key: "", Value: core.StringLiteral(""), Loc: start.Loc}
	}
	p.advance()
	p.expect(core.TokenColon, "PARSE_EXPECTED_COLON", "Expected \":\" after field name")
	token := p.current()
	var value core.Literal
	switch token.Type {
	case core.TokenString:
		value = core.StringLiteral(token.Value)
		p.advance()
	case core.TokenNumber:
		number, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			number = 0
		}
		value = core.NumberLiteral(number)
		p.advance()
	case core.TokenIdent:
		switch token.Value {
		case "true":
			value = core.BoolLiteral(true)
		case "false":
			value = core.BoolLiteral(false)
		default:
			value = core.StringLiteral(token.Value)
		}
		p.advance()
	default:
		p.report("PARSE_INVALID_VALUE", fmt.Sprintf("Invalid field value %q", token.Value), token)
		p.advance()
		value = core.StringLiteral("")
	}
	return core.FieldNode{Key: start.Value, Value: value, Loc: p.span(start)}
}

// #region Expressions

func (p *parser) parseExpressionList() []core.Expression {
	if p.current().Type == core.TokenLBrace {
		p.advance()
		expressions := make([]core.Expression, 0)
		for !p.isEnd() && p.current().Type != core.TokenRBrace {
			if p.current().Type == core.TokenComma {
				p.advance()
				continue
			}
			expressions = append(expressions, p.parseExpression())
		}
		p.expect(core.TokenRBrace, "PARSE_EXPECTED_RBRACE", "Expected \"}\" to close expression block")
		return expressions
	}
	return []core.Expression{p.parseExpression()}
}

func (p *parser) parseExpression() core.Expression {
	return p.parseOr()
}

func (p *parser) parseOr() core.Expression {
	left := p.parseAnd()
	for p.current().Type == core.TokenOr || p.check("or") {
		operator := p.advance()
		right := p.parseAnd()
		left = p.binary(left, right, core.OpOr, operator)
	}
	return left
}

func (p *parser) parseAnd() core.Expression {
	left := p.parseComparison()
	for p.current().Type == core.TokenAnd || p.check("and") {
		operator := p.advance()
		right := p.parseComparison()
		left = p.binary(left, right, core.OpAnd, operator)
	}
	return left
}

func (p *parser) parseComparison() core.Expression {
	left := p.parseUnary()
	token := p.current()
	var operator core.BinaryOperator
	switch token.Type {
	case core.TokenEQ:
		operator = core.OpEq
	case core.TokenNEQ:
		operator = core.OpNe
	case core.TokenGT:
		operator = core.OpGt
	case core.TokenLT:
		operator = core.OpLt
	case core.TokenGTE:
		operator = core.OpGte
	case core.TokenLTE:
		operator = core.OpLte
	default:
		if token.Type == core.TokenIdent && token.Value == "in" {
			operator = core.OpIn
		}
	}
	if operator == "" {
		return left
	}
	p.advance()
	right := p.parseUnary()
	return p.binary(left, right, operator, token)
}

func (p *parser) parseUnary() core.Expression {
	if p.current().Type == core.TokenNot || p.check("not") {
		operator := p.advance()
		operand := p.parseUnary()
		return &core.UnaryExpr{Operator: "!", Operand: operand, Loc: p.span(operator)}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() core.Expression {
	token := p.current()
	if token.Type == core.TokenString || token.Type == core.TokenNumber {
		p.advance()
		var value core.Literal
		if token.Type == core.TokenNumber {
			number, err := strconv.ParseFloat(token.Value, 64)
			if err != nil {
				number = 0
			}
			value = core.NumberLiteral(number)
		} else {
			value = core.StringLiteral(token.Value)
		}
		return &core.LiteralExpr{Value: value, Loc: token.Loc}
	}
	if token.Type == core.TokenIdent {
		p.advance()
		if token.Value == "true" {
			return &core.LiteralExpr{Value: core.BoolLiteral(true), Loc: token.Loc}
		}
		if token.Value == "false" {
			return &core.LiteralExpr{Value: core.BoolLiteral(false), Loc: token.Loc}
		}
		parts := []string{token.Value}
		for p.current().Type == core.TokenDot {
			p.advance()
			part := p.current()
			if part.Type != core.TokenIdent {
				p.report("PARSE_EXPECTED_IDENTIFIER", "Expected an identifier after \".\"", part)
				break
			}
			parts = append(parts, p.advance().Value)
		}
		return &core.PathExpr{Parts: parts, Loc: p.span(token)}
	}
	if token.Type == core.TokenLParen {
		p.advance()
		expression := p.parseExpression()
		p.expect(core.TokenRParen, "PARSE_EXPECTED_RPAREN", "Expected \")\" to close expression")
		return expression
	}
	p.skipUnexpected("PARSE_EXPECTED_EXPRESSION", fmt.Sprintf("Expected an expression, found %q", token.Value))
	return &core.LiteralExpr{Value: core.StringLiteral(""), Loc: token.Loc}
}

func (p *parser) binary(left, right core.Expression, operator core.BinaryOperator, operatorToken core.Token) core.Expression {
	endOffset := operatorToken.Loc.EndOffset
	if right.NodeLocation().EndOffset > endOffset {
		endOffset = right.NodeLocation().EndOffset
	}
	return &core.BinaryExpr{
		Operator: operator,
		Left:     left,
		Right:    right,
		Loc: core.SourceLocation{
			Line:      left.NodeLocation().Line,
			Column:    left.NodeLocation().Column,
			Offset:    left.NodeLocation().Offset,
			EndOffset: endOffset,
		},
	}
}

// #endregion

// Parse tokenizes and parses AIL source text into a source file AST.
func Parse(source, file string) core.ParseResult {
	lexed := Tokenize(source, file)
	p := &parser{
		tokens:      lexed.Tokens,
		diagnostics: lexed.Diagnostics,
		file:        file,
	}
	ast := p.parseFile()
	return core.ParseResult{AST: ast, Diagnostics: p.diagnostics}
}
