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

package agent

import "testing"

// A20i / A20m underlying: ParseAction extracts a run action.
func TestParseAction_Run(t *testing.T) {
	act, ok := ParseAction(`{"action":"run_shell","command":"hostname"}`)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if !act.IsRun() {
		t.Fatalf("expected run, got %+v", act)
	}
	if act.Command != "hostname" {
		t.Fatalf("command=%q", act.Command)
	}
}

// ParseAction accepts the canonical "run" verb in addition to "run_shell".
func TestParseAction_RunCanonical(t *testing.T) {
	act, ok := ParseAction(`{"action":"run","command":"echo hi"}`)
	if !ok || !act.IsRun() {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
}

// ParseAction recognises aish, aish_cli, and legacy ait/ait_cli verbs.
func TestParseAction_Ait(t *testing.T) {
	for _, verb := range []string{"aish", "aish_cli", "ait", "ait_cli"} {
		act, ok := ParseAction(`{"action":"` + verb + `","args":"list"}`)
		if !ok || !act.IsAit() {
			t.Fatalf("verb=%q act=%+v ok=%v", verb, act, ok)
		}
		if act.Args != "list" {
			t.Fatalf("args=%q", act.Args)
		}
	}
}

// ParseAction recognises the terminal answer verb.
func TestParseAction_Answer(t *testing.T) {
	act, ok := ParseAction(`{"action":"answer","text":"hello"}`)
	if !ok || !act.IsAnswer() {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
	if act.Text != "hello" {
		t.Fatalf("text=%q", act.Text)
	}
}

// ParseAction accepts ES6 \u{XXXX} escapes inside provider JSON answer payloads.
func TestParseAction_ES6UnicodeAnswer(t *testing.T) {
	raw := `{"action":"answer","text":"\u2705 report \u{1F4BB} CPU"}`
	act, ok := ParseAction(raw)
	if !ok || !act.IsAnswer() {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
	if act.Text == "" {
		t.Fatal("expected decoded answer text")
	}
}

// ParseAction returns false for empty or unparseable input.
func TestParseAction_Empty(t *testing.T) {
	for _, in := range []string{"", "   ", "not json"} {
		if _, ok := ParseAction(in); ok {
			t.Fatalf("expected !ok for %q", in)
		}
	}
}

// ParseAction tolerates surrounding text and extracts the JSON object.
func TestParseAction_EmbeddedJSON(t *testing.T) {
	in := "let me think...\n{\"action\":\"answer\",\"text\":\"done\"}\nokay"
	act, ok := ParseAction(in)
	if !ok || !act.IsAnswer() || act.Text != "done" {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
}

// Empty Action verb is rejected.
func TestParseAction_EmptyVerb(t *testing.T) {
	if _, ok := ParseAction(`{"action":"","text":"x"}`); ok {
		t.Fatal("expected !ok for empty action verb")
	}
}

// RA5: hybrid MiniMax XML fragment falls back to run_shell when JSON parse fails.
func TestParseAction_XMLFallback(t *testing.T) {
	raw := `{"action":"run_shell","command">whoami</parameter>` + "\n" +
		`</invoke>` + "\n" +
		`</minimax:tool_call>`
	act, ok := ParseAction(raw)
	if !ok {
		t.Fatal("expected parse ok via XML fallback")
	}
	if !act.IsRun() || act.Command != "whoami" {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
}
