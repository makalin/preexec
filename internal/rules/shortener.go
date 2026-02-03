package rules

import (
	"regexp"
	"strings"

	"github.com/makalin/preexec/internal/types"
)

// ShortenerRule detects URL shortener and redirect domains in commands.
type ShortenerRule struct{}

func (r *ShortenerRule) ID() string { return "shortener-domains" }

// Common shortener/redirect domains (deny list heuristics).
var shortenerDomains = []string{
	"bit.ly", "tinyurl.com", "t.co", "goo.gl", "ow.ly",
	"is.gd", "buff.ly", "adf.ly", "j.mp", "bc.vc",
	"bit.do", "lnkd.in", "db.tt", "short.link", "cutt.ly",
}

var urlLikeRe = regexp.MustCompile(`https?://[^\s'"<>|&;]+`)

func (r *ShortenerRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("shortener_domains") {
		return nil
	}
	var out []types.Finding
	for _, loc := range urlLikeRe.FindAllStringIndex(cmd, -1) {
		url := cmd[loc[0]:loc[1]]
		lower := strings.ToLower(url)
		for _, domain := range shortenerDomains {
			if strings.Contains(lower, domain) {
				out = append(out, types.Finding{
					RuleID:     r.ID(),
					Severity:   types.Warn,
					Token:      url,
					Issue:      "URL shortener or redirect domain: " + domain,
					Position:   loc[0],
					Suggestion: "use full URL or trusted source",
				})
				break
			}
		}
	}
	return out
}
