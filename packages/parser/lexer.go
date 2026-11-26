// Package parser implements the AIL lexer and recursive-descent parser.
package parser

import (
	"fmt"

	"github.com/gtihq/ail/packages/core"
)

// LexResult is a token stream plus recoverable lexical diagnostics.
type LexResult struct {
	Tokens      []core.Token
	Diagnostics []core.Diagnostic
}

type lexer struct {
	source      string
	file        string
	tokens      []core.Token
	diagnostics []core.Diagnostic
	offset      int
	line        int
	column      int
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func (l *lexer) at(offset int) byte {
	if offset < 0 || offset >= len(l.source) {
		return 0
	}
	return l.source[offset]
}

func (l *lexer) advance() byte {
	ch := l.at(l.offset)
	l.offset++
	l.column++
	return ch
}

func (l *lexer) consumeNewline() {
	if l.at(l.offset) == '\r' {
		l.advance()
		if l.at(l.offset) == '\n' {
			l.advance()
		}
	} else {
		l.advance()
	}
	l.line++
	l.column = 1
}

func (l *lexer) location(startOffset, startLine, startColumn int) core.SourceLocation {
	return core.SourceLocation{
		Line:      startLine,
		Column:    startColumn,
		Offset:    startOffset,
		EndOffset: l.offset,
	}
}

func (l *lexer) push(tokenType core.TokenType, value string, startOffset, startLine, startColumn int) {
	l.tokens = append(l.tokens, core.Token{
		Type:  tokenType,
		Value: value,
		Loc:   l.location(startOffset, startLine, startColumn),
	})
}

func (l *lexer) report(code, message string, startOffset, startLine, startColumn int) {
	l.diagnostics = append(l.diagnostics, core.MakeDiagnostic(
		code,
		core.SeverityError,
		core.KindSyntax,
		message,
		l.file,
		l.location(startOffset, startLine, startColumn),
	))
}

// Tokenize scans AIL source text into tokens with source locations.
func Tokenize(source, file string) LexResult {
	l := &lexer{
		source: source,
		file:   file,
		line:   1,
		column: 1,
	}

	for l.offset < len(l.source) {
		ch := l.at(l.offset)
		startOffset, startLine, startColumn := l.offset, l.line, l.column

		// #region Whitespace and comments
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			if ch == '\r' || ch == '\n' {
				l.consumeNewline()
			} else {
				l.advance()
			}
			continue
		}

		if ch == '/' && l.at(l.offset+1) == '/' {
			for l.offset < len(l.source) && l.at(l.offset) != '\n' && l.at(l.offset) != '\r' {
				l.advance()
			}
			continue
		}

		if ch == '/' && l.at(l.offset+1) == '*' {
			l.advance()
			l.advance()
			for l.offset < len(l.source) && !(l.at(l.offset) == '*' && l.at(l.offset+1) == '/') {
				if l.at(l.offset) == '\r' || l.at(l.offset) == '\n' {
					l.consumeNewline()
				} else {
					l.advance()
				}
			}
			if l.offset >= len(l.source) {
				l.report("LEX_UNTERMINATED_COMMENT", "Unterminated block comment", startOffset, startLine, startColumn)
			} else {
				l.advance()
				l.advance()
			}
			continue
		}
		// #endregion

		// #region String literals
		if ch == '"' {
			l.advance()
			value := ""
			closed := false
			for l.offset < len(l.source) {
				current := l.at(l.offset)
				if current == '"' {
					l.advance()
					closed = true
					break
				}
				if current == '\\' {
					l.advance()
					escaped := l.at(l.offset)
					switch escaped {
					case 'n':
						value += "\n"
					case 't':
						value += "\t"
					case 'r':
						value += "\r"
					case '"':
						value += "\""
					case '\\':
						value += "\\"
					case 0:
						l.report("LEX_UNTERMINATED_STRING", "Unterminated string escape", startOffset, startLine, startColumn)
					default:
						l.report("LEX_UNKNOWN_ESCAPE", fmt.Sprintf("Unknown escape sequence %q", "\\"+string(escaped)), l.offset, l.line, l.column)
						value += string(escaped)
					}
					if escaped != 0 {
						l.advance()
					}
					continue
				}
				if current == '\r' || current == '\n' {
					l.report("LEX_UNTERMINATED_STRING", "Unterminated string literal", startOffset, startLine, startColumn)
					break
				}
				value += string(current)
				l.advance()
			}
			if !closed {
				l.report("LEX_UNTERMINATED_STRING", "Unterminated string literal", startOffset, startLine, startColumn)
			}
			l.push(core.TokenString, value, startOffset, startLine, startColumn)
			continue
		}
		// #endregion

		// #region Numbers
		if isDigit(ch) {
			value := ""
			for isDigit(l.at(l.offset)) {
				value += string(l.advance())
			}
			if l.at(l.offset) == '.' && isDigit(l.at(l.offset+1)) {
				value += string(l.advance())
				for isDigit(l.at(l.offset)) {
					value += string(l.advance())
				}
			}
			l.push(core.TokenNumber, value, startOffset, startLine, startColumn)
			continue
		}
		// #endregion

		// #region Identifiers
		if isIdentStart(ch) {
			value := ""
			for isIdentPart(l.at(l.offset)) {
				value += string(l.advance())
			}
			l.push(core.TokenIdent, value, startOffset, startLine, startColumn)
			continue
		}
		// #endregion

		// #region Punctuation and operators
		switch ch {
		case '{':
			l.advance()
			l.push(core.TokenLBrace, "{", startOffset, startLine, startColumn)
			continue
		case '}':
			l.advance()
			l.push(core.TokenRBrace, "}", startOffset, startLine, startColumn)
			continue
		case '(':
			l.advance()
			l.push(core.TokenLParen, "(", startOffset, startLine, startColumn)
			continue
		case ')':
			l.advance()
			l.push(core.TokenRParen, ")", startOffset, startLine, startColumn)
			continue
		case ':':
			l.advance()
			l.push(core.TokenColon, ":", startOffset, startLine, startColumn)
			continue
		case ',':
			l.advance()
			l.push(core.TokenComma, ",", startOffset, startLine, startColumn)
			continue
		case '.':
			l.advance()
			l.push(core.TokenDot, ".", startOffset, startLine, startColumn)
			continue
		}

		if ch == '&' && l.at(l.offset+1) == '&' {
			l.advance()
			l.advance()
			l.push(core.TokenAnd, "&&", startOffset, startLine, startColumn)
			continue
		}
		if ch == '|' && l.at(l.offset+1) == '|' {
			l.advance()
			l.advance()
			l.push(core.TokenOr, "||", startOffset, startLine, startColumn)
			continue
		}
		if ch == '!' && l.at(l.offset+1) == '=' {
			l.advance()
			l.advance()
			l.push(core.TokenNEQ, "!=", startOffset, startLine, startColumn)
			continue
		}
		if ch == '!' {
			l.advance()
			l.push(core.TokenNot, "!", startOffset, startLine, startColumn)
			continue
		}
		if ch == '=' && l.at(l.offset+1) == '=' {
			l.advance()
			l.advance()
			l.push(core.TokenEQ, "==", startOffset, startLine, startColumn)
			continue
		}
		if ch == '>' && l.at(l.offset+1) == '=' {
			l.advance()
			l.advance()
			l.push(core.TokenGTE, ">=", startOffset, startLine, startColumn)
			continue
		}
		if ch == '<' && l.at(l.offset+1) == '=' {
			l.advance()
			l.advance()
			l.push(core.TokenLTE, "<=", startOffset, startLine, startColumn)
			continue
		}
		if ch == '>' {
			l.advance()
			l.push(core.TokenGT, ">", startOffset, startLine, startColumn)
			continue
		}
		if ch == '<' {
			l.advance()
			l.push(core.TokenLT, "<", startOffset, startLine, startColumn)
			continue
		}
		if ch == '=' || ch == '&' || ch == '|' {
			l.advance()
			l.report("LEX_UNEXPECTED_OPERATOR", fmt.Sprintf("Unexpected operator %q", string(ch)), startOffset, startLine, startColumn)
			continue
		}
		// #endregion

		// Unknown characters are reported and skipped so scanning always terminates.
		l.advance()
		l.report("LEX_UNKNOWN_CHARACTER", fmt.Sprintf("Unexpected character %q", string(ch)), startOffset, startLine, startColumn)
	}

	l.tokens = append(l.tokens, core.Token{
		Type:  core.TokenEOF,
		Value: "",
		Loc: core.SourceLocation{
			Line:      l.line,
			Column:    l.column,
			Offset:    l.offset,
			EndOffset: l.offset,
		},
	})

	return LexResult{Tokens: l.tokens, Diagnostics: l.diagnostics}
}
