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

// TestAIPromptFromMetaLine (R18c) verifies client meta wiring preserves ask one-shot
// InitialQuestion text for !ait ai ask and !! ask meta lines.
func TestAIPromptFromMetaLine(t *testing.T) {
	tests := []struct {
		line   string
		prompt string
		isAI   bool
	}{
		{"!ait ai ask hello", "ask hello", true},
		{"!! ask hello", "ask hello", true},
		{"!ait ai ask ping", "ask ping", true},
		{"!! ask ping", "ask ping", true},
		{"!ait ai how to grep", "how to grep", true},
		{"!ait suspend", "", false},
	}
	for _, tc := range tests {
		p, ok := aiPromptFromMetaLine(tc.line)
		if ok != tc.isAI || p != tc.prompt {
			t.Errorf("aiPromptFromMetaLine(%q) = (%q, %v), want (%q, %v)", tc.line, p, ok, tc.prompt, tc.isAI)
		}
	}
}
