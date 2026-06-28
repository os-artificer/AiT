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

import (
	"encoding/json"
	"strings"
	"testing"
)

// M1: well-formed MiniMax XML invoke converts to run_shell JSON.
func TestXMLToolCallToJSON_WellFormedInvoke(t *testing.T) {
	raw := `<minimax:tool_call>` + "\n" +
		`<invoke name="run">` + "\n" +
		`<parameter name="command">uname -a</parameter>` + "\n" +
		`</invoke>` + "\n" +
		`</minimax:tool_call>`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	var act struct {
		Action  string `json:"action"`
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(got), &act); err != nil {
		t.Fatalf("unmarshal: %v got=%q", err, got)
	}
	if act.Action != ToolRunShell || act.Command != "uname -a" {
		t.Fatalf("got=%q action=%q command=%q", got, act.Action, act.Command)
	}
}

// M2: hybrid JSON+XML fragment from MiniMax Agent bug converts to run_shell JSON.
func TestXMLToolCallToJSON_HybridMalformed(t *testing.T) {
	raw := `{"action":"run_shell","command">whoami</parameter>` + "\n" +
		`</invoke>` + "\n" +
		`</minimax:tool_call>`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(got, `"action":"run_shell"`) || !strings.Contains(got, "whoami") {
		t.Fatalf("got %q", got)
	}
}

// M3: invoke name="run" maps to run_shell action.
func TestXMLToolCallToJSON_RunAlias(t *testing.T) {
	raw := `<invoke name="run"><parameter name="command">hostname</parameter></invoke>`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(got, `"action":"run_shell"`) {
		t.Fatalf("got %q", got)
	}
}

// M5: truncated MiniMax XML without closing tags still converts to run_shell JSON.
func TestXMLToolCallToJSON_OpenInvokeTruncated(t *testing.T) {
	raw := `<minimax:tool_call>
<invoke name="run_shell">
<parameter name="command">echo "=== CPU Info ===" && lscpu && free -h}`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok for open invoke/parameter")
	}
	if !strings.Contains(got, `"action":"run_shell"`) || !strings.Contains(got, "lscpu") {
		t.Fatalf("got %q", got)
	}
}

func TestXMLToolCallToJSON_PureJSONReturnsFalse(t *testing.T) {
	raw := `{"action":"answer","text":"done"}`
	if _, ok := XMLToolCallToJSON(raw); ok {
		t.Fatal("pure JSON should not match XML parser")
	}
}

func TestXMLToolCallToJSON_AnswerInvoke(t *testing.T) {
	raw := `<invoke name="answer"><parameter name="text">系统配置正常</parameter></invoke>`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(got, `"action":"answer"`) || !strings.Contains(got, "系统配置正常") {
		t.Fatalf("got %q", got)
	}
}

// M4: invoke name="aish" maps to canonical aish_cli action (same as FC ToolCallToJSON).
func TestXMLToolCallToJSON_AISshInvoke(t *testing.T) {
	raw := `<invoke name="aish"><parameter name="args">session --list</parameter></invoke>`
	got, ok := XMLToolCallToJSON(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(got, `"action":"`+ToolAISshCLI+`"`) || !strings.Contains(got, "session --list") {
		t.Fatalf("got %q", got)
	}
}
