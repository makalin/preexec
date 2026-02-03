# PreExec
**A local-first shell firewall that inspects commands *before* they execute.**  
Stops Unicode homoglyph attacks, hidden characters, ANSI injection, and dangerous `curl | bash` patterns.

> Trust nothing after paste.

---

## Why PreExec exists

Modern dev workflows are built on:
- copying commands from GitHub READMEs
- pasting from ChatGPT
- one-liners like `curl | bash`

This is convenient — and extremely dangerous.

Attackers exploit:
- **Unicode homoglyphs** (Latin `a` vs Cyrillic `а`)
- **Zero-width characters** that hide real payloads
- **ANSI escape sequences** that lie in your terminal
- **Hidden pipes & subshells**
- **Dotfile overwrites for persistence**
- **Blind remote execution**

PreExec makes your shell **paranoid by default**.

---

## What it does

PreExec hooks into your shell and inspects every command **before execution**.

It can:
- block execution
- warn with explanation
- require confirmation
- or silently allow

All locally. No telemetry. No network calls.

---

## Features

### Text & Unicode Attacks
- Detects mixed-script tokens (Latin/Cyrillic/Greek)
- Flags homoglyph characters
- Catches zero-width & invisible Unicode
- BiDi control characters (RLO/LRO/FSI/PDI)

### Terminal Deception
- ANSI escape sequence injection
- OSC clipboard abuse
- Cursor movement & output rewriting
- Prompt spoofing

### Shell Risk Patterns
- `curl|bash`, `wget|sh`, `eval $(...)`
- Subshells: `$()`, backticks
- Hidden heredocs / here-strings
- Silent redirects (`>/dev/null 2>&1`)
- Process substitution `<( )`

### Persistence & Hijacking
- Writes to:
  - `~/.bashrc`
  - `~/.zshrc`
  - `~/.profile`
  - `~/.ssh/*`
- Alias / function hijacking
- PATH poisoning
- Cron / systemd / launch agents

### Exfil Heuristics
- SSH key reads
- Token patterns
- Suspicious POST requests
- Shortener domains

---

## Installation

### Homebrew (planned)
```bash
brew install preexec
```

### Go (recommended)

```bash
go install github.com/makalin/preexec/cmd/preexec@latest
```

From source:

```bash
git clone https://github.com/makalin/preexec.git && cd preexec && go build -o preexec ./cmd/preexec
```

### Rust (planned)

```bash
cargo install preexec
```

---

## Quick Start

### Scan a command

```bash
preexec check -- "curl -sSL https://instаll.example | bash"
```

### Scan clipboard (macOS)

```bash
preexec check --clipboard
```

### Machine-readable output (JSON)

```bash
preexec check --json -- "curl example.com | sh"
```

### Scan a script or README

```bash
preexec scan ./install.sh
preexec scan ./README.md --extract
```

### Scan shell history

```bash
preexec history --shell zsh --last 1000
```

### Safe-rewrite (strip ANSI, normalize homoglyphs)

```bash
echo "curl https://instаll.example" | preexec rewrite --normalize
```

### Compare two commands

```bash
preexec diff "curl example.com" "curl instаll.example"
```

### Pre-commit hook (scan staged or given files)

```bash
preexec pre-commit
# or with paths:
preexec pre-commit ./scripts/install.sh
```

---

## Shell Hook

### Zsh

```bash
# ~/.zshrc
eval "$(preexec hook zsh)"
```

### Bash

```bash
# ~/.bashrc
eval "$(preexec hook bash)"
```

### Fish

```bash
# ~/.config/fish/config.fish
preexec hook fish | source
```

### PowerShell

```bash
# Add to $PROFILE
preexec hook powershell
```

Behavior:

* **BLOCK** → command not executed
* **WARN** → explain + require confirmation
* **PASS** → normal execution

---

## Example Output

### Unicode Homoglyph

```text
BLOCK unicode-homoglyph
 token: instаll.example
 issue: Cyrillic 'а' (U+0430) used instead of Latin 'a' (U+0061)
 position: 18
 suggestion: use ASCII-only domain
```

### ANSI Injection

```text
WARN ansi-escape
 issue: ESC (U+001B) control sequence detected
 suggestion: paste as plain text
```

### curl | bash

```text
WARN pipe-to-shell
 pattern: curl ... | bash
 suggestion: download, inspect, then execute
```

### Describe a rule

```bash
preexec explain pipe-to-shell
# pipe-to-shell: Flags curl|bash, wget|sh and similar patterns; suggests download-then-inspect.
```

---

## Commands

| Command | Description |
| ------- | ----------- |
| `preexec check [--json] [--clipboard] -- <command>` | Inspect a single command (or clipboard) |
| `preexec scan [--extract] <path>` | Scan files or directories; `--extract` = code blocks only from Markdown |
| `preexec history [--shell zsh\|bash] [--last N]` | Scan shell history |
| `preexec show --codepoints [--] [string]` | Reveal hidden Unicode (codepoints) |
| `preexec show --urls [--] [string]` | Extract URLs and show IDN (non-ASCII host) hints |
| `preexec hook zsh\|bash\|fish\|powershell` | Print shell hook script |
| `preexec rules list` | List all rules |
| `preexec rules test <rule> [command]` | Test one rule on a command |
| `preexec explain [rule...]` | Describe rule(s); no args = list all descriptions |
| `preexec diff <cmd1> <cmd2>` | Compare two commands (or two lines from stdin) |
| `preexec rewrite [--normalize] [--] <command>` | Safe-rewrite: strip ANSI, zero-width; `--normalize` = homoglyphs → ASCII |
| `preexec pre-commit [path...]` | Scan staged files (git) or given paths; for use in pre-commit hooks |
| `preexec config init` | Create default config at `~/.config/preexec/config.toml` |

---

## Exit Codes

| Code | Meaning |
| ---- | ------- |
| 0    | PASS    |
| 10   | WARN    |
| 20   | BLOCK   |
| 2    | ERROR   |

---

## Configuration

`~/.config/preexec/config.toml`

```toml
mode = "warn"
confirm_on_warn = true
ascii_only_domains = true

[allow]
domains = ["github.com", "raw.githubusercontent.com"]

[deny]
domains = ["bit.ly", "tinyurl.com"]

[rules]
unicode_homoglyph = true
zero_width = true
bidi_controls = true
ansi_escape = true
pipe_to_shell = true
dotfile_write = true
persistence_patterns = true
shortener_domains = true
subshell_command = true
```

Rule names for `preexec explain` and `rules test`: `unicode-homoglyph`, `zero-width`, `bidi-controls`, `ansi-escape`, `pipe-to-shell`, `dotfile-write`, `persistence-patterns`, `shortener-domains`, `subshell-command`.

---

## Security Philosophy

PreExec assumes:

* input is hostile
* terminals lie
* users don’t read
* copy/paste is the new attack vector

It prefers **false positives over silent compromise**.

---

## Use Cases

* Personal dev machines
* CI pipelines (lint install scripts)
* Security teams auditing onboarding docs
* Open source maintainers checking README snippets
* AI-generated shell commands

---

## Roadmap

* [x] IDN / URL hints (`show --urls`)
* [x] Fish & PowerShell hooks
* [x] Pre-commit integration (`pre-commit`)
* [x] Safe-rewrite mode (`rewrite [--normalize]`)
* [ ] Clipboard watcher daemon
* [ ] TUI interactive inspector
* [ ] Full Punycode decode in output

---

## Contributing

PRs welcome:

* new rules
* real-world malicious samples
* better Unicode confusable mapping
* platform hooks

---

## License

MIT

---

## Author
Mehmet Turgay Akalın — Full Stack Developer  
GitHub: https://github.com/makalin

---

## Motto

> **You are always one paste away from compromise.
> PreExec makes sure that paste isn’t fatal.**
