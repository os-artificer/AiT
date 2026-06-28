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

func TestAppendUserTurn(t *testing.T) {
	h := []ChatMessage{{Role: "user", Content: "a"}}
	out := AppendUserTurn(h, "b")
	if len(out) != 2 || out[1].Content != "b" {
		t.Fatalf("out=%+v", out)
	}
	if len(h) != 1 {
		t.Fatal("must not mutate input")
	}
}

func TestReplaceLastAssistant(t *testing.T) {
	updated := []ChatMessage{
		{Role: "user", Content: "q"},
		{Role: "assistant", Content: "raw malformed"},
	}
	out := ReplaceLastAssistant(updated, `{"action":"answer","text":"ok"}`)
	if out[1].Content != `{"action":"answer","text":"ok"}` {
		t.Fatalf("assistant=%q", out[1].Content)
	}
}

func TestCloneHistory(t *testing.T) {
	h := []ChatMessage{{Role: "user", Content: "x"}}
	c := CloneHistory(h)
	c[0].Content = "y"
	if h[0].Content != "x" {
		t.Fatal("clone must be independent")
	}
}
