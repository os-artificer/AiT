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

package eino

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"
)

// TestSystemPrompt_OpsDomain (OA8) uses Ops-ReAct prompt with toolbox when Domain=ops.
func TestSystemPrompt_OpsDomain(t *testing.T) {
	meta := apitypes.PromptMeta{Provider: "deepseek", Model: "m"}
	got := AgentSystemPrompt(apitypes.ScopeLocal, apitypes.AssistEnv{Domain: "ops", AgentMode: true}, meta)
	for _, want := range []string{
		"Ops Agent",
		"observe-first",
		agent.ToolToolbox,
		"aihost",
		"ailog",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in ops prompt:\n%s", want, got)
		}
	}
	def := AgentSystemPrompt(apitypes.ScopeLocal, apitypes.AssistEnv{AgentMode: true}, meta)
	if strings.Contains(def, "Ops Agent") {
		t.Fatal("default domain should not use Ops Agent prompt")
	}
}
