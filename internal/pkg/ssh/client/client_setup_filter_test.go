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

package client

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

// TestStripSetupLeakStdout_PreservesCRLF (TM11b) keeps original line endings on stdout-bound output.
func TestStripSetupLeakStdout_PreservesCRLF(t *testing.T) {
	user, host := "u", "h"
	leak := "stty -echo 2>/dev/null\r\n"
	prompt := "AiT u@h:/tmp> \r\n"
	out := stripSetupLeakStdout([]byte(leak+prompt), user, host)
	if strings.Contains(string(out), "stty -echo") {
		t.Fatal("leak should be stripped")
	}
	if string(out) != prompt {
		t.Fatalf("want preserved bytes %q, got %q", prompt, out)
	}
}

func TestStripSetupLeakStdout_PreservesLoneLF(t *testing.T) {
	user, host := "u", "h"
	prompt := "AiT u@h:/tmp> \n"
	out := stripSetupLeakStdout([]byte(prompt), user, host)
	if string(out) != prompt {
		t.Fatalf("want %q, got %q", prompt, out)
	}
}

func TestIsSetupLeakLine(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	leaks := []string{
		"stty -echo 2>/dev/null",
		"stty echo 2>/dev/null",
		".__AITPWD__.",
		"AITPWDEND__.",
		`$(d=$(basename "${PWD:-/}" 2>/dev/null); printf '%s' "${d:-/}")`,
		"set +o history 2>/dev/null",
		"[ -n \"$ZSH_VERSION\" ] && _AIT_HIST=$HISTFILE HISTFILE=/dev/null",
		"export PS1='\\[\\]AiT root@remote.example.com:${PWD/#$HOME/\\~}$(b=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) && [ -n \"$b\" ] && printf \":%s\" \"$b\")> \\[\\] '",
		// Wrapped/redrawn PS1 fragments (line-wrap can cut off the basename prefix).
		`"${PWD:-/}" 2>/dev/null); printf '%s' "${d:-/}")$(b=$( 2>/dev/null) && [ -n "$b" ] && printf ":%s" "$b")> \[\]'; [ -n`,
		`printf '%s' "${d:-/}")`,
		`&& printf ":%s" "$b")>`,
		// Lone-character noise lines left by ZLE line redraw.
		"#", "S", "a", "H", " # ",
	}
	for _, line := range leaks {
		if !isSetupLeakLine(line, user, host) {
			t.Errorf("expected leak: %q", line)
		}
	}
	keep := "AiT root@remote.example.com:/data/workspace/AiT:main> "
	if isSetupLeakLine(keep, user, host) {
		t.Errorf("live prompt should not be leak: %q", keep)
	}
}

func TestSetupOutputFilterAcrossChunks(t *testing.T) {
	user, host := "u", "h"
	var f setupOutputFilter

	chunk1 := f.filter([]byte("stty -echo 2>/dev/null\r\n"), user, host)
	if len(chunk1) != 0 {
		t.Fatalf("chunk1 = %q", chunk1)
	}

	setup := "export PS1='\\[\\]AiT u@h:${PWD/#$HOME/\\~}$(git rev-parse"
	chunk2 := f.filter([]byte(setup+"\nAiT u@h:/tmp:main> \n"), user, host)
	out := string(chunk2)
	if strings.Contains(out, "stty -echo") || strings.Contains(out, "export PS1=") {
		t.Fatalf("filtered output still has setup: %q", out)
	}
	if !strings.Contains(out, "AiT u@h:/tmp:main>") {
		t.Fatalf("prompt missing: %q", out)
	}
}

// TestSetupOutputFilter_CRLines (TM9c) verifies setup lines terminated with lone CR are filtered.
func TestSetupOutputFilter_CRLines(t *testing.T) {
	user, host := "u", "h"
	var f setupOutputFilter

	chunk := f.filter([]byte("stty -echo 2>/dev/null\r"), user, host)
	if len(chunk) != 0 {
		t.Fatalf("CR-terminated leak should be consumed: %q", chunk)
	}
	if tail := f.flush(user, host); len(tail) != 0 {
		t.Fatalf("flush after CR line = %q", tail)
	}

	chunk2 := f.filter([]byte("export PS1='AiT u@h:/tmp> '\rAiT u@h:/tmp> "), user, host)
	out := string(chunk2) + string(f.flush(user, host))
	if strings.Contains(out, "export PS1=") {
		t.Fatalf("still has setup: %q", out)
	}
	if !strings.Contains(out, "AiT u@h:/tmp>") {
		t.Fatalf("prompt missing: %q", out)
	}
}

func TestSetupOutputFilterFlush(t *testing.T) {
	var f setupOutputFilter
	user, host := "u", "h"
	_ = f.filter([]byte("stty -echo"), user, host)
	if tail := f.flush(user, host); len(tail) != 0 {
		t.Fatalf("flush leak tail = %q", tail)
	}
}

// TestDuplicatePromptOutput_PrefixMatch (W4) drops duplicate dynamic prompts with different paths.
func TestDuplicatePromptOutput_PrefixMatch(t *testing.T) {
	user, host := "u", "h"
	if !isDuplicatePromptOutput([]byte("AiT u@h:/tmp> "), user, host) {
		t.Fatal("path-only prompt should be duplicate")
	}
	if !isDuplicatePromptOutput([]byte("AiT u@h:/proj:main> "), user, host) {
		t.Fatal("path+branch prompt should be duplicate")
	}
	if isDuplicatePromptOutput([]byte("AiT u@h:/tmp> ls\n"), user, host) {
		t.Fatal("prompt plus trailing text should not be duplicate")
	}
	if !prompt.IsRemotePromptLine(user, host, "AiT u@h:~/x>") {
		t.Fatal("tilde path should match")
	}
}

// TestSetupGraceStripper_PartialChunk (N16) keeps incomplete setup lines across stdout chunks during grace.
func TestSetupGraceStripper_PartialChunk(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	var s setupStdoutStripper

	out1 := s.filterChunk([]byte("stty echo 2>/dev/null"), user, host, true)
	if len(out1) != 0 {
		t.Fatalf("partial leak chunk1 = %q", out1)
	}

	out2 := s.filterChunk([]byte("\r\nAiT root@remote.example.com:/tmp> \r\n"), user, host, true)
	if strings.Contains(string(out2), "stty echo") {
		t.Fatalf("leak in chunk2: %q", out2)
	}
	if !strings.Contains(string(out2), "AiT root@remote.example.com:/tmp>") {
		t.Fatalf("prompt missing: %q", out2)
	}
}

// TestSetupOutputFilter_HoldsNativePendingWithoutNewline (P1c) keeps native prompts out of scrollback during setup.
func TestSetupOutputFilter_HoldsNativePendingWithoutNewline(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	var f setupOutputFilter

	native := "root@remote [~] ➜"
	out := f.filter([]byte(native), user, host)
	if len(out) != 0 {
		t.Fatalf("native pending should stay buffered, filter = %q", out)
	}
	if string(f.pending) != native {
		t.Fatalf("pending = %q, want %q", f.pending, native)
	}
}
