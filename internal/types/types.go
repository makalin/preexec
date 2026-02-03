package types

// Severity is pass, warn, or block.
type Severity int

const (
	Pass  Severity = 0
	Warn  Severity = 10
	Block Severity = 20
)

// Finding is a single rule hit.
type Finding struct {
	RuleID     string
	Severity   Severity
	Token      string
	Issue      string
	Position   int
	Suggestion string
}
