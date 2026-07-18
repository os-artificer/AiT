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

package session

import (
	"reflect"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

// TestRemotePromptInitialCommands_LoginLoadSamePath (W10): connect and load share the same
// non-empty AiT PS1 setup when session.remote_prompt.path is absolute.
func TestRemotePromptInitialCommands_LoginLoadSamePath(t *testing.T) {
	m := newTestManager(t)
	if err := m.Config.SetKey("session.remote_prompt.path", config.RemotePromptPathAbsolute); err != nil {
		t.Fatal(err)
	}

	loginCmds := m.remotePromptInitialCommands("u", "h")
	loadCmds := m.remotePromptInitialCommands("u", "h")
	if len(loginCmds) == 0 {
		t.Fatal("expected non-empty AiT PS1 setup for login/load")
	}
	if !reflect.DeepEqual(loginCmds, loadCmds) {
		t.Fatalf("login/load cmds differ:\nlogin=%v\nload=%v", loginCmds, loadCmds)
	}
	joined := strings.Join(loginCmds, "\n")
	if !strings.Contains(joined, "${PWD/#$HOME") && !strings.Contains(joined, "PWD") {
		t.Fatalf("absolute path mode should mention PWD expansion, got %q", joined)
	}
}

// TestRemotePromptInitialCommands_NativeModeEmpty verifies AIT_REMOTE_SYSTEM_PROMPT=1
// skips PS1 injection for both login and load.
func TestRemotePromptInitialCommands_NativeModeEmpty(t *testing.T) {
	t.Setenv("AIT_REMOTE_SYSTEM_PROMPT", "1")
	m := newTestManager(t)
	loginCmds := m.remotePromptInitialCommands("u", "h")
	loadCmds := m.remotePromptInitialCommands("u", "h")
	if len(loginCmds) != 0 || len(loadCmds) != 0 {
		t.Fatalf("native prompt mode should inject no PS1 setup, login=%v load=%v", loginCmds, loadCmds)
	}
}
