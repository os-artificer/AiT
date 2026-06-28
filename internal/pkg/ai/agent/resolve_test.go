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

// AF2: multiline answer JSON prefix yields markdown body only.
func TestResolveAgentReply_MultilineAnswerPrefix(t *testing.T) {
	raw := `{"action":"answer","text":"# Hardware Report` + "\n\n## CPU\n- 8 cores"
	result := ResolveAgentReply(raw)
	if !result.OK() {
		t.Fatalf("expected ok, got errs=%v", result.Errs)
	}
	if result.Source != SourceHeuristic {
		t.Fatalf("source=%q want heuristic", result.Source)
	}
	if !result.Envelope.IsAnswer() {
		t.Fatalf("envelope=%+v", result.Envelope)
	}
	if !strings.HasPrefix(result.Envelope.Text, "# Hardware Report") {
		t.Fatalf("text=%q", result.Envelope.Text)
	}
	if strings.Contains(result.Envelope.Text, `"action"`) {
		t.Fatal("text must not contain JSON prefix")
	}
}

// AF3: pure prose without action envelope becomes answer.
func TestResolveAgentReply_ProseAsAnswer(t *testing.T) {
	result := ResolveAgentReply("Here is a plain answer without JSON.")
	if !result.OK() || result.Source != SourceProse {
		t.Fatalf("result=%+v", result)
	}
	if result.Envelope.Text != "Here is a plain answer without JSON." {
		t.Fatalf("text=%q", result.Envelope.Text)
	}
}

// AF6: ValidateEnvelope rejects empty run command.
func TestValidateEnvelope_EmptyRunCommand(t *testing.T) {
	err := ValidateEnvelope(AgentEnvelope{Action: ToolRunShell, Command: ""})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

// AF6: ValidateEnvelope rejects empty ait args.
func TestValidateEnvelope_EmptyAitArgs(t *testing.T) {
	err := ValidateEnvelope(AgentEnvelope{Action: ToolAISshCLI, Args: ""})
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

// AF9: broken run envelope must not prose-fallback.
func TestResolveAgentReply_RunSignalNoProseFallback(t *testing.T) {
	raw := `{"action":"run_shell","command":"hostname`
	result := ResolveAgentReply(raw)
	if result.OK() {
		t.Fatalf("expected failure, got %+v", result)
	}
	if result.Source == SourceProse {
		t.Fatal("must not prose-fallback when run envelope signal present")
	}
}

// AF12: markdown sample with "command": still prose→answer when no envelope signal.
func TestResolveAgentReply_MarkdownCommandSampleProseOK(t *testing.T) {
	raw := "# Config sample\n\n```json\n{\"command\":\"ls -la\"}\n```\n\nUse with care."
	result := ResolveAgentReply(raw)
	if !result.OK() || result.Source != SourceProse {
		t.Fatalf("result=%+v", result)
	}
}

func TestResolveAgentReply_StrictJSON(t *testing.T) {
	result := ResolveAgentReply(`{"action":"answer","text":"hello"}`)
	if !result.OK() || result.Source != SourceJSON {
		t.Fatalf("result=%+v", result)
	}
}

func TestAgentEnvelope_CanonicalJSON(t *testing.T) {
	env := AgentEnvelope{Action: ToolAnswer, Text: "line1\nline2"}
	canonical, err := env.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(canonical, "\n") {
		t.Fatalf("canonical must be single line: %q", canonical)
	}
	re := ResolveAgentReply(canonical)
	if !re.OK() || re.Envelope.Text != "line1\nline2" {
		t.Fatalf("round-trip failed: %+v", re)
	}
}

func TestResolveAgentReply_InvalidVerbFails(t *testing.T) {
	result := ResolveAgentReply(`{"action":"bogus","text":"x"}`)
	if result.OK() {
		t.Fatalf("expected failure, got %+v", result)
	}
}

func TestFormatCorrectionPrompt_IncludesOriginal(t *testing.T) {
	prompt := FormatCorrectionPrompt("analyze hardware", `{"bad":`, []error{errTest})
	if !strings.Contains(prompt, "analyze hardware") {
		t.Fatalf("missing original question: %q", prompt)
	}
}

var errTest = &testErr{"invalid json"}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
