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

package provider

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// TestHints_ConfigSetAPIKeyCmd (C13) keeps API key hint command format stable.
func TestHints_ConfigSetAPIKeyCmd(t *testing.T) {
	got := ConfigSetAPIKeyCmd(Deepseek)
	want := toolname.CLI("config set ai.deepseek.api_key <key>")
	if got != want {
		t.Fatalf("ConfigSetAPIKeyCmd: got %q want %q", got, want)
	}
	if ConfigGetAPIKeyCmd(Minimax) != toolname.CLI("config get ai.minimax.api_key") {
		t.Fatal("ConfigGetAPIKeyCmd mismatch")
	}
	if ProviderEnableCmd(Ollama) != toolname.AICLI("provider enable ollama") {
		t.Fatal("ProviderEnableCmd mismatch")
	}
}

// TestHints_NotConfiguredTryLines (C13) preserves NotConfiguredErr Try block lines.
func TestHints_NotConfiguredTryLines(t *testing.T) {
	err := NotConfiguredErr()
	s := err.Error()
	for _, line := range []string{
		hintSetCloudKeysOrOllama,
		hintSetDeepseekAPIKey,
		hintSetMinimaxAPIKey,
		hintEnableOllama,
		hintAIHelp,
	} {
		if !strings.Contains(s, line) {
			t.Fatalf("NotConfiguredErr missing %q in %q", line, s)
		}
	}
}
