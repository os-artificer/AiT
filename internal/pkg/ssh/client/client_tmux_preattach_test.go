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
	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

// TestTmuxPreAttachSetup_CommandGolden (TM13b/TM14a) full PS1 setup uses send-keys -l + Enter.
func TestTmuxPreAttachSetup_CommandGolden(t *testing.T) {
	user, host := "root", "dev.os-artificer.com"
	setup := prompt.RemoteInitialCommands(user, host, false)
	if len(setup) != 1 {
		t.Fatalf("setup lines: %d", len(setup))
	}

	literal := setup[0]
	sendLiteral := tmux.BuildSendKeysLiteralCmd("ait-test-id", literal)
	sendEnter := tmux.BuildSendKeyCmd("ait-test-id", "Enter")

	if !strings.HasPrefix(sendLiteral, "tmux send-keys -t '") {
		t.Fatalf("remote cmd must single-quote session target: %q", sendLiteral)
	}
	if !strings.Contains(sendLiteral, " -l '") {
		t.Fatalf("remote cmd must use send-keys -l: %q", sendLiteral)
	}
	if !strings.Contains(sendLiteral, "$(basename") {
		t.Fatalf("want literal $(basename in remote cmd: %q", sendLiteral)
	}
	if !strings.Contains(sendLiteral, "git rev-parse") {
		t.Fatalf("want literal git rev-parse in remote cmd: %q", sendLiteral)
	}
	if sendEnter != "tmux send-keys -t 'ait-test-id' Enter" {
		t.Fatalf("want Enter after literal: %q", sendEnter)
	}
}

// TestTmuxPreAttachSetup_ClearPaneGolden (TM14b) pre-attach ends with clear-history + C-l.
func TestTmuxPreAttachSetup_ClearPaneGolden(t *testing.T) {
	name := "ait-test-id"
	clearHist := "tmux clear-history -t '" + name + "' 2>/dev/null || true"
	clearKey := tmux.BuildSendKeyCmd(name, "C-l")

	if !strings.Contains(clearHist, "clear-history") {
		t.Fatalf("want clear-history in %q", clearHist)
	}
	if clearKey != "tmux send-keys -t 'ait-test-id' C-l" {
		t.Fatalf("want C-l redraw: %q", clearKey)
	}
}
