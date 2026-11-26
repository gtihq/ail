// Package core defines the shared data model for the AIL toolchain.
// Every node carries source provenance and is plain data, never executable code.
package core

// #region Source locations and diagnostics

// SourceLocation is a position range inside a source file. Line and column are 1-based, offsets are 0-based.
type SourceLocation struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	Offset    int `json:"offset"`
	EndOffset int `json:"endOffset"`
}

// Severity is a diagnostic severity level.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
	SeverityHint    Severity = "HINT"
)

// DiagnosticKind separates correctness findings from style preferences.
type DiagnosticKind string

const (
	KindSyntax   DiagnosticKind = "syntax"
	KindSemantic DiagnosticKind = "semantic"
	KindStyle    DiagnosticKind = "style"
)

// Diagnostic is a machine-readable finding with full provenance.
type Diagnostic struct {
	Code     string         `json:"code"`
	Severity Severity       `json:"severity"`
	Kind     DiagnosticKind `json:"kind"`
	Message  string         `json:"message"`
	File     string         `json:"file"`
	Loc      SourceLocation `json:"loc"`
}

// #endregion

// #region Tokens and literals

// TokenType is a lexical token category. Keywords stay IDENT and are resolved by the parser.
type TokenType string

const (
	TokenIdent  TokenType = "IDENT"
	TokenString TokenType = "STRING"
	TokenNumber TokenType = "NUMBER"
	TokenLBrace TokenType = "LBRACE"
	TokenRBrace TokenType = "RBRACE"
	TokenLParen TokenType = "LPAREN"
	TokenRParen TokenType = "RPAREN"
	TokenColon  TokenType = "COLON"
	TokenComma  TokenType = "COMMA"
	TokenDot    TokenType = "DOT"
	TokenEQ     TokenType = "EQ"
	TokenNEQ    TokenType = "NEQ"
	TokenGT     TokenType = "GT"
	TokenLT     TokenType = "LT"
	TokenGTE    TokenType = "GTE"
	TokenLTE    TokenType = "LTE"
	TokenAnd    TokenType = "AND"
	TokenOr     TokenType = "OR"
	TokenNot    TokenType = "NOT"
	TokenEOF    TokenType = "EOF"
)

// Token is a lexical token with its source range.
type Token struct {
	Type  TokenType      `json:"type"`
	Value string         `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

// LiteralKind identifies the primitive type of a Literal.
type LiteralKind string

const (
	LiteralString LiteralKind = "string"
	LiteralNumber LiteralKind = "number"
	LiteralBool   LiteralKind = "bool"
)

// Literal is a typed primitive value used in META and CONTEXT fields.
type Literal struct {
	Kind   LiteralKind `json:"kind"`
	Text   string      `json:"text"`
	Number float64     `json:"number,omitempty"`
	Bool   bool        `json:"bool,omitempty"`
}

// StringLiteral builds a string Literal.
func StringLiteral(value string) Literal {
	return Literal{Kind: LiteralString, Text: value}
}

// NumberLiteral builds a numeric Literal.
func NumberLiteral(value float64) Literal {
	return Literal{Kind: LiteralNumber, Number: value}
}

// BoolLiteral builds a boolean Literal.
func BoolLiteral(value bool) Literal {
	return Literal{Kind: LiteralBool, Bool: value}
}

// String returns the canonical text form of a Literal.
func (l Literal) String() string {
	switch l.Kind {
	case LiteralNumber:
		return formatNumber(l.Number)
	case LiteralBool:
		if l.Bool {
			return "true"
		}
		return "false"
	default:
		return l.Text
	}
}

// FieldNode is a key/value pair inside a META or CONTEXT block.
type FieldNode struct {
	Key   string         `json:"key"`
	Value Literal        `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

// #endregion

// #region Expressions

// BinaryOperator is a comparison or logical operator in WHEN and ASSERT expressions.
type BinaryOperator string

const (
	OpAnd BinaryOperator = "&&"
	OpOr  BinaryOperator = "||"
	OpEq  BinaryOperator = "=="
	OpNe  BinaryOperator = "!="
	OpGt  BinaryOperator = ">"
	OpLt  BinaryOperator = "<"
	OpGte BinaryOperator = ">="
	OpLte BinaryOperator = "<="
	OpIn  BinaryOperator = "in"
)

// Expression is a declarative expression node. It is data only and is never evaluated as code.
type Expression interface {
	expressionNode()
	NodeLocation() SourceLocation
}

// LiteralExpr is a literal value inside an expression.
type LiteralExpr struct {
	Value Literal        `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

func (*LiteralExpr) expressionNode() {}
func (e *LiteralExpr) NodeLocation() SourceLocation { return e.Loc }

// PathExpr is a dotted identifier path such as file.language.
type PathExpr struct {
	Parts []string       `json:"parts"`
	Loc   SourceLocation `json:"loc"`
}

func (*PathExpr) expressionNode() {}
func (e *PathExpr) NodeLocation() SourceLocation { return e.Loc }

// UnaryExpr is a logical negation.
type UnaryExpr struct {
	Operator string         `json:"operator"`
	Operand  Expression     `json:"operand"`
	Loc      SourceLocation `json:"loc"`
}

func (*UnaryExpr) expressionNode() {}
func (e *UnaryExpr) NodeLocation() SourceLocation { return e.Loc }

// BinaryExpr combines two expressions with a logical or comparison operator.
type BinaryExpr struct {
	Operator BinaryOperator `json:"operator"`
	Left     Expression     `json:"left"`
	Right    Expression     `json:"right"`
	Loc      SourceLocation `json:"loc"`
}

func (*BinaryExpr) expressionNode() {}
func (e *BinaryExpr) NodeLocation() SourceLocation { return e.Loc }

// #endregion

// #region Instructions and AST

// InstructionTerm is one declarative term inside DO, NEVER, VERIFY or FALLBACK.
type InstructionTerm struct {
	Kind string         `json:"kind"`
	Text string         `json:"text"`
	Loc  SourceLocation `json:"loc"`
}

// Priority orders rules from most to least urgent.
type Priority string

const (
	PriorityCritical Priority = "CRITICAL"
	PriorityHigh     Priority = "HIGH"
	PriorityMedium   Priority = "MEDIUM"
	PriorityLow      Priority = "LOW"
	PriorityInfo     Priority = "INFO"
)

// RuleField is one field inside a RULE block.
type RuleField interface {
	ruleFieldNode()
	NodeLocation() SourceLocation
}

// PriorityField declares a rule priority.
type PriorityField struct {
	Value Priority       `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

func (*PriorityField) ruleFieldNode() {}
func (f *PriorityField) NodeLocation() SourceLocation { return f.Loc }

// ScopeField declares a glob scope.
type ScopeField struct {
	Value string         `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

func (*ScopeField) ruleFieldNode() {}
func (f *ScopeField) NodeLocation() SourceLocation { return f.Loc }

// ContextField declares key/value context data.
type ContextField struct {
	Fields []FieldNode    `json:"fields"`
	Loc    SourceLocation `json:"loc"`
}

func (*ContextField) ruleFieldNode() {}
func (f *ContextField) NodeLocation() SourceLocation { return f.Loc }

// WhenField declares conditions.
type WhenField struct {
	Expressions []Expression   `json:"expressions"`
	Loc         SourceLocation `json:"loc"`
}

func (*WhenField) ruleFieldNode() {}
func (f *WhenField) NodeLocation() SourceLocation { return f.Loc }

// AssertField declares verification conditions.
type AssertField struct {
	Expressions []Expression   `json:"expressions"`
	Loc         SourceLocation `json:"loc"`
}

func (*AssertField) ruleFieldNode() {}
func (f *AssertField) NodeLocation() SourceLocation { return f.Loc }

// InstructionField is a DO, NEVER, VERIFY or FALLBACK block.
type InstructionField struct {
	Kind  string           `json:"kind"`
	Terms []InstructionTerm `json:"terms"`
	Loc   SourceLocation   `json:"loc"`
}

func (*InstructionField) ruleFieldNode() {}
func (f *InstructionField) NodeLocation() SourceLocation { return f.Loc }

// RefsField is a REQUIRES or INHERIT block.
type RefsField struct {
	Kind string         `json:"kind"`
	IDs  []string       `json:"ids"`
	Loc  SourceLocation `json:"loc"`
}

func (*RefsField) ruleFieldNode() {}
func (f *RefsField) NodeLocation() SourceLocation { return f.Loc }

// ExecutionField declares an execution policy token.
type ExecutionField struct {
	Value string         `json:"value"`
	Loc   SourceLocation `json:"loc"`
}

func (*ExecutionField) ruleFieldNode() {}
func (f *ExecutionField) NodeLocation() SourceLocation { return f.Loc }

// MetaNode is a parsed META declaration.
type MetaNode struct {
	Loc    SourceLocation `json:"loc"`
	Fields []FieldNode    `json:"fields"`
}

// RuleNode is a parsed rule, explicit or implicit.
type RuleNode struct {
	Loc    SourceLocation `json:"loc"`
	ID     string         `json:"id"`
	Fields []RuleField    `json:"fields"`
}

// SourceFile is a parsed AIL file in standalone or aggregated mode.
type SourceFile struct {
	File  string         `json:"file"`
	Loc   SourceLocation `json:"loc"`
	Mode  string         `json:"mode"`
	Meta  *MetaNode      `json:"meta"`
	Rules []*RuleNode    `json:"rules"`
}

// ParseResult pairs an AST with its syntax diagnostics.
type ParseResult struct {
	AST         *SourceFile   `json:"ast"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// #endregion

// #region Intermediate representation

// IRRule is a fully resolved rule ready for rendering or validation.
type IRRule struct {
	ID            string             `json:"id"`
	Priority      Priority           `json:"priority"`
	Scope         string             `json:"scope"`
	Context       map[string]Literal `json:"context"`
	When          []string           `json:"when"`
	Do            []string           `json:"do"`
	Never         []string           `json:"never"`
	Verify        []string           `json:"verify"`
	Fallback      []string           `json:"fallback"`
	Requires      []string           `json:"requires"`
	Inherit       []string           `json:"inherit"`
	Assert        []string           `json:"assert"`
	Execution     string             `json:"execution"`
	Source        SourceRef          `json:"source"`
	InheritedFrom []string           `json:"inheritedFrom"`
}

// SourceRef records where an IR rule came from.
type SourceRef struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// ToLocation converts a source reference into a minimal source location.
func (ref SourceRef) ToLocation() SourceLocation {
	return SourceLocation{Line: ref.Line, Column: ref.Column}
}

// ProvenanceEntry maps a source file to the rule IDs it produced.
type ProvenanceEntry struct {
	File  string   `json:"file"`
	Rules []string `json:"rules"`
}

// IR is the deterministic normalized representation produced by the interpreter.
type IR struct {
	Kind       string            `json:"kind"`
	Version    string            `json:"version"`
	Rules      []IRRule          `json:"rules"`
	Provenance []ProvenanceEntry `json:"provenance"`
}

// #endregion

// #region Renderer and analyzer results

// HostFormat identifies a host instruction convention.
type HostFormat string

const (
	HostCursor    HostFormat = "cursor"
	HostOpenCode  HostFormat = "opencode"
	HostAgentsMD  HostFormat = "agents-md"
	HostMarkdown  HostFormat = "markdown"
)

// InjectionMode selects a rendering strategy with identical semantic intent.
type InjectionMode string

const (
	ModeInline  InjectionMode = "INLINE"
	ModeFile    InjectionMode = "FILE"
	ModeCompact InjectionMode = "COMPACT"
	ModeVerbose InjectionMode = "VERBOSE"
)

// RenderOptions controls host format, injection mode and delimiter generation.
type RenderOptions struct {
	Host              HostFormat    `json:"host"`
	Mode              InjectionMode `json:"mode"`
	IncludeDelimiters *bool         `json:"includeDelimiters,omitempty"`
	IncludeProvenance bool          `json:"includeProvenance"`
}

// RenderResult is rendered text plus sanitization diagnostics.
type RenderResult struct {
	Text        string       `json:"text"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Summary counts diagnostics by severity.
type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
	Hints    int `json:"hints"`
}

// LintCategory identifies a linter diagnostic class.
type LintCategory string

const (
	LintDuplicateRule          LintCategory = "duplicate-rule"
	LintRedundantInstruction   LintCategory = "redundant-instruction"
	LintContradictory          LintCategory = "contradictory-instruction"
	LintAmbiguousCondition     LintCategory = "ambiguous-condition"
	LintUnnecessaryNesting     LintCategory = "unnecessary-nesting"
	LintUnusedDeclaration      LintCategory = "unused-declaration"
	LintUnreachableRule        LintCategory = "unreachable-rule"
	LintMissingFallback        LintCategory = "missing-fallback"
	LintOverlyBroadScope       LintCategory = "overly-broad-scope"
	LintExcessiveComplexity    LintCategory = "excessive-complexity"
	LintInconsistentNaming     LintCategory = "inconsistent-naming"
	LintUnnecessaryVerbosity   LintCategory = "unnecessary-verbosity"
	LintIneffectiveInstruction LintCategory = "ineffective-instruction"
)

// LintConfig holds per-category severity overrides. Missing categories use built-in defaults.
type LintConfig map[LintCategory]Severity

// AnalyzerResult is the result of a linter pass.
type AnalyzerResult struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Summary     Summary      `json:"summary"`
}

// ValidationResult is the machine-readable validator output.
type ValidationResult struct {
	Valid   bool         `json:"valid"`
	Errors  []Diagnostic `json:"errors"`
	Warning []Diagnostic `json:"warnings"`
	Summary struct {
		Rules    int `json:"rules"`
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"summary"`
}

// #endregion
