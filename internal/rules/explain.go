package rules

// Descriptions maps rule ID to short description for "preexec explain".
var Descriptions = map[string]string{
	"unicode-homoglyph":    "Detects mixed scripts and homoglyphs (e.g. Cyrillic 'а' vs Latin 'a') used to spoof domains or commands.",
	"zero-width":           "Detects zero-width and invisible Unicode (ZWSP, ZWJ, BOM) that can hide payloads.",
	"bidi-controls":        "Detects BiDi override characters (RLO, LRO, FSI, PDI) that can reorder text and hide content.",
	"ansi-escape":          "Detects ANSI escape sequences that can deceive the terminal or inject output.",
	"pipe-to-shell":        "Flags curl|bash, wget|sh and similar patterns; suggests download-then-inspect.",
	"dotfile-write":        "Warns when command writes to .bashrc, .zshrc, .profile, .ssh, or similar.",
	"persistence-patterns": "Flags cron, systemd, launchd, and eval $(...) as potential persistence or code execution.",
	"shortener-domains":    "Warns on URL shortener or redirect domains (e.g. bit.ly, tinyurl.com) in commands.",
	"subshell-command":     "Flags $(), backticks, and <( ) process substitution as hidden command execution.",
}

// Describe returns the description for a rule ID, or empty string.
func Describe(ruleID string) string {
	return Descriptions[ruleID]
}
