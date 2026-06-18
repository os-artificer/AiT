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
	"strings"
	"testing"
)

func TestToolCallToAgentJSON_RunShell(t *testing.T) {
	tc := FCCall{
		Name:      ToolRunShell,
		Arguments: `{"command":"hostname -I"}`,
	}
	got, err := ToolCallToJSON(tc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"action":"run_shell"`) || !strings.Contains(got, "hostname -I") {
		t.Fatalf("got %q", got)
	}
}

func TestChatAgent_ToolCall(t *testing.T) {
	tc := FCCall{
		Name:      ToolAnswer,
		Arguments: `{"text":"ok"}`,
	}
	got, err := ToolCallToJSON(tc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"action":"answer"`) {
		t.Fatalf("got %q", got)
	}
}
