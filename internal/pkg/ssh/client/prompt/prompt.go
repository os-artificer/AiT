/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

// Package prompt provides AiT's display prompt strings and the shell setup
// fragments needed to install them on a remote PTY. All functions here are
// pure (no I/O); the parent ssh/client package consumes them when wiring an
// interactive session.
package prompt

import (
	"fmt"
	"strings"
)

// Base is the AiT interactive shell prompt when not connected to a remote host.
const Base = "AiT> "

// Local is shown in AiT local mode (!ai-ssh local) within an SSH session.
const Local = "AiT Local> "

const remoteGitPS1 = `$(b=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) && [ -n "$b" ] && printf ":%s" "$b")`

const remoteDirShort = `$(d=$(basename "${PWD:-/}" 2>/dev/null); printf '%s' "${d:-/}")`

const remoteDirAbsoluteBash = `${PWD/#$HOME/\~}`

const remoteDirAbsoluteZsh = `${PWD/#$HOME/~}`

// RemotePromptReadySentinel is printed at the end of remote PS1 setup so the client
// can detect setup completion without polling heuristics alone.
const RemotePromptReadySentinel = ".__AIT_PROMPT_READY__"

const remotePromptReadyCmd = `; printf '%s\n' '` + RemotePromptReadySentinel + `'`

// remoteZshPromptPrintCmd redraws zsh PROMPT and resets the line editor. Bash must not
// manually print PS1 during setup—that leaves readline without an active prompt.
const remoteZshPromptPrintCmd = `[ -n "$ZSH_VERSION" ] && print -Pn "$PROMPT"` +
	`; [ -n "$ZSH_VERSION" ] && { zle 2>/dev/null && zle reset-prompt 2>/dev/null; }`

// remoteBashPromptRefreshCmd is a no-op that lets bash redraw PS1 after inject completes.
const remoteBashPromptRefreshCmd = `[ -z "$ZSH_VERSION" ] && true`

// RemotePromptRedrawCmd returns a shell one-liner that refreshes the remote prompt.
// Bash redraws via a completing no-op; zsh uses print -Pn and zle reset-prompt.
func RemotePromptRedrawCmd() string {
	return remoteBashPromptRefreshCmd + "; " + remoteZshPromptPrintCmd
}

// RemotePrefix returns the fixed AiT remote prompt prefix (e.g. "AiT user@host:").
func RemotePrefix(user, host string) string {
	return fmt.Sprintf("AiT %s@%s:", user, host)
}

// RemoteFormat builds a full display prompt for tests and documentation.
// branch empty omits the ":branch" segment.
func RemoteFormat(user, host, dir, branch string) string {
	if branch != "" {
		return fmt.Sprintf("AiT %s@%s:%s:%s> ", user, host, dir, branch)
	}
	return fmt.Sprintf("AiT %s@%s:%s> ", user, host, dir)
}

// Remote returns a minimal remote prompt placeholder for status messages (no cwd/git).
func Remote(user, host string) string {
	return RemotePrefix(user, host) + "> "
}

// RemoteSetupMarker returns the substring used to detect AiT remote prompt readiness.
func RemoteSetupMarker(user, host string) string {
	return RemotePrefix(user, host)
}

// IsRemotePromptLine reports whether line looks like an AiT remote shell prompt.
func IsRemotePromptLine(user, host, line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	prefix := RemotePrefix(user, host)
	if !strings.HasPrefix(line, prefix) {
		return false
	}
	return strings.HasSuffix(strings.TrimRight(line, " \t\r"), ">")
}

// BashPS1 wraps a display prompt for bash readline width counting.
func BashPS1(display string) string {
	return `\[\]` + display + `\[\]`
}

// LocalPS1 is bash PS1 for local mode.
func LocalPS1() string {
	return BashPS1(Local)
}

// remoteDisplayBody is the dynamic PS1/PROMPT body (cwd + optional git branch).
func remoteDisplayBody(user, host string, dirAbsolute bool) string {
	dir := remoteDirShort
	if dirAbsolute {
		dir = remoteDirAbsoluteBash
	}
	return escapeForSingleQuotedShell(RemotePrefix(user, host)) + dir + remoteGitPS1 + "> "
}

func remotePromptBody(user, host string, dirAbsolute bool) string {
	dir := remoteDirShort
	if dirAbsolute {
		dir = remoteDirAbsoluteZsh
	}
	return RemotePrefix(user, host) + dir + remoteGitPS1 + "> "
}

// RemotePS1Dynamic is bash PS1 with live cwd and git branch on the remote host.
func RemotePS1Dynamic(user, host string, dirAbsolute bool) string {
	return BashPS1(remoteDisplayBody(user, host, dirAbsolute))
}

// RemotePROMPTDynamic is zsh PROMPT with live cwd and git branch (requires PROMPT_SUBST).
func RemotePROMPTDynamic(user, host string, dirAbsolute bool) string {
	return escapeForSingleQuotedShell(remotePromptBody(user, host, dirAbsolute))
}

// RemotePS1 is bash PS1 for remote mode (dynamic).
func RemotePS1(user, host string, dirAbsolute bool) string {
	return RemotePS1Dynamic(user, host, dirAbsolute)
}

// RemoteSetupLine sets PS1/PROMPT and HISTCONTROL while history is off. Zsh clears the
// default prompt line and redraws via zle; bash only exports PS1 and lets readline
// redraw when the injected compound command completes. Echo suppression is provided
// by Client.Inject; local filter suppresses any leaked echo.
func remoteZshPrecmdHook(prompt string) string {
	return `[ -n "$ZSH_VERSION" ] && ait_precmd() { PROMPT='` + prompt + `'; RPROMPT=; }` +
		`; [ -n "$ZSH_VERSION" ] && precmd_functions=(${precmd_functions:#ait_precmd} ait_precmd)`
}

func remoteBashPromptCommandHook(ps1 string) string {
	return `[ -z "$ZSH_VERSION" ] && ait_prompt_command() { export PS1='` + ps1 + `'; }` +
		`; [ -z "$ZSH_VERSION" ] && export PROMPT_COMMAND=ait_prompt_command`
}

func RemoteSetupLine(user, host string, dirAbsolute bool) string {
	ps1 := RemotePS1Dynamic(user, host, dirAbsolute)
	prompt := RemotePROMPTDynamic(user, host, dirAbsolute)
	body := "export HISTCONTROL=ignorespace:erasedups" +
		"; [ -n \"$ZSH_VERSION\" ] && setopt HIST_IGNORE_SPACE PROMPT_SUBST" +
		"; export PS1='" + ps1 + "'" +
		"; " + remoteBashPromptCommandHook(ps1) +
		"; [ -n \"$ZSH_VERSION\" ] && export PROMPT='" + prompt + "'" +
		"; " + remoteZshPrecmdHook(prompt) +
		"; [ -n \"$ZSH_VERSION\" ] && printf '\\033[1A\\033[2K\\r' 2>/dev/null" +
		"; " + remoteZshPromptPrintCmd +
		remotePromptReadyCmd
	return wrapHistoryOff(body)
}

// RemoteSetupStripFragments returns scrollback/filter fragments for remote PS1 setup leakage.
func RemoteSetupStripFragments(user, host string, dirAbsolute bool) []string {
	return []string{
		"stty -echo 2>/dev/null",
		"stty echo 2>/dev/null",
		RemoteSetupLine(user, host, dirAbsolute),
		historyOffBashBegin,
		historyOffBashEnd,
		historyOffZshBegin,
		historyOffZshEnd,
		"_AIT_HIST",
		"export HISTCONTROL=ignorespace:erasedups",
		"setopt HIST_IGNORE_SPACE",
		"PROMPT_SUBST",
		"git rev-parse --abbrev-ref HEAD",
		`basename "${PWD`,
		"${PWD/#$HOME",
		".__AITPWD",
		"AITPWDEND",
		"AIT_CAPTURE_B_",
		"AIT_CAPTURE_E_",
		"export PS1='" + RemotePS1Dynamic(user, host, dirAbsolute) + "'",
		"export PROMPT='" + RemotePROMPTDynamic(user, host, dirAbsolute) + "'",
		remoteZshPromptPrintCmd,
		remoteBashPromptRefreshCmd,
		RemotePromptRedrawCmd(),
		"ait_precmd()",
		"precmd_functions=(${precmd_functions:#ait_precmd} ait_precmd)",
		"ait_prompt_command()",
		"export PROMPT_COMMAND=ait_prompt_command",
		"zle reset-prompt",
		RemotePromptReadySentinel,
	}
}

// RemoteInitialCommands returns shell setup lines to set PS1/PROMPT after the remote shell starts.
// Returns nil when user or host is empty.
func RemoteInitialCommands(user, host string, dirAbsolute bool) []string {
	if user == "" || host == "" {
		return nil
	}
	return []string{RemoteSetupLine(user, host, dirAbsolute)}
}

// escapeForSingleQuotedShell escapes s for embedding inside a single-quoted sh string.
func escapeForSingleQuotedShell(s string) string {
	if !strings.Contains(s, "'") {
		return s
	}
	return strings.ReplaceAll(s, "'", "'\\''")
}
