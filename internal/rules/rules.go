package rules

import "github.com/makalin/preexec/internal/types"

// Rule inspects a command string and reports findings.
type Rule interface {
	ID() string
	Check(cmd string, cfg RuleConfig) []types.Finding
}

// RuleConfig provides rule-specific config (e.g. enabled, allow/deny lists).
type RuleConfig interface {
	RuleEnabled(name string) bool
}

// All returns all built-in rules.
func All() []Rule {
	return []Rule{
		&UnicodeHomoglyphRule{},
		&ZeroWidthRule{},
		&BidiControlsRule{},
		&ANSIEscapeRule{},
		&PipeToShellRule{},
		&DotfileWriteRule{},
		&PersistenceRule{},
		&ShortenerRule{},
		&SubshellRule{},
	}
}
