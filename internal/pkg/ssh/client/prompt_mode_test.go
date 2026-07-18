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

import "testing"

// enableRemoteSystemPromptForTest forces native remote prompts for this test only.
func enableRemoteSystemPromptForTest(t *testing.T) {
	t.Helper()
	t.Setenv(envRemoteSystemPrompt, "1")
}

// TestUseRemoteSystemPrompt_EnvOverride verifies default AiT mode and env opt-in.
func TestUseRemoteSystemPrompt_EnvOverride(t *testing.T) {
	t.Setenv(envRemoteSystemPrompt, "")
	if UseRemoteSystemPrompt() {
		t.Fatal("default should be AiT-managed prompt (false)")
	}

	t.Setenv(envRemoteSystemPrompt, "1")
	if !UseRemoteSystemPrompt() {
		t.Fatal("AIT_REMOTE_SYSTEM_PROMPT=1 should enable native prompts")
	}

	t.Setenv(envRemoteSystemPrompt, "0")
	if UseRemoteSystemPrompt() {
		t.Fatal("only value \"1\" enables native prompts")
	}
}
