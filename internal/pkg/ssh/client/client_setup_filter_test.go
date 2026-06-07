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

func TestIsSetupLeakLine(t *testing.T) {
	user, host := "root", "ylgeeker.com"
	leaks := []string{
		"stty -echo 2>/dev/null",
		"set +o history 2>/dev/null",
		"[ -n \"$ZSH_VERSION\" ] && _AIT_HIST=$HISTFILE HISTFILE=/dev/null",
		"export PS1='\\[\\]AiT root@ylgeeker.com:${PWD/#$HOME/\\~}$(b=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) && [ -n \"$b\" ] && printf \":%s\" \"$b\")> \\[\\] '",
	}
	for _, line := range leaks {
		if !isSetupLeakLine(line, user, host) {
			t.Errorf("expected leak: %q", line)
		}
	}
	keep := "AiT root@ylgeeker.com:/data/workspace/AiT:main> "
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
