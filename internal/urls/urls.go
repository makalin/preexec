package urls

import (
	"regexp"
	"strings"
)

// ExtractURLs returns URL-like substrings from s (http/https).
var urlRe = regexp.MustCompile(`https?://[^\s'"<>|&;)\]]+`)

func ExtractURLs(s string) []string {
	return urlRe.FindAllString(s, -1)
}

// HasIDN returns true if the host part of the URL contains non-ASCII (potential IDN).
func HasIDN(url string) bool {
	host := hostFromURL(url)
	if host == "" {
		return false
	}
	for _, r := range host {
		if r > 0x7F {
			return true
		}
	}
	return false
}

func hostFromURL(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	if i := strings.Index(s, ":"); i >= 0 {
		s = s[:i]
	}
	return s
}

// HostForDisplay returns the host part for display (e.g. for punycode hint).
func HostForDisplay(url string) string {
	return hostFromURL(url)
}
