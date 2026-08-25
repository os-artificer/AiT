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

func TestToolDefinitions_NamesAndJSON(t *testing.T) {
	defs := ToolDefinitions()
	if len(defs) != 5 {
		t.Fatalf("len=%d want 5", len(defs))
	}
	want := map[string]string{
		ToolRunShell: `{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`,
		ToolAISshCLI: `{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`,
		ToolAitCLI:   `{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`,
		ToolToolbox:  `{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`,
		ToolAnswer:   `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`,
	}
	for _, d := range defs {
		golden, ok := want[d.Name]
		if !ok {
			t.Fatalf("unexpected tool %q", d.Name)
		}
		if d.ParamJSON != golden {
			t.Fatalf("tool %q ParamJSON=%q want %q", d.Name, d.ParamJSON, golden)
		}
		if d.ParamName == "" || d.Desc == "" {
			t.Fatalf("tool %q missing metadata", d.Name)
		}
	}
}
