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
	want := "AiT u@192.168.1.1> "
	if got != want {
		t.Errorf("Remote = %q, want %q", got, want)
	}
}

func TestRemoteInitialCommands(t *testing.T) {
	cmds := RemoteInitialCommands("u", "h")
	if len(cmds) != 1 {
		t.Fatalf("cmds = %v, want single setup line", cmds)
	}
	line := cmds[0]
	for _, part := range []string{"stty -echo", "export PS1=", "u@h", "PROMPT=", "033[1A", "stty echo", "HISTCONTROL=ignorespace", "HIST_IGNORE_SPACE"} {
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
