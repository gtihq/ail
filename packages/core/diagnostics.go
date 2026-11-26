package core

import (
	"fmt"
	"strconv"
)

// MakeDiagnostic builds a diagnostic with all required provenance fields.
func MakeDiagnostic(code string, severity Severity, kind DiagnosticKind, message, file string, loc SourceLocation) Diagnostic {
	return Diagnostic{
		Code:     code,
		Severity: severity,
		Kind:     kind,
		Message:  message,
		File:     file,
		Loc:      loc,
	}
}

// Summarize counts diagnostics by severity.
func Summarize(diagnostics []Diagnostic) Summary {
	var summary Summary
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case SeverityError:
			summary.Errors++
		case SeverityWarning:
			summary.Warnings++
		case SeverityInfo:
			summary.Infos++
		case SeverityHint:
			summary.Hints++
		}
	}
	return summary
}

// FormatDiagnostic renders one diagnostic as a human-readable line.
func FormatDiagnostic(diagnostic Diagnostic) string {
	return fmt.Sprintf("%s:%d:%d %s %s: %s",
		diagnostic.File,
		diagnostic.Loc.Line,
		diagnostic.Loc.Column,
		diagnostic.Severity,
		diagnostic.Code,
		diagnostic.Message,
	)
}

// MergeDiagnostics combines diagnostic groups while preserving first-seen order.
func MergeDiagnostics(groups ...[]Diagnostic) []Diagnostic {
	merged := make([]Diagnostic, 0)
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, diagnostic := range group {
			key := fmt.Sprintf("%s|%d|%d|%s|%s",
				diagnostic.File,
				diagnostic.Loc.Line,
				diagnostic.Loc.Column,
				diagnostic.Code,
				diagnostic.Message,
			)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, diagnostic)
		}
	}
	return merged
}

// formatNumber renders a float without a trailing decimal zero.
func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
