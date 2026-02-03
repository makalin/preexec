package rewrite

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SafeRewrite returns a safer version of the command:
// - strips ANSI escape sequences
// - removes zero-width and BiDi control characters
// - optionally normalizes common homoglyphs to ASCII (Latin)
func SafeRewrite(cmd string, normalizeHomoglyphs bool) string {
	var b strings.Builder
	b.Grow(len(cmd))
	i := 0
	for i < len(cmd) {
		r, w := utf8.DecodeRuneInString(cmd[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		// Strip ANSI (ESC ...)
		if r == 0x1B {
			end := i + 1
			for end < len(cmd) {
				c := cmd[end]
				if c >= 0x40 && c <= 0x7E || c == 0x07 || c == '\n' {
					end++
					break
				}
				end++
				if end-i > 64 {
					break
				}
			}
			i = end
			continue
		}
		// Drop zero-width and BiDi
		if isZeroWidthOrBidi(r) {
			i += w
			continue
		}
		// Optional: replace common Cyrillic homoglyphs with Latin
		if normalizeHomoglyphs {
			if lat := homoglyphToLatin(r); lat != 0 {
				r = lat
			}
		}
		b.WriteRune(r)
		i += w
	}
	return b.String()
}

func isZeroWidthOrBidi(r rune) bool {
	switch r {
	case '\u200B', '\u200C', '\u200D', '\uFEFF', '\u2060', '\u180E':
		return true
	case '\u202E', '\u202D', '\u2066', '\u2069':
		return true
	}
	return false
}

// homoglyphToLatin returns Latin equivalent for common Cyrillic lookalikes, or 0.
func homoglyphToLatin(r rune) rune {
	cyrToLat := map[rune]rune{
		'\u0430': 'a', '\u0435': 'e', '\u043E': 'o', '\u043F': 'p',
		'\u0441': 'c', '\u0443': 'y', '\u0445': 'x', '\u0501': 'd',
		'\u0432': 'b', '\u0437': 'z', '\u0438': 'u', '\u0439': 'i',
		'\u043A': 'k', '\u043C': 'm', '\u043D': 'n', '\u0433': 'g',
		'\u0442': 't', '\u0440': 'r', '\u0444': 'f',
	}
	if lat, ok := cyrToLat[r]; ok {
		return lat
	}
	return 0
}

// StripANSI removes all ANSI escape sequences from s.
func StripANSI(s string) string {
	return SafeRewrite(s, false)
}

// VisibleRunes returns the number of visible (non-ANSI, non-zero-width) runes.
func VisibleRunes(s string) int {
	n := 0
	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		if r == 0x1B {
			end := i + 1
			for end < len(s) {
				c := s[end]
				if c >= 0x40 && c <= 0x7E || c == 0x07 || c == '\n' {
					end++
					break
				}
				end++
				if end-i > 64 {
					break
				}
			}
			i = end
			continue
		}
		if !isZeroWidthOrBidi(r) && !unicode.Is(unicode.Cc, r) {
			n++
		}
		i += w
	}
	return n
}
