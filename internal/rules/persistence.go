package rules

import (
	"regexp"
	"strings"

	"github.com/makalin/preexec/internal/types"
)

// PersistenceRule detects cron, systemd, launchd, eval $(...), etc.
type PersistenceRule struct{}

func (r *PersistenceRule) ID() string { return "persistence-patterns" }

var (
	cronRe    = regexp.MustCompile(`(?i)(crontab\s|/etc/cron|/var/spool/cron)`)
	systemdRe = regexp.MustCompile(`(?i)(systemctl\s+(enable|start)|/etc/systemd)`)
	launchdRe = regexp.MustCompile(`(?i)(launchd|launchctl\s+load|~/Library/LaunchAgents)`)
	evalRe    = regexp.MustCompile(`eval\s*\$?\s*\(`)
)

func (r *PersistenceRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("persistence_patterns") {
		return nil
	}
	var out []types.Finding
	for _, re := range []*regexp.Regexp{cronRe, systemdRe, launchdRe, evalRe} {
		for _, loc := range re.FindAllStringIndex(cmd, -1) {
			token := cmd[loc[0]:loc[1]]
			issue := "suspicious persistence or dynamic execution: " + token
			if strings.Contains(token, "eval") {
				issue = "eval $(...) can run arbitrary code"
			}
			out = append(out, types.Finding{
				RuleID:     r.ID(),
				Severity:   types.Warn,
				Token:      token,
				Issue:      issue,
				Position:   loc[0],
				Suggestion: "review before enabling cron/systemd/launchd or using eval",
			})
		}
	}
	return out
}
