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

package sshclient

import (
	"strings"
	"testing"
)

func TestRemotePrompt(t *testing.T) {
	got := RemotePrompt("u", "192.168.1.1")
	want := "AiT u@192.168.1.1> "
	if got != want {
		t.Errorf("RemotePrompt = %q, want %q", got, want)
	}
}

func TestRemotePromptInitialCommands(t *testing.T) {
	cmds := RemotePromptInitialCommands("u", "h")
	if len(cmds) != 1 {
		t.Fatalf("cmds = %v", cmds)
	}
	for _, part := range []string{"export PS1=", "u@h", "PROMPT="} {
		if !strings.Contains(cmds[0], part) {
			t.Errorf("export line missing %q: %s", part, cmds[0])
		}
	}
	if len(RemotePromptInitialCommands("", "h")) != 0 {
		t.Error("empty user should return nil")
	}
}
