package rules

import (
	"regexp"

	"github.com/makalin/preexec/internal/types"
)

// SubshellRule detects hidden subshells and command substitution.
type SubshellRule struct{}

func (r *SubshellRule) ID() string { return "subshell-command" }

var (
	subshellDollarRe   = regexp.MustCompile(`\$\([^)]*\)`)
	subshellBacktickRe = regexp.MustCompile("`[^`]*`")
	processSubstRe     = regexp.MustCompile(`<\([^)]*\)`)
)

func (r *SubshellRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("subshell_command") {
		return nil
	}
	var out []types.Finding
	for _, re := range []*regexp.Regexp{subshellDollarRe, subshellBacktickRe, processSubstRe} {
		for _, loc := range re.FindAllStringIndex(cmd, -1) {
			token := cmd[loc[0]:loc[1]]
			issue := "command substitution or subshell: " + token
			if re == processSubstRe {
				issue = "process substitution: " + token
			}
			out = append(out, types.Finding{
				RuleID:     r.ID(),
				Severity:   types.Warn,
				Token:      token,
				Issue:      issue,
				Position:   loc[0],
				Suggestion: "ensure substituted command is trusted",
			})
		}
	}
	return out
}
