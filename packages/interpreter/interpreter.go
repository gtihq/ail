// Package interpreter reduces a parsed AST into a deterministic intermediate
// representation. It resolves metadata, scopes, priorities, inheritance and
// dependencies without executing any instruction embedded in AIL files.
package interpreter

import (
	"fmt"
	"sort"

	"github.com/gtihq/ail/packages/core"
)

// DefaultVersion is used when no version is declared in META.
const DefaultVersion = "0.3"

// InterpretResult pairs normalized IR with interpreter diagnostics.
type InterpretResult struct {
	IR          *core.IR
	Diagnostics []core.Diagnostic
}

var priorities = map[string]core.Priority{
	"CRITICAL": core.PriorityCritical,
	"HIGH":     core.PriorityHigh,
	"MEDIUM":   core.PriorityMedium,
	"LOW":      core.PriorityLow,
	"INFO":     core.PriorityInfo,
}

// #region Parsed rule accumulation

type parsedRule struct {
	id          string
	priority    core.Priority
	hasPriority bool
	scope       string
	hasScope    bool
	context     map[string]core.Literal
	hasContext  bool
	when        []string
	hasWhen     bool
	do          []string
	hasDo       bool
	never       []string
	hasNever    bool
	verify      []string
	hasVerify   bool
	fallback    []string
	hasFallback bool
	requires    []string
	hasRequires bool
	inherit     []string
	assert      []string
	hasAssert   bool
	execution   string
	hasExec     bool
	source      core.SourceRef
	loc         core.SourceLocation
}

func newParsedRule() *parsedRule {
	return &parsedRule{context: make(map[string]core.Literal)}
}

func extractRule(rule *core.RuleNode) *parsedRule {
	parsed := newParsedRule()
	parsed.id = rule.ID
	for _, field := range rule.Fields {
		switch value := field.(type) {
		case *core.PriorityField:
			parsed.priority = value.Value
			parsed.hasPriority = true
		case *core.ScopeField:
			parsed.scope = value.Value
			parsed.hasScope = true
		case *core.ContextField:
			for _, item := range value.Fields {
				parsed.context[item.Key] = item.Value
			}
			parsed.hasContext = true
		case *core.WhenField:
			parsed.when = core.ExpressionTexts(value.Expressions)
			parsed.hasWhen = true
		case *core.AssertField:
			parsed.assert = core.ExpressionTexts(value.Expressions)
			parsed.hasAssert = true
		case *core.InstructionField:
			terms := make([]string, 0, len(value.Terms))
			for _, term := range value.Terms {
				terms = append(terms, term.Text)
			}
			switch value.Kind {
			case "DO":
				parsed.do = terms
				parsed.hasDo = true
			case "NEVER":
				parsed.never = terms
				parsed.hasNever = true
			case "VERIFY":
				parsed.verify = terms
				parsed.hasVerify = true
			case "FALLBACK":
				parsed.fallback = terms
				parsed.hasFallback = true
			}
		case *core.RefsField:
			if value.Kind == "REQUIRES" {
				parsed.requires = value.IDs
				parsed.hasRequires = true
			} else {
				parsed.inherit = value.IDs
			}
		case *core.ExecutionField:
			parsed.execution = value.Value
			parsed.hasExec = true
		}
	}
	return parsed
}

// #endregion

// #region Resolution

type resolver struct {
	byID           map[string]*parsedRule
	resolved       map[string]*core.IRRule
	visiting       map[string]bool
	diagnostics    []core.Diagnostic
	defaultPriority core.Priority
	defaultScope   string
	file           string
}

func (r *resolver) report(code string, severity core.Severity, message string, loc core.SourceLocation) {
	r.diagnostics = append(r.diagnostics, core.MakeDiagnostic(
		code,
		severity,
		core.KindSemantic,
		message,
		r.file,
		loc,
	))
}

func (r *resolver) resolve(parsed *parsedRule) (*core.IRRule, bool) {
	if existing, ok := r.resolved[parsed.id]; ok {
		return existing, false
	}
	if r.visiting[parsed.id] {
		r.report("INT_INHERITANCE_CYCLE", core.SeverityError,
			fmt.Sprintf("Inheritance cycle detected at rule %q", parsed.id), parsed.loc)
		return nil, true
	}
	r.visiting[parsed.id] = true
	defer delete(r.visiting, parsed.id)

	base := &core.IRRule{
		ID:            parsed.id,
		Priority:      r.defaultPriority,
		Scope:         r.defaultScope,
		Context:       make(map[string]core.Literal),
		When:          []string{},
		Do:            []string{},
		Never:         []string{},
		Verify:        []string{},
		Fallback:      []string{},
		Requires:      []string{},
		Inherit:       append([]string{}, parsed.inherit...),
		Assert:        []string{},
		Execution:     "",
		Source:        parsed.source,
		InheritedFrom: []string{},
	}

	for _, parentID := range parsed.inherit {
		parent, ok := r.byID[parentID]
		if !ok {
			r.report("INT_UNRESOLVED_INHERIT", core.SeverityWarning,
				fmt.Sprintf("Inherited rule %q is undefined", parentID), parsed.loc)
			continue
		}
		inherited, cycle := r.resolve(parent)
		if cycle {
			continue
		}
		inheritFields(base, parsed, inherited)
		base.InheritedFrom = append(base.InheritedFrom, parentID)
		base.InheritedFrom = appendUnique(base.InheritedFrom, inherited.InheritedFrom...)
	}

	applyOwnFields(base, parsed)
	r.resolved[parsed.id] = base
	return base, false
}

func inheritFields(base *core.IRRule, child *parsedRule, parent *core.IRRule) {
	if !child.hasPriority {
		base.Priority = parent.Priority
	}
	if !child.hasScope {
		base.Scope = parent.Scope
	}
	if !child.hasContext {
		base.Context = cloneContext(parent.Context)
	}
	if !child.hasWhen {
		base.When = append([]string{}, parent.When...)
	}
	if !child.hasDo {
		base.Do = append([]string{}, parent.Do...)
	}
	if !child.hasNever {
		base.Never = append([]string{}, parent.Never...)
	}
	if !child.hasVerify {
		base.Verify = append([]string{}, parent.Verify...)
	}
	if !child.hasFallback {
		base.Fallback = append([]string{}, parent.Fallback...)
	}
	if !child.hasRequires {
		base.Requires = append([]string{}, parent.Requires...)
	}
	if !child.hasAssert {
		base.Assert = append([]string{}, parent.Assert...)
	}
	if !child.hasExec {
		base.Execution = parent.Execution
	}
}

func applyOwnFields(base *core.IRRule, parsed *parsedRule) {
	if parsed.hasPriority {
		base.Priority = parsed.priority
	}
	if parsed.hasScope {
		base.Scope = parsed.scope
	}
	if parsed.hasContext {
		base.Context = cloneContext(parsed.context)
	}
	if parsed.hasWhen {
		base.When = append([]string{}, parsed.when...)
	}
	if parsed.hasDo {
		base.Do = append([]string{}, parsed.do...)
	}
	if parsed.hasNever {
		base.Never = append([]string{}, parsed.never...)
	}
	if parsed.hasVerify {
		base.Verify = append([]string{}, parsed.verify...)
	}
	if parsed.hasFallback {
		base.Fallback = append([]string{}, parsed.fallback...)
	}
	if parsed.hasRequires {
		base.Requires = append([]string{}, parsed.requires...)
	}
	if parsed.hasAssert {
		base.Assert = append([]string{}, parsed.assert...)
	}
	if parsed.hasExec {
		base.Execution = parsed.execution
	}
}

func cloneContext(source map[string]core.Literal) map[string]core.Literal {
	clone := make(map[string]core.Literal, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func appendUnique(target []string, values ...string) []string {
	for _, value := range values {
		found := false
		for _, existing := range target {
			if existing == value {
				found = true
				break
			}
		}
		if !found {
			target = append(target, value)
		}
	}
	return target
}

// #endregion

// Interpret converts a parse result into deterministic IR.
func Interpret(parsed core.ParseResult) InterpretResult {
	file := ""
	if parsed.AST != nil {
		file = parsed.AST.File
	}
	version := DefaultVersion
	metaID := ""
	defaultPriority := core.PriorityMedium
	defaultScope := ""
	if parsed.AST != nil && parsed.AST.Meta != nil {
		for _, field := range parsed.AST.Meta.Fields {
			switch field.Key {
			case "id":
				metaID = field.Value.String()
			case "version":
				version = field.Value.String()
			case "priority":
				if priority, ok := priorities[field.Value.String()]; ok {
					defaultPriority = priority
				}
			case "scope":
				defaultScope = field.Value.String()
			}
		}
	}

	parsedRules := make([]*parsedRule, 0)
	rules := []*core.RuleNode{}
	if parsed.AST != nil {
		rules = parsed.AST.Rules
	}
	for index, rule := range rules {
		item := extractRule(rule)
		if item.id == "" {
			if len(rules) == 1 && metaID != "" {
				item.id = metaID
			} else {
				item.id = fmt.Sprintf("R%03d", index+1)
			}
		}
		item.source = core.SourceRef{File: file, Line: rule.Loc.Line, Column: rule.Loc.Column}
		item.loc = rule.Loc
		parsedRules = append(parsedRules, item)
	}

	diagnostics := append([]core.Diagnostic{}, parsed.Diagnostics...)
	byID := make(map[string]*parsedRule)
	for _, item := range parsedRules {
		if _, exists := byID[item.id]; exists {
			diagnostics = append(diagnostics, core.MakeDiagnostic(
				"INT_DUPLICATE_RULE", core.SeverityError, core.KindSemantic,
				fmt.Sprintf("Duplicate rule id %q", item.id), file, item.loc,
			))
			continue
		}
		byID[item.id] = item
	}

	r := &resolver{
		byID:            byID,
		resolved:        make(map[string]*core.IRRule),
		visiting:        make(map[string]bool),
		diagnostics:     []core.Diagnostic{},
		defaultPriority: defaultPriority,
		defaultScope:   defaultScope,
		file:            file,
	}
	for _, item := range parsedRules {
		r.resolve(item)
	}
	diagnostics = append(diagnostics, r.diagnostics...)

	ids := make([]string, 0, len(r.resolved))
	for id := range r.resolved {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		rule := r.resolved[id]
		for _, dependency := range rule.Requires {
			if _, ok := r.resolved[dependency]; !ok {
				diagnostics = append(diagnostics, core.MakeDiagnostic(
					"INT_UNRESOLVED_REFERENCE", core.SeverityWarning, core.KindSemantic,
					fmt.Sprintf("Rule %q requires undefined rule %q", id, dependency), file, rule.Source.ToLocation(),
				))
			}
		}
	}

	ordered := append([]*parsedRule{}, parsedRules...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].id < ordered[j].id
	})
	irRules := make([]core.IRRule, 0, len(ordered))
	for _, item := range ordered {
		if resolvedRule, ok := r.resolved[item.id]; ok {
			irRules = append(irRules, *resolvedRule)
		}
	}

	provenance := make([]core.ProvenanceEntry, 0, 1)
	if file != "" {
		provenance = append(provenance, core.ProvenanceEntry{File: file, Rules: append([]string{}, ids...)})
	}

	return InterpretResult{
		IR: &core.IR{
			Kind:       "IR",
			Version:    version,
			Rules:      irRules,
			Provenance: provenance,
		},
		Diagnostics: diagnostics,
	}
}
