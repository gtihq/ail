// Package cli wires the parser, interpreter, linter, validator and renderer
// into a single command line interface used by the AIL executable.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gtihq/ail/packages/core"
	"github.com/gtihq/ail/packages/interpreter"
	"github.com/gtihq/ail/packages/linter"
	"github.com/gtihq/ail/packages/parser"
	"github.com/gtihq/ail/packages/renderer"
	"github.com/gtihq/ail/packages/validator"
)

const usage = `AIL 0.3 command line toolchain

Usage:
  ail <command> <file> [options]

Commands:
  parse      Emit the AST and syntax diagnostics as JSON
  interpret  Emit deterministic IR as JSON
  lint       Emit advisory analysis as JSON
  validate   Emit machine-readable validation as JSON
  render     Emit host instructions with mandatory delimiters
  check      Run every stage and emit a combined JSON summary

Options:
  --host=cursor|opencode|agents-md|markdown   Renderer host (default markdown)
  --mode=INLINE|FILE|COMPACT|VERBOSE          Injection mode (default INLINE)
  --no-delimiters                             Omit mandatory delimiters
  --provenance                                Include source provenance in verbose output
`

type options struct {
	path            string
	host            core.HostFormat
	mode            core.InjectionMode
	includeDelims   bool
	includeProv     bool
}

func parseArgs(args []string) options {
	opts := options{
		host:          core.HostMarkdown,
		mode:          core.ModeInline,
		includeDelims: true,
	}
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--host="):
			opts.host = core.HostFormat(strings.TrimPrefix(arg, "--host="))
		case strings.HasPrefix(arg, "--mode="):
			opts.mode = core.InjectionMode(strings.TrimPrefix(arg, "--mode="))
		case arg == "--no-delimiters":
			opts.includeDelims = false
		case arg == "--provenance":
			opts.includeProv = true
		case strings.HasPrefix(arg, "--"):
			continue
		case opts.path == "":
			opts.path = arg
		}
	}
	return opts
}

func emitJSON(stdout io.Writer, value any) int {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintf(stdout, "{\"error\":%q}\n", err.Error())
		return 1
	}
	fmt.Fprintln(stdout, string(encoded))
	return 0
}

// Run executes a CLI command and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	command := strings.ToLower(args[0])
	if command == "help" || command == "--help" || command == "-h" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	opts := parseArgs(args[1:])
	if opts.path == "" {
		fmt.Fprintln(stderr, "error: a file path is required")
		fmt.Fprint(stderr, usage)
		return 2
	}
	source, err := os.ReadFile(opts.path)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot read %q: %v\n", opts.path, err)
		return 1
	}

	parsed := parser.Parse(string(source), opts.path)
	interpreted := interpreter.Interpret(parsed)

	switch command {
	case "parse":
		return emitJSON(stdout, parsed)
	case "interpret":
		return emitJSON(stdout, interpreted)
	case "lint":
		return emitJSON(stdout, linter.Lint(parsed, interpreted.IR, core.LintConfig{}))
	case "validate":
		return emitJSON(stdout, validator.Validate(parsed, interpreted.IR))
	case "render":
		includeDelims := opts.includeDelims
		result := renderer.Render(interpreted.IR, core.RenderOptions{
			Host:              opts.host,
			Mode:              opts.mode,
			IncludeDelimiters: &includeDelims,
			IncludeProvenance: opts.includeProv,
		})
		fmt.Fprint(stdout, result.Text)
		return 0
	case "check":
		lintResult := linter.Lint(parsed, interpreted.IR, core.LintConfig{})
		validation := validator.Validate(parsed, interpreted.IR)
		return emitJSON(stdout, struct {
			File          string                 `json:"file"`
			ParseErrors   int                    `json:"parseErrors"`
			Interpreter   int                    `json:"interpreterDiagnostics"`
			Rules         int                    `json:"rules"`
			Lint          core.AnalyzerResult    `json:"lint"`
			Validation    core.ValidationResult  `json:"validation"`
			OverallValid  bool                   `json:"overallValid"`
		}{
			File:         opts.path,
			ParseErrors:  countSeverity(parsed.Diagnostics, core.SeverityError),
			Interpreter:  len(interpreted.Diagnostics),
			Rules:        len(interpreted.IR.Rules),
			Lint:         lintResult,
			Validation:   validation,
			OverallValid: validation.Valid && countSeverity(parsed.Diagnostics, core.SeverityError) == 0,
		})
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n", command)
		fmt.Fprint(stderr, usage)
		return 2
	}
}

func countSeverity(diagnostics []core.Diagnostic, severity core.Severity) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severity {
			count++
		}
	}
	return count
}
