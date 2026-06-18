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

package prompt

import (
	"strings"
	"testing"
)

func TestRemotePrompt(t *testing.T) {
	got := Remote("u", "192.168.1.1")
	want := "AiT u@192.168.1.1:> "
	if got != want {
		t.Errorf("Remote = %q, want %q", got, want)
	}
}

// TestRemoteFormat_WithAndWithoutBranch (W1) verifies RemoteFormat with and without git branch.
func TestRemoteFormat_WithAndWithoutBranch(t *testing.T) {
	with := RemoteFormat("root", "h.com", "/data/AiT", "main")
	wantWith := "AiT root@h.com:/data/AiT:main> "
	if with != wantWith {
		t.Fatalf("with branch = %q, want %q", with, wantWith)
	}

	without := RemoteFormat("root", "h.com", "/tmp", "")
	wantWithout := "AiT root@h.com:/tmp> "
	if without != wantWithout {
		t.Fatalf("without branch = %q, want %q", without, wantWithout)
	}
}

// TestRemoteSetupMarker_Prefix (W2) verifies readiness marker is prefix only.
func TestRemoteSetupMarker_Prefix(t *testing.T) {
	got := RemoteSetupMarker("u", "h")
	want := "AiT u@h:"
	if got != want {
		t.Fatalf("marker = %q, want %q", got, want)
	}
	if strings.Contains(got, ">") {
		t.Fatal("marker should not include closing >")
	}
}

// TestRemoteSetupLine_ContainsDirAndGit (W3) verifies dynamic PS1 includes cwd and git (default short).
func TestRemoteSetupLine_ContainsDirAndGit(t *testing.T) {
	line := RemoteSetupLine("u", "h", false)
	for _, part := range []string{
		"AiT u@h:",
		`basename "${PWD:-/}"`,
		"git rev-parse --abbrev-ref HEAD",
		"PROMPT_SUBST",
		"export PS1=",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing %q: %s", part, line)
		}
	}
}

// TestRemoteSetupLine_ShortDir (W7) verifies short mode uses basename for cwd segment.
func TestRemoteSetupLine_ShortDir(t *testing.T) {
	line := RemoteSetupLine("u", "h", false)
	if !strings.Contains(line, `basename "${PWD:-/}"`) {
		t.Fatalf("short setup should use basename: %s", line)
	}
	if strings.Contains(line, "${PWD/#$HOME") {
		t.Fatalf("short setup should not use absolute PWD substitution: %s", line)
	}
}

// TestRemoteSetupLine_AbsoluteDir (W8) verifies absolute mode uses $HOME tilde substitution.
func TestRemoteSetupLine_AbsoluteDir(t *testing.T) {
	line := RemoteSetupLine("u", "h", true)
	if !strings.Contains(line, "${PWD/#$HOME") {
		t.Fatalf("absolute setup should use PWD/#$HOME: %s", line)
	}
	if strings.Contains(line, "basename") {
		t.Fatalf("absolute setup should not use basename: %s", line)
	}
}

func TestIsRemotePromptLine(t *testing.T) {
	if !IsRemotePromptLine("u", "h", "AiT u@h:/tmp:main>") {
		t.Fatal("expected git prompt line")
	}
	if !IsRemotePromptLine("u", "h", "AiT u@h:~/proj> ") {
		t.Fatal("expected path prompt line")
	}
	if IsRemotePromptLine("u", "h", "ls -la") {
		t.Fatal("command line should not match")
	}
	if IsRemotePromptLine("u", "h", "AiT u@h>") {
		t.Fatal("legacy static prompt should not match without colon path")
	}
}

func TestRemoteInitialCommands(t *testing.T) {
	cmds := RemoteInitialCommands("u", "h", false)
	if len(cmds) != 1 {
		t.Fatalf("cmds = %v, want single setup line", cmds)
	}
	line := cmds[0]
	for _, part := range []string{
		"set +o history",
		"_AIT_HIST",
		"HISTFILE=/dev/null",
		"export PS1=", "u@h:", "PROMPT=", "033[1A",
		"HISTCONTROL=ignorespace", "HIST_IGNORE_SPACE",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing %q: %s", part, line)
		}
	}
	if strings.Contains(line, "stty -echo") {
		t.Errorf("setup line should not embed stty (Inject provides echo suppression): %s", line)
	}
	if strings.Contains(line, "__ait_p=${PS1@P}") {
		t.Errorf("setup line should not manually print bash PS1 (breaks readline): %s", line)
	}
	if !strings.Contains(line, `print -Pn "$PROMPT"`) {
		t.Errorf("setup line should print expanded zsh PROMPT: %s", line)
	}
	if !strings.Contains(line, `[ -n "$ZSH_VERSION" ] && printf '\033[1A\033[2K\r'`) {
		t.Errorf("setup line should guard line-clear for zsh only: %s", line)
	}
	if len(RemoteInitialCommands("", "h", false)) != 0 {
		t.Error("empty user should return nil")
	}
}

// TestRemoteSetupLine_HistoryOffWrapper (W17) verifies setup disables bash/zsh history during inject.
func TestRemoteSetupLine_HistoryOffWrapper(t *testing.T) {
	line := RemoteSetupLine("u", "h", false)
	for _, part := range []string{
		"set +o history",
		"set -o history",
		"_AIT_HIST",
		"HISTFILE=/dev/null",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing history-off %q: %s", part, line)
		}
	}
}

// TestRemoteSetupLine_PrintsExpandedPrompt (P1) verifies setup ends with zsh redraw and sentinel.
func TestRemoteSetupLine_PrintsExpandedPrompt(t *testing.T) {
	line := RemoteSetupLine(testSSHUser, testPrimaryHostFQDN, false)
	for _, part := range []string{
		`print -Pn "$PROMPT"`,
		"zle reset-prompt",
		RemotePromptReadySentinel,
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing prompt-print %q: %s", part, line)
		}
	}
	if strings.Contains(line, "__ait_p=${PS1@P}") {
		t.Errorf("bash setup must not manually print PS1: %s", line)
	}
}

// TestRemotePromptRedrawCmd_BashNoOp (P2) verifies bash refresh uses a completing no-op.
func TestRemotePromptRedrawCmd_BashNoOp(t *testing.T) {
	line := RemotePromptRedrawCmd()
	if !strings.Contains(line, `[ -z "$ZSH_VERSION" ] && true`) {
		t.Fatalf("redraw cmd missing bash no-op: %s", line)
	}
}

// TestRemoteSetupLine_ZshPrecmdHook (P1b) verifies zsh setup installs a precmd hook after theme init.
func TestRemoteSetupLine_ZshPrecmdHook(t *testing.T) {
	line := RemoteSetupLine(testSSHUser, testPrimaryHostFQDN, false)
	for _, part := range []string{
		"ait_precmd()",
		"precmd_functions=(${precmd_functions:#ait_precmd} ait_precmd)",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing zsh precmd %q: %s", part, line)
		}
	}
}

// TestRemoteSetupLine_BashPromptCommandHook (P3) verifies bash setup installs PROMPT_COMMAND persistence.
func TestRemoteSetupLine_BashPromptCommandHook(t *testing.T) {
	line := RemoteSetupLine(testSSHUser, testPrimaryHostFQDN, false)
	for _, part := range []string{
		"ait_prompt_command()",
		"export PROMPT_COMMAND=ait_prompt_command",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing bash PROMPT_COMMAND %q: %s", part, line)
		}
	}
}

func TestRemoteSetupStripFragments(t *testing.T) {
	frags := RemoteSetupStripFragments("u", "h", false)
	if len(frags) < 2 {
		t.Fatalf("fragments = %v", frags)
	}
	if !strings.Contains(frags[0], "stty -echo") {
		t.Errorf("first fragment = %q", frags[0])
	}
}
