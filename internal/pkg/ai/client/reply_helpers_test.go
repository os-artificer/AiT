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
	"strings"
	"testing"
)

// wrapThinking is a test helper that wraps text in reasoning tags used by
// MiniMax M2.x models (matched by thinkingRe in reply_helpers.go).
func wrapThinking(s string) string {
	return "\u003Cthink\u003E" + s + "\u003C/think\u003E"
}

func TestExtractReply_NoTags(t *testing.T) {
	reply, thinking := ExtractReply("Hello, world!")
	if reply != "Hello, world!" {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_Empty(t *testing.T) {
	reply, thinking := ExtractReply("")
	if reply != "" || thinking != "" {
		t.Fatalf("reply=%q thinking=%q", reply, thinking)
	}
}

func TestExtractReply_WhitespaceOnly(t *testing.T) {
	reply, thinking := ExtractReply("   ")
	if reply != "" || thinking != "" {
		t.Fatalf("reply=%q thinking=%q", reply, thinking)
	}
}

func TestExtractReply_ThinkingOnly(t *testing.T) {
	raw := wrapThinking("I am reasoning")
	reply, thinking := ExtractReply(raw)
	if reply != "" {
		t.Fatalf("reply=%q, want empty", reply)
	}
	if thinking != "I am reasoning" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_ThinkingAndReply(t *testing.T) {
	raw := wrapThinking("Let me think") + "The answer is 42"
	reply, thinking := ExtractReply(raw)
	if reply != "The answer is 42" {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "Let me think" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_ReplyThenThinking(t *testing.T) {
	raw := "The answer is 42" + wrapThinking("I reasoned")
	reply, thinking := ExtractReply(raw)
	if reply != "The answer is 42" {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "I reasoned" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_MultilineThinking(t *testing.T) {
	raw := wrapThinking("Line 1\nLine 2\nLine 3") + "Final answer"
	reply, thinking := ExtractReply(raw)
	if reply != "Final answer" {
		t.Fatalf("reply=%q", reply)
	}
	if !strings.Contains(thinking, "Line 1") || !strings.Contains(thinking, "Line 3") {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_MultipleThinkingBlocks(t *testing.T) {
	raw := wrapThinking("First") + " middle " + wrapThinking("Second") + " end"
	reply, thinking := ExtractReply(raw)
	if reply != "middle  end" {
		t.Fatalf("reply=%q", reply)
	}
	// regex captures the first match
	if thinking != "First" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_ThinkingWithNewlines(t *testing.T) {
	raw := wrapThinking("Step 1\nStep 2") + "\n\nResult"
	reply, thinking := ExtractReply(raw)
	if reply != "Result" {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "Step 1\nStep 2" {
		t.Fatalf("thinking=%q", thinking)
	}
}

// --- XML tool call stripping tests ---

func TestExtractReply_XMLToolCallOnly(t *testing.T) {
	raw := `<minimax:tool_call>` + "\n" +
		`<invoke name="run">` + "\n" +
		`<parameter name="command">uname -a</parameter>` + "\n" +
		`</invoke>` + "\n" +
		`</minimax:tool_call>`
	reply, thinking := ExtractReply(raw)
	if reply != "" {
		t.Fatalf("reply=%q, want empty", reply)
	}
	if thinking != "" {
		t.Fatalf("thinking=%q, want empty (tool calls are not thinking)", thinking)
	}
}

func TestExtractReply_XMLToolCallWithReply(t *testing.T) {
	raw := "I will check the system.\n" +
		`<minimax:tool_call>` + "\n" +
		`<invoke name="run"><parameter name="command">uname -a</parameter></invoke>` + "\n" +
		`</minimax:tool_call>`
	reply, thinking := ExtractReply(raw)
	if reply != "I will check the system." {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "" {
		t.Fatalf("thinking=%q, want empty", thinking)
	}
}

func TestExtractReply_InvokeOnly(t *testing.T) {
	raw := `<invoke name="run"><parameter name="command">hostname</parameter></invoke>`
	reply, _ := ExtractReply(raw)
	if reply != "" {
		t.Fatalf("reply=%q, want empty", reply)
	}
}

func TestExtractReply_ParameterOnly(t *testing.T) {
	raw := `<parameter name="command">hostname</parameter>`
	reply, _ := ExtractReply(raw)
	if reply != "" {
		t.Fatalf("reply=%q, want empty", reply)
	}
}

func TestExtractReply_ParameterWithText(t *testing.T) {
	raw := "Checking now " + `<parameter name="command">hostname</parameter>` + " done"
	reply, _ := ExtractReply(raw)
	if reply != "Checking now  done" {
		t.Fatalf("reply=%q", reply)
	}
}

func TestExtractReply_ThinkingAndXMLToolCall(t *testing.T) {
	raw := wrapThinking("I need to check") + "Checking now\n" +
		`<minimax:tool_call><invoke name="run"><parameter name="command">ls</parameter></invoke></minimax:tool_call>`
	reply, thinking := ExtractReply(raw)
	if reply != "Checking now" {
		t.Fatalf("reply=%q", reply)
	}
	if thinking != "I need to check" {
		t.Fatalf("thinking=%q", thinking)
	}
}

func TestExtractReply_MultipleToolCallsWithReply(t *testing.T) {
	raw := "Let me gather info\n" +
		`<minimax:tool_call><invoke name="run"><parameter name="command">uname</parameter></invoke></minimax:tool_call>` + "\n" +
		`<minimax:tool_call><invoke name="run"><parameter name="command">free</parameter></invoke></minimax:tool_call>` + "\n" +
		"Done collecting."
	reply, _ := ExtractReply(raw)
	if !strings.Contains(reply, "Let me gather info") || !strings.Contains(reply, "Done collecting.") {
		t.Fatalf("reply=%q", reply)
	}
	// Tool call blocks removed, leaving whitespace gaps
	if strings.Contains(reply, "minimax:tool_call") || strings.Contains(reply, "invoke") {
		t.Fatalf("reply still contains XML tags: %q", reply)
	}
}

// M5: ExtractReply strips hybrid JSON+XML fragments so Ask mode does not leak provider format.
func TestExtractReply_HybridFragmentStripped(t *testing.T) {
	raw := `{"action":"run_shell","command">whoami</parameter>` + "\n" +
		`</invoke>` + "\n" +
		`</minimax:tool_call>`
	reply, thinking := ExtractReply(raw)
	if reply != "" {
		t.Fatalf("reply=%q want empty", reply)
	}
	if thinking != "" {
		t.Fatalf("thinking=%q want empty", thinking)
	}
}

func TestExtractReply_PreservesCompleteAgentJSON(t *testing.T) {
	raw := `{"action":"run_shell","command":"echo hi"}`
	reply, _ := ExtractReply(raw)
	if reply != raw {
		t.Fatalf("reply=%q want unchanged JSON", reply)
	}
}
