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

import (
	"encoding/json"
	"testing"
)

func TestChatAgent_JSONFallbackContent(t *testing.T) {
	raw := `{"action":"answer","text":"done"}`
	act := struct {
		Action string `json:"action"`
		Text   string `json:"text"`
	}{}
	if err := json.Unmarshal([]byte(raw), &act); err != nil || act.Action != "answer" {
		t.Fatalf("parse: %v action=%q", err, act.Action)
	}
}
