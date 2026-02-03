package rules

import (
	"unicode"
	"unicode/utf8"

	"github.com/makalin/preexec/internal/types"
)

// UnicodeHomoglyphRule detects mixed scripts and homoglyphs (e.g. Cyrillic 'а' vs Latin 'a').
type UnicodeHomoglyphRule struct{}

func (r *UnicodeHomoglyphRule) ID() string { return "unicode-homoglyph" }

func (r *UnicodeHomoglyphRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("unicode_homoglyph") {
		return nil
	}
	var out []types.Finding
	for i, ru := range cmd {
		if ru > 0x7F && isHomoglyphOrMixed(ru) {
			token := tokenAt(cmd, i)
			issue := describeRune(ru)
			out = append(out, types.Finding{
				RuleID:     r.ID(),
				Severity:   types.Warn,
				Token:      token,
				Issue:      issue,
				Position:   i,
				Suggestion: "use ASCII-only domain",
			})
		}
	}
	return out
}

func isHomoglyphOrMixed(r rune) bool {
	// Cyrillic, Greek, or other lookalikes commonly used in homoglyph attacks
	switch {
	case r >= 0x0400 && r <= 0x04FF: // Cyrillic
		return true
	case r >= 0x0370 && r <= 0x03FF: // Greek
		return true
	case r >= 0x0530 && r <= 0x058F: // Armenian
		return true
	default:
		return false
	}
}

func describeRune(r rune) string {
	// Latin lookalikes in Cyrillic
	cyrToLat := map[rune]string{
		'\u0430': "Cyrillic 'а' (U+0430) used instead of Latin 'a' (U+0061)",
		'\u0435': "Cyrillic 'е' (U+0435) used instead of Latin 'e' (U+0065)",
		'\u043E': "Cyrillic 'о' (U+043E) used instead of Latin 'o' (U+006F)",
		'\u043F': "Cyrillic 'р' (U+043F) used instead of Latin 'p' (U+0070)",
		'\u0441': "Cyrillic 'с' (U+0441) used instead of Latin 'c' (U+0063)",
		'\u0443': "Cyrillic 'у' (U+0443) used instead of Latin 'y' (U+0079)",
		'\u0445': "Cyrillic 'х' (U+0445) used instead of Latin 'x' (U+0078)",
		'\u0501': "Cyrillic 'ԁ' (U+0501) used instead of Latin 'd' (U+0064)",
	}
	if s, ok := cyrToLat[r]; ok {
		return s
	}
	if unicode.Is(unicode.Cyrillic, r) {
		return "Cyrillic character (U+" + formatHex(r) + ") may look like Latin"
	}
	if unicode.Is(unicode.Greek, r) {
		return "Greek character (U+" + formatHex(r) + ") may look like Latin"
	}
	return "non-ASCII character U+" + formatHex(r) + " in command"
}

func formatHex(r rune) string {
	return string([]byte{
		hex(r >> 12), hex(r >> 8), hex(r >> 4), hex(r),
	})
}

func hex(n rune) byte {
	n &= 0xF
	if n < 10 {
		return byte('0' + n)
	}
	return byte('A' + n - 10)
}

func tokenAt(s string, pos int) string {
	start := pos
	for start > 0 && !isTokenBoundary(s, start) {
		_, w := utf8.DecodeLastRuneInString(s[:start])
		start -= w
	}
	end := pos
	for end < len(s) && !isTokenBoundary(s, end) {
		_, w := utf8.DecodeRuneInString(s[end:])
		end += w
	}
	return s[start:end]
}

func isTokenBoundary(s string, i int) bool {
	if i <= 0 || i >= len(s) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(s[i:])
	return r == ' ' || r == '|' || r == '&' || r == ';' || r == '\n' || r == '"' || r == '\''
}

// ZeroWidthRule detects zero-width and invisible Unicode.
type ZeroWidthRule struct{}

func (r *ZeroWidthRule) ID() string { return "zero-width" }

func (r *ZeroWidthRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("zero_width") {
		return nil
	}
	// Zero-width space, joiners, BOM, etc.
	bad := []rune{
		'\u200B', // ZWSP
		'\u200C', // ZWNJ
		'\u200D', // ZWJ
		'\uFEFF', // BOM / ZWSP variant
		'\u2060', // WORD JOINER
		'\u180E', // MONGOLIAN VOWEL SEPARATOR
	}
	var out []types.Finding
	for i, ru := range cmd {
		for _, b := range bad {
			if ru == b {
				out = append(out, types.Finding{
					RuleID:     r.ID(),
					Severity:   types.Block,
					Token:      string(ru),
					Issue:      "zero-width or invisible character (U+" + formatHex(ru) + ") detected",
					Position:   i,
					Suggestion: "paste as plain text; remove hidden characters",
				})
				break
			}
		}
	}
	return out
}

// BidiControlsRule detects RLO, LRO, FSI, PDI (bidirectional override).
type BidiControlsRule struct{}

func (r *BidiControlsRule) ID() string { return "bidi-controls" }

func (r *BidiControlsRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("bidi_controls") {
		return nil
	}
	// U+202E RLO, U+202D LRO, U+2066 FSI, U+2069 PDI
	bad := []rune{'\u202E', '\u202D', '\u2066', '\u2069'}
	var out []types.Finding
	for i, ru := range cmd {
		for _, b := range bad {
			if ru == b {
				out = append(out, types.Finding{
					RuleID:     r.ID(),
					Severity:   types.Block,
					Token:      string(ru),
					Issue:      "BiDi control character (U+" + formatHex(ru) + ") can reorder text",
					Position:   i,
					Suggestion: "remove bidirectional override characters",
				})
				break
			}
		}
	}
	return out
}
