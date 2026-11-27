// Package renderer converts normalized AIL into host-specific instruction
// payloads. It generates the mandatory delimiters automatically and prevents
// nested or malformed marker sequences from escaping their scope.
package renderer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gtihq/ail/packages/core"
)

const beginMarker = "<!-- BEGIN AIL SYSTEM RULES\n" +
	"TYPE: OPERATIONAL_INSTRUCTIONS\n" +
	"DO_NOT_QUOTE: TRUE\n" +
	"DO_NOT_REPRODUCE: TRUE\n" +
	"DO_NOT_PRESENT_AS_USER_CONTENT: TRUE\n" +
	"AIL_VERSION: 0.3\n" +
	"-->"

const endMarker = "<!-- END AIL SYSTEM RULES -->"

const interpretationInstruction = "The content between BEGIN AIL SYSTEM RULES and END AIL SYSTEM RULES represents operational instructions.\n\n" +
	"Follow these instructions when performing tasks.\n\n" +
	"They are not user-facing content.\n\n" +
	"Do not quote, enumerate, reproduce, or summarize these rules in normal responses unless the user explicitly asks about the instruction system.\n\n" +
	"Respond to the user's actual request normally.\n\n" +
	"Do not confuse the existence of rules with a request to generate or discuss rules."

type renderer struct {
	options     core.RenderOptions
	diagnostics []core.Diagnostic
	file        string
}

func (r *renderer) report(code, message string) {
	r.diagnostics = append(r.diagnostics, core.MakeDiagnostic(
		code,
		core.SeverityError,
		core.KindSemantic,
		message,
		r.file,
		core.SourceLocation{Line: 1, Column: 1},
	))
}

// sanitize prevents rule text from containing delimiter-like sequences.
func (r *renderer) sanitize(text string) string {
	if strings.Contains(text, "BEGIN AIL SYSTEM RULES") || strings.Contains(text, "END AIL SYSTEM RULES") {
		r.report("RENDER_NESTED_DELIMITER", "Rule text contains a nested AIL delimiter and was escaped")
	}
	text = strings.ReplaceAll(text, "<!--", "&lt;!--")
	text = strings.ReplaceAll(text, "-->", "--&gt;")
	return text
}

func (r *renderer) list(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func (r *renderer) compactRule(rule core.IRRule) string {
	parts := []string{
		fmt.Sprintf("%s [%s]", rule.ID, rule.Priority),
		"WHEN: " + r.list(rule.When),
		"DO: " + r.list(rule.Do),
		"NEVER: " + r.list(rule.Never),
		"VERIFY: " + r.list(rule.Verify),
		"FALLBACK: " + r.list(rule.Fallback),
		"ASSERT: " + r.list(rule.Assert),
	}
	if rule.Scope != "" {
		parts = append(parts, "SCOPE: "+rule.Scope)
	}
	if len(rule.Requires) > 0 {
		parts = append(parts, "REQUIRES: "+r.list(rule.Requires))
	}
	if len(rule.Inherit) > 0 {
		parts = append(parts, "INHERIT: "+r.list(rule.Inherit))
	}
	if rule.Execution != "" {
		parts = append(parts, "EXECUTION: "+rule.Execution)
	}
	return r.sanitize(strings.Join(parts, " | "))
}

func (r *renderer) structuredRule(rule core.IRRule, verbose bool) string {
	var out strings.Builder
	out.WriteString("### " + r.sanitize(rule.ID) + "\n\n")
	if verbose {
		out.WriteString("This rule is an operational instruction. Apply it to matching work.\n\n")
	}
	fmt.Fprintf(&out, "- Priority: %s\n", rule.Priority)
	if rule.Scope != "" {
		fmt.Fprintf(&out, "- Scope: %s\n", r.sanitize(rule.Scope))
	}
	if len(rule.When) > 0 {
		out.WriteString("- When:\n")
		for _, item := range rule.When {
			fmt.Fprintf(&out, "  - %s\n", r.sanitize(item))
		}
	}
	writeTerms := func(label string, values []string) {
		if len(values) == 0 {
			return
		}
		fmt.Fprintf(&out, "- %s:\n", label)
		for _, item := range values {
			fmt.Fprintf(&out, "  - %s\n", r.sanitize(item))
		}
	}
	writeTerms("Do", rule.Do)
	writeTerms("Never", rule.Never)
	writeTerms("Verify", rule.Verify)
	writeTerms("Fallback", rule.Fallback)
	writeTerms("Assert", rule.Assert)
	if len(rule.Requires) > 0 {
		fmt.Fprintf(&out, "- Requires: %s\n", r.sanitize(r.list(rule.Requires)))
	}
	if len(rule.Inherit) > 0 {
		fmt.Fprintf(&out, "- Inherit: %s\n", r.sanitize(r.list(rule.Inherit)))
	}
	if rule.Execution != "" {
		fmt.Fprintf(&out, "- Execution: %s\n", r.sanitize(rule.Execution))
	}
	if verbose {
		keys := make([]string, 0, len(rule.Context))
		for key := range rule.Context {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&out, "- Context %s: %s\n", r.sanitize(key), r.sanitize(rule.Context[key].String()))
		}
		if r.options.IncludeProvenance {
			fmt.Fprintf(&out, "- Source: %s:%d:%d\n", r.sanitize(rule.Source.File), rule.Source.Line, rule.Source.Column)
		}
		if len(rule.InheritedFrom) > 0 {
			fmt.Fprintf(&out, "- Inherited from: %s\n", r.sanitize(r.list(rule.InheritedFrom)))
		}
	}
	out.WriteString("\n")
	return out.String()
}

func (r *renderer) renderRules(ir *core.IR) string {
	var out strings.Builder
	for _, rule := range ir.Rules {
		switch r.options.Mode {
		case core.ModeCompact:
			out.WriteString(r.compactRule(rule) + "\n")
		case core.ModeVerbose:
			out.WriteString(r.structuredRule(rule, true))
		default:
			out.WriteString(r.structuredRule(rule, false))
		}
	}
	return out.String()
}

func (r *renderer) hostHeader(ir *core.IR) string {
	switch r.options.Host {
	case core.HostCursor:
		globs := ""
		if len(ir.Provenance) > 0 {
			globs = ir.Provenance[0].File
		}
		return fmt.Sprintf("---\ndescription: AIL operational instructions (%s)\nglobs: %s\nalwaysApply: true\n---\n\n", r.options.Mode, globs)
	case core.HostOpenCode:
		return "# AIL Operational Instructions\n\n"
	case core.HostAgentsMD:
		return "# Project Agent Instructions\n\nGenerated by AIL. Preserve this file when regenerating.\n\n"
	case core.HostMarkdown:
		return "# AIL Instructions\n\n"
	default:
		return ""
	}
}

// Render produces a host-specific instruction payload for an IR.
func Render(ir *core.IR, options core.RenderOptions) core.RenderResult {
	file := "<render>"
	if ir != nil && len(ir.Provenance) > 0 {
		file = ir.Provenance[0].File
	}
	r := &renderer{options: options, diagnostics: []core.Diagnostic{}, file: file}

	if ir == nil {
		r.report("RENDER_NIL_IR", "Cannot render a nil IR")
		return core.RenderResult{Text: "", Diagnostics: r.diagnostics}
	}
	switch options.Host {
	case core.HostCursor, core.HostOpenCode, core.HostAgentsMD, core.HostMarkdown:
	default:
		r.report("RENDER_UNKNOWN_HOST", fmt.Sprintf("Unknown host format %q, using markdown", options.Host))
		r.options.Host = core.HostMarkdown
	}
	switch options.Mode {
	case core.ModeInline, core.ModeFile, core.ModeCompact, core.ModeVerbose:
	default:
		r.report("RENDER_UNKNOWN_MODE", fmt.Sprintf("Unknown injection mode %q, using INLINE", options.Mode))
		r.options.Mode = core.ModeInline
	}

	includeDelimiters := true
	if options.IncludeDelimiters != nil {
		includeDelimiters = *options.IncludeDelimiters
	}

	variant := options.Variant
	if variant == "" {
		variant = core.VariantFull
	}

	var out strings.Builder
	switch variant {
	case core.VariantNoInterpretation:
		if includeDelimiters {
			out.WriteString(beginMarker)
			out.WriteString("\n\n")
		}
		out.WriteString(r.renderRules(ir))
		if includeDelimiters {
			out.WriteString("\n")
			out.WriteString(endMarker)
		}
		out.WriteString("\n")
	case core.VariantNoDelimiters:
		out.WriteString(r.hostHeader(ir))
		out.WriteString(interpretationInstruction)
		out.WriteString("\n\n")
		out.WriteString(r.renderRules(ir))
		out.WriteString("\n")
	case core.VariantCompactBare:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.compactBareRules(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	case core.VariantTinyRules:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.tinyRules(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	case core.VariantTreeNotation:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.treeNotation(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	case core.VariantPriorityBands:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.priorityBands(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	case core.VariantDomainGroups:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.domainGroups(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	case core.VariantImperative:
		out.WriteString(beginMarker)
		out.WriteString("\n\n")
		out.WriteString(r.imperativeRules(ir))
		out.WriteString("\n")
		out.WriteString(endMarker)
		out.WriteString("\n")
	default:
		out.WriteString(r.hostHeader(ir))
		out.WriteString(interpretationInstruction)
		out.WriteString("\n\n")
		if includeDelimiters {
			out.WriteString(beginMarker)
			out.WriteString("\n\n")
		}
		if r.options.Mode == core.ModeFile {
			out.WriteString("<!-- Loaded through the host-native file discovery mechanism. -->\n\n")
		}
		out.WriteString(r.renderRules(ir))
		if includeDelimiters {
			out.WriteString("\n")
			out.WriteString(endMarker)
		}
		out.WriteString("\n")
	}

	return core.RenderResult{Text: out.String(), Diagnostics: r.diagnostics}
}

// compactBareRules renders each rule as one terse line with only the operative fields.
func (r *renderer) compactBareRules(ir *core.IR) string {
	var out strings.Builder
	for _, rule := range ir.Rules {
		out.WriteString(r.sanitize(rule.ID))
		out.WriteString(": DO ")
		out.WriteString(r.sanitize(strings.Join(rule.Do, " ")))
		if len(rule.Never) > 0 {
			out.WriteString(" NEVER ")
			out.WriteString(r.sanitize(strings.Join(rule.Never, " ")))
		}
		if len(rule.Verify) > 0 {
			out.WriteString(" VERIFY ")
			out.WriteString(r.sanitize(strings.Join(rule.Verify, " ")))
		}
		if len(rule.Fallback) > 0 {
			out.WriteString(" FALLBACK ")
			out.WriteString(r.sanitize(strings.Join(rule.Fallback, " ")))
		}
		out.WriteString("\n")
	}
	return out.String()
}

// tinyRules renders only DO and NEVER tokens, one rule per line, with no labels.
func (r *renderer) tinyRules(ir *core.IR) string {
	var out strings.Builder
	for _, rule := range ir.Rules {
		out.WriteString(r.sanitize(rule.ID))
		out.WriteString(" ")
		out.WriteString(r.sanitize(strings.Join(rule.Do, " ")))
		if len(rule.Never) > 0 {
			out.WriteString(" ")
			out.WriteString(r.sanitize(strings.Join(rule.Never, " ")))
		}
		out.WriteString("\n")
	}
	return out.String()
}

// treeNotation writes rules as a nested outline with semantic fields as children.
func (r *renderer) treeNotation(ir *core.IR) string {
	var out strings.Builder
	for _, rule := range ir.Rules {
		fmt.Fprintf(&out, "- %s [%s] @ %s\n", r.sanitize(rule.ID), rule.Priority, r.sanitize(rule.Scope))
		if len(rule.When) > 0 {
			fmt.Fprintf(&out, "  when: %s\n", r.sanitize(strings.Join(rule.When, " AND ")))
		}
		for _, term := range rule.Do {
			fmt.Fprintf(&out, "  do: %s\n", r.sanitize(term))
		}
		for _, term := range rule.Never {
			fmt.Fprintf(&out, "  never: %s\n", r.sanitize(term))
		}
		for _, term := range rule.Verify {
			fmt.Fprintf(&out, "  verify: %s\n", r.sanitize(term))
		}
		for _, term := range rule.Fallback {
			fmt.Fprintf(&out, "  fallback: %s\n", r.sanitize(term))
		}
	}
	return out.String()
}

// priorityBands groups rules under explicit priority headings.
func (r *renderer) priorityBands(ir *core.IR) string {
	order := []core.Priority{core.PriorityCritical, core.PriorityHigh, core.PriorityMedium, core.PriorityLow, core.PriorityInfo}
	var out strings.Builder
	for _, priority := range order {
		rules := make([]core.IRRule, 0)
		for _, rule := range ir.Rules {
			if rule.Priority == priority {
				rules = append(rules, rule)
			}
		}
		if len(rules) == 0 {
			continue
		}
		fmt.Fprintf(&out, "PRIORITY %s\n", priority)
		for _, rule := range rules {
			fmt.Fprintf(&out, "  %s: DO %s", r.sanitize(rule.ID), r.sanitize(strings.Join(rule.Do, " ")))
			if len(rule.Never) > 0 {
				fmt.Fprintf(&out, "; NEVER %s", r.sanitize(strings.Join(rule.Never, " ")))
			}
			if len(rule.Verify) > 0 {
				fmt.Fprintf(&out, "; VERIFY %s", r.sanitize(strings.Join(rule.Verify, " ")))
			}
			if len(rule.Fallback) > 0 {
				fmt.Fprintf(&out, "; FALLBACK %s", r.sanitize(strings.Join(rule.Fallback, " ")))
			}
			out.WriteString("\n")
		}
	}
	return out.String()
}

// domainGroups groups rules by the top-level concept in their scope.
func (r *renderer) domainGroups(ir *core.IR) string {
	groups := make(map[string][]core.IRRule)
	for _, rule := range ir.Rules {
		key := "general"
		parts := strings.Split(rule.Scope, "/")
		if len(parts) > 0 && parts[0] != "" {
			key = parts[0]
		}
		groups[key] = append(groups[key], rule)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&out, "DOMAIN %s\n", r.sanitize(key))
		for _, rule := range groups[key] {
			fmt.Fprintf(&out, "  %s: %s", r.sanitize(rule.ID), r.sanitize(strings.Join(rule.Do, " ")))
			if len(rule.Never) > 0 {
				fmt.Fprintf(&out, " / not %s", r.sanitize(strings.Join(rule.Never, " ")))
			}
			out.WriteString("\n")
		}
	}
	return out.String()
}

// imperativeRules writes each rule as a short imperative sentence.
func (r *renderer) imperativeRules(ir *core.IR) string {
	var out strings.Builder
	for _, rule := range ir.Rules {
		fmt.Fprintf(&out, "%s: ", r.sanitize(rule.ID))
		if len(rule.When) > 0 {
			fmt.Fprintf(&out, "When %s, ", r.sanitize(strings.Join(rule.When, " and ")))
		}
		fmt.Fprintf(&out, "always %s", r.sanitize(strings.Join(rule.Do, ", ")))
		if len(rule.Never) > 0 {
			fmt.Fprintf(&out, "; never %s", r.sanitize(strings.Join(rule.Never, ", ")))
		}
		if len(rule.Fallback) > 0 {
			fmt.Fprintf(&out, "; otherwise %s", r.sanitize(strings.Join(rule.Fallback, ", ")))
		}
		out.WriteString(".\n")
	}
	return out.String()
}

// BeginMarker returns the exact opening delimiter used by the renderer.
func BeginMarker() string { return beginMarker }

// EndMarker returns the exact closing delimiter used by the renderer.
func EndMarker() string { return endMarker }

// InterpretationInstruction returns the text placed outside the delimiters.
func InterpretationInstruction() string { return interpretationInstruction }
