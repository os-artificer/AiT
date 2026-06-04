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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

// TestRemotePromptInitialCommands_LoginLoadSamePath (W10) verifies login and load share one helper.
func TestRemotePromptInitialCommands_LoginLoadSamePath(t *testing.T) {
	m := newTestManager(t)
	if err := m.Config.SetKey("session.remote_prompt.path", config.RemotePromptPathAbsolute); err != nil {
		t.Fatal(err)
	}

	loginCmds := m.remotePromptInitialCommands("u", "h")
	loadCmds := m.remotePromptInitialCommands("u", "h")
	if len(loginCmds) != 1 || len(loadCmds) != 1 {
		t.Fatalf("cmds login=%v load=%v", loginCmds, loadCmds)
	}
	if loginCmds[0] != loadCmds[0] {
		t.Fatalf("login and load differ:\nlogin=%q\nload=%q", loginCmds[0], loadCmds[0])
	}
	if !strings.Contains(loginCmds[0], "${PWD/#$HOME") {
		t.Fatalf("absolute mode should inject PWD substitution: %s", loginCmds[0])
	}
}
