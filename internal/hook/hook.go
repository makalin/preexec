package hook

import (
	"fmt"
	"strings"
)

// Zsh returns the zsh preexec hook script.
func Zsh(preexecPath string) string {
	if preexecPath == "" {
		preexecPath = "preexec"
	}
	return fmt.Sprintf(`# PreExec hook for zsh
__preexec_cmd() {
  local cmd="$1"
  local code
  %s check -- "$cmd"
  code=$?
  if [ "$code" -eq 20 ]; then
    echo "PreExec BLOCK: command not executed."
    return 1
  fi
  if [ "$code" -eq 10 ]; then
    echo "PreExec WARN: review above. Run again to execute."
    return 1
  fi
  return 0
}
preexec_functions+=(__preexec_cmd)
`, preexecPath)
}

// Bash returns the bash preexec hook script.
func Bash(preexecPath string) string {
	if preexecPath == "" {
		preexecPath = "preexec"
	}
	// Bash doesn't have preexec natively; use DEBUG trap
	return fmt.Sprintf(`# PreExec hook for bash (DEBUG trap)
__preexec_trap() {
  if [ -n "$BASH_COMMAND" ] && [ "$BASH_COMMAND" != "printf" ]; then
    local cmd="$BASH_COMMAND"
    local code
    %s check -- "$cmd"
    code=$?
    if [ "$code" -eq 20 ]; then
      echo "PreExec BLOCK: command not executed."
      return 1
    fi
    if [ "$code" -eq 10 ]; then
      echo "PreExec WARN: review above. Run again to execute."
      return 1
    fi
  fi
}
trap '__preexec_trap' DEBUG
`, preexecPath)
}

// Fish returns the fish shell preexec hook script.
func Fish(preexecPath string) string {
	if preexecPath == "" {
		preexecPath = "preexec"
	}
	return fmt.Sprintf(`# PreExec hook for fish
function __preexec_cmd --on-event fish_preexec
  set -l cmd (string join " " $argv)
  %s check -- $cmd
  set -l code $status
  if [ $code -eq 20 ]
    echo "PreExec BLOCK: command not executed."
    return 1
  end
  if [ $code -eq 10 ]
    echo "PreExec WARN: review above. Run again to execute."
    return 1
  end
  return 0
end
`, preexecPath)
}

// PowerShell returns the PowerShell preexec hook script (Prompt override).
func PowerShell(preexecPath string) string {
	if preexecPath == "" {
		preexecPath = "preexec"
	}
	return fmt.Sprintf(`# PreExec hook for PowerShell - add to $PROFILE
function preexec_check {
  param([string]$cmd)
  $result = & %s check -- $cmd 2>&1
  $exitCode = $LASTEXITCODE
  if ($exitCode -eq 20) {
    Write-Host "PreExec BLOCK: command not executed."
    return $false
  }
  if ($exitCode -eq 10) {
    Write-Host $result
    Write-Host "PreExec WARN: review above. Run again to execute."
    return $false
  }
  return $true
}
# Wrap default prompt to run preexec on submitted command (simplified: run on each prompt)
$originalPrompt = Get-Command prompt -ErrorAction SilentlyContinue
`, preexecPath)
}

// Script returns hook script for the given shell.
func Script(shell string, preexecPath string) (string, error) {
	switch strings.ToLower(shell) {
	case "zsh":
		return Zsh(preexecPath), nil
	case "bash":
		return Bash(preexecPath), nil
	case "fish":
		return Fish(preexecPath), nil
	case "powershell", "pwsh":
		return PowerShell(preexecPath), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (use zsh, bash, fish, powershell)", shell)
	}
}
