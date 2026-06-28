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
	"testing"
)

func TestNormalizeES6UnicodeJSON_EnablesUnmarshal(t *testing.T) {
	raw := `{"action":"answer","text":"\u2705 report \u{1F4BB} CPU"}`
	normalized := NormalizeES6UnicodeJSON(raw)

	var parsed struct {
		Action string `json:"action"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal([]byte(normalized), &parsed); err != nil {
		t.Fatalf("unmarshal normalized: %v", err)
	}
	if parsed.Action != ToolAnswer {
		t.Fatalf("action=%q want %q", parsed.Action, ToolAnswer)
	}
	if parsed.Text == "" {
		t.Fatal("expected non-empty text")
	}
}

func TestNormalizeES6UnicodeJSON_UnchangedStandardJSON(t *testing.T) {
	raw := `{"action":"run_shell","command":"echo hi"}`
	if got := NormalizeES6UnicodeJSON(raw); got != raw {
		t.Fatalf("got=%q want unchanged", got)
	}
}
