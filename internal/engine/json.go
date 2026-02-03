package engine

import (
	"encoding/json"

	"github.com/makalin/preexec/internal/types"
)

// JSONResult is the machine-readable result for --json output.
type JSONResult struct {
	Severity string        `json:"severity"`
	ExitCode int           `json:"exit_code"`
	Findings []JSONFinding `json:"findings,omitempty"`
}

// JSONFinding is a single finding for JSON output.
type JSONFinding struct {
	RuleID     string `json:"rule_id"`
	Severity   string `json:"severity"`
	Token      string `json:"token,omitempty"`
	Issue      string `json:"issue"`
	Position   int    `json:"position"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ToJSONResult converts Result to JSONResult.
func (r *Result) ToJSONResult() JSONResult {
	j := JSONResult{
		Severity: severityToStr(r.WorstSeverity()),
		ExitCode: r.ExitCode(),
	}
	for _, f := range r.Findings {
		j.Findings = append(j.Findings, JSONFinding{
			RuleID:     f.RuleID,
			Severity:   severityToStr(f.Severity),
			Token:      f.Token,
			Issue:      f.Issue,
			Position:   f.Position,
			Suggestion: f.Suggestion,
		})
	}
	return j
}

func severityToStr(s types.Severity) string {
	switch s {
	case types.Block:
		return "block"
	case types.Warn:
		return "warn"
	default:
		return "pass"
	}
}

// MarshalJSON returns JSON bytes for the result.
func (r *Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.ToJSONResult())
}
