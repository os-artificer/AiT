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

// TestRemoteSetupLine_ContainsDirAndGit (W3) verifies dynamic PS1 includes cwd and git.
func TestRemoteSetupLine_ContainsDirAndGit(t *testing.T) {
	line := RemoteSetupLine("u", "h")
	for _, part := range []string{
		"AiT u@h:",
		"${PWD/#$HOME/",
		"git rev-parse --abbrev-ref HEAD",
		"PROMPT_SUBST",
		"export PS1=",
	} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing %q: %s", part, line)
		}
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
	cmds := RemoteInitialCommands("u", "h")
	if len(cmds) != 1 {
		t.Fatalf("cmds = %v, want single setup line", cmds)
	}
	line := cmds[0]
	for _, part := range []string{"stty -echo", "export PS1=", "u@h:", "PROMPT=", "033[1A", "stty echo", "HISTCONTROL=ignorespace", "HIST_IGNORE_SPACE"} {
		if !strings.Contains(line, part) {
			t.Errorf("setup line missing %q: %s", part, line)
		}
	}
	if strings.Contains(line, "printf '\\033[1A\\033[2K\\r%s'") {
		t.Errorf("setup line should not printf display: %s", line)
	}
	if len(RemoteInitialCommands("", "h")) != 0 {
		t.Error("empty user should return nil")
	}
}

func TestRemoteSetupStripFragments(t *testing.T) {
	frags := RemoteSetupStripFragments("u", "h")
	if len(frags) < 2 {
		t.Fatalf("fragments = %v", frags)
	}
	if !strings.Contains(frags[0], "stty -echo") {
		t.Errorf("first fragment = %q", frags[0])
	}
}
