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

package tmux

import (
	"strings"
	"testing"
)

// TestBuildSendKeysLiteralCmd_Golden (TM9d/TM14a) verifies send-keys -l uses single-quoted args.
func TestBuildSendKeysLiteralCmd_Golden(t *testing.T) {
	name := "ait-test-id"
	literal := "stty -echo 2>/dev/null; export PS1='AiT u@h:'; stty echo 2>/dev/null"
	got := BuildSendKeysLiteralCmd(name, literal)

	if !strings.HasPrefix(got, "tmux send-keys -t 'ait-test-id' -l '") {
		t.Fatalf("want single-quoted session name and -l literal, got %q", got)
	}
	if !strings.Contains(got, "stty -echo") {
		t.Fatalf("want literal cmd in %q", got)
	}
	if strings.Contains(got, `"`) {
		t.Fatalf("send-keys cmd should not use double quotes: %q", got)
	}
}

// TestBuildSendKeysLiteralCmd_PreservesDollarLiterals (TM13a/TM14a) keeps $(...) literal for the pane shell.
func TestBuildSendKeysLiteralCmd_PreservesDollarLiterals(t *testing.T) {
	name := "ait-83080b38-c0b6-47f2-9eca-13a65f395df7"
	literal := "export PS1='$(basename \"${PWD:-/}\")'"
	got := BuildSendKeysLiteralCmd(name, literal)

	if !strings.Contains(got, "$(basename") {
		t.Fatalf("want literal $(basename in %q", got)
	}
	if !strings.HasPrefix(got, "tmux send-keys -t '") {
		t.Fatalf("want cmd wrapped with single-quoted target: %q", got)
	}
}

func TestBuildSendKeyCmd_EnterGolden(t *testing.T) {
	got := BuildSendKeyCmd("ait-test-id", tmuxSendKeysEnter)
	want := "tmux send-keys -t 'ait-test-id' Enter"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildSendKeyCmd_ClearGolden(t *testing.T) {
	got := BuildSendKeyCmd("ait-test-id", tmuxSendKeysClear)
	want := "tmux send-keys -t 'ait-test-id' C-l"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestShSingleQuote_EscapesEmbeddedSingleQuotes(t *testing.T) {
	got := shSingleQuote("export PS1='AiT u@h:'")
	want := "'export PS1='\\''AiT u@h:'\\'''"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
