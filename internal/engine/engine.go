package engine

import (
	"github.com/makalin/preexec/internal/rules"
	"github.com/makalin/preexec/internal/types"
)

// Run runs all rules on cmd and returns aggregated result.
func Run(cmd string, cfg rules.RuleConfig) *Result {
	res := &Result{Severity: types.Pass}
	for _, rule := range rules.All() {
		f := rule.Check(cmd, cfg)
		res.Findings = append(res.Findings, f...)
	}
	res.Severity = res.WorstSeverity()
	return res
}
