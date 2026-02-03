package engine

import "github.com/makalin/preexec/internal/types"

// Result is the outcome of checking a command.
type Result struct {
	Severity types.Severity
	Findings []types.Finding
}

// WorstSeverity returns the highest severity in findings.
func (r *Result) WorstSeverity() types.Severity {
	s := types.Pass
	for _, f := range r.Findings {
		if f.Severity > s {
			s = f.Severity
		}
	}
	return s
}

// ExitCode returns the CLI exit code for this result.
func (r *Result) ExitCode() int {
	switch r.WorstSeverity() {
	case types.Block:
		return 20
	case types.Warn:
		return 10
	default:
		return 0
	}
}
