package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds preexec configuration.
type Config struct {
	Mode             string          `toml:"mode"` // pass, warn, block
	ConfirmOnWarn    bool            `toml:"confirm_on_warn"`
	ASCIIOnlyDomains bool            `toml:"ascii_only_domains"`
	Allow            AllowDeny       `toml:"allow"`
	Deny             AllowDeny       `toml:"deny"`
	Rules            map[string]bool `toml:"rules"`
}

// AllowDeny holds allow/deny domain lists.
type AllowDeny struct {
	Domains []string `toml:"domains"`
}

// DefaultConfig returns a config with safe defaults.
func DefaultConfig() *Config {
	return &Config{
		Mode:             "warn",
		ConfirmOnWarn:    true,
		ASCIIOnlyDomains: true,
		Rules: map[string]bool{
			"unicode_homoglyph":    true,
			"zero_width":           true,
			"bidi_controls":        true,
			"ansi_escape":          true,
			"pipe_to_shell":        true,
			"dotfile_write":        true,
			"persistence_patterns": true,
			"shortener_domains":    true,
			"subshell_command":     true,
		},
	}
}

// ConfigPath returns the default config path (~/.config/preexec/config.toml).
func ConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "preexec", "config.toml")
}

// Load reads config from path. Returns default config if file missing or on error.
func Load(path string) (*Config, error) {
	if path == "" {
		path = ConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), nil
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return DefaultConfig(), err
	}
	// Merge defaults for missing rules
	def := DefaultConfig()
	for k, v := range def.Rules {
		if _, ok := c.Rules[k]; !ok {
			c.Rules[k] = v
		}
	}
	if c.Mode == "" {
		c.Mode = def.Mode
	}
	return &c, nil
}

// RuleEnabled returns whether a rule is enabled.
func (c *Config) RuleEnabled(name string) bool {
	if c.Rules == nil {
		return true
	}
	v, ok := c.Rules[name]
	return !ok || v
}
