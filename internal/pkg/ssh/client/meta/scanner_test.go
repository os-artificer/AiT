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

package meta

import (
	"strings"
	"testing"
)

// drive feeds input one byte at a time and aggregates the scanner outputs.
func drive(input string, aiAvailable bool) (forwarded string, echoed string, metaLines []string) {
	s := New(aiAvailable)
	var fwd, ech strings.Builder
	for i := 0; i < len(input); i++ {
		out, echo, line, ok := s.Feed(input[i])
		fwd.Write(out)
		ech.Write(echo)
		if ok {
			metaLines = append(metaLines, line)
		}
	}
	return fwd.String(), ech.String(), metaLines
}

func TestMetaScanner_BashHistoryWithArgsPassthrough(t *testing.T) {
	fwd, _, meta := drive("!!ls\n", true)
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta lines: %v", meta)
	}
}

func TestMetaScanner_BashHistoryWithArgsCapturedWhenAIDisabled(t *testing.T) {
	fwd, _, meta := drive("!!ls\n", false)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!!ls" {
		t.Errorf("meta=%v want [!!ls]", meta)
	}
}

func TestMetaScanner_AitPrefixCapturedWhenAIDisabled(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh local\n", false)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh local" {
		t.Errorf("meta=%v want [!ai-ssh local]", meta)
	}
}

// TestScanner_MaxLookCoversLongestPrefix ensures holding buf fits "!ai-ssh local ".
func TestScanner_MaxLookCoversLongestPrefix(t *testing.T) {
	if maxLook < len("!ai-ssh local ") {
		t.Fatalf("maxLook=%d shorter than len(!ai-ssh local )", maxLook)
	}
	fwd, _, meta := drive("!ai-ssh local\n", false)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh local" {
		t.Errorf("meta=%v want [!ai-ssh local]", meta)
	}
}

func TestMetaScanner_BashHistoryColonModifierPassthrough(t *testing.T) {
	fwd, _, meta := drive("!!:gs/x/y/\n", true)
	if fwd != "!!:gs/x/y/\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BangBangAloneIsMeta(t *testing.T) {
	fwd, _, meta := drive("!!\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!!" {
		t.Errorf("meta=%v want [!!]", meta)
	}
}

func TestMetaScanner_BangBangSpaceQuery(t *testing.T) {
	fwd, _, meta := drive("!! how to grep\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty (meta capture)", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! how to grep" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_BangBangUTF8Query(t *testing.T) {
	line := "!! 当前系统是什么名字?\n"
	fwd, _, meta := drive(line, true)
	if fwd != "" {
		t.Errorf("forward=%q want empty (meta capture)", fwd)
	}
	want := "!! 当前系统是什么名字?"
	if len(meta) != 1 || meta[0] != want {
		t.Errorf("meta=%v want [%q]", meta, want)
	}
	prompt, ok := AIPromptFromLine(meta[0])
	if !ok || prompt != "当前系统是什么名字?" {
		t.Errorf("prompt=%q ok=%v", prompt, ok)
	}
}

func TestMetaScanner_AitAIQuery(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh ai how to grep\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh ai how to grep" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitAIAsk(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh ai ask ping\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh ai ask ping" {
		t.Errorf("meta=%v", meta)
	}
}

// RA6: bare !ai-ssh ai line is captured as meta and yields empty initial prompt.
func TestMetaScanner_AitAIBareLine(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh ai\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh ai" {
		t.Errorf("meta=%v", meta)
	}
	prompt, ok := AIPromptFromLine(meta[0])
	if !ok || prompt != "" {
		t.Errorf("prompt=%q ok=%v", prompt, ok)
	}
}

func TestMetaScanner_BangBangAsk(t *testing.T) {
	fwd, _, meta := drive("!! ask ping\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! ask ping" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitAIAgent(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh ai agent find large files\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh ai agent find large files" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_BangBangAgent(t *testing.T) {
	fwd, _, meta := drive("!! agent find large files\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! agent find large files" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitLocal(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh local\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh local" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitRemote(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh remote\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh remote" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitSuspend(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh suspend\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh suspend" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_HelpLine(t *testing.T) {
	fwd, _, meta := drive("!ai-ssh help\n", true)
	if len(fwd) != 0 {
		t.Errorf("forwarded %q", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ai-ssh help" {
		t.Errorf("meta=%v want [!ai-ssh help]", meta)
	}
}

func TestMetaScanner_LegacyBackslashAitPassthrough(t *testing.T) {
	fwd, _, meta := drive("\\ait suspend\n", true)
	if fwd != "\\ait suspend\n" {
		t.Errorf("forward=%q want passthrough", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BackslashArtifactPassthrough(t *testing.T) {
	fwd, _, meta := drive("\\artifact\n", true)
	if fwd != "\\artifact\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_NotAtLineStart(t *testing.T) {
	fwd, _, meta := drive("x!!y\n", true)
	if fwd != "x!!y\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_RunawayMetaLineBailsOut(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("!! ")
	for i := 0; i < maxLine+50; i++ {
		sb.WriteByte('x')
	}
	sb.WriteByte('\n')
	fwd, _, meta := drive(sb.String(), true)
	if len(meta) != 0 {
		t.Errorf("expected no meta dispatch for runaway line, got %v", meta)
	}
	if !strings.HasPrefix(fwd, "!! ") || !strings.HasSuffix(fwd, "\n") {
		t.Errorf("runaway line should be flushed verbatim, got len=%d", len(fwd))
	}
}

func TestMetaScanner_TwoMetasInARow(t *testing.T) {
	fwd, _, meta := drive("!! one\n!! two\n", true)
	if fwd != "" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 2 || meta[0] != "!! one" || meta[1] != "!! two" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_MetaLineLocalEcho(t *testing.T) {
	_, echo, meta := drive("!! how\n", true)
	if len(meta) != 1 {
		t.Fatalf("meta=%v", meta)
	}
	if echo != "!! how\n" {
		t.Errorf("echo=%q want %q", echo, "!! how\n")
	}
}

func TestMetaScanner_CtrlUThenBangBang(t *testing.T) {
	fwd, _, meta := drive("\x15!! how\n", true)
	if fwd != "\x15" {
		t.Errorf("forward=%q want \\x15 only", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! how" {
		t.Errorf("meta=%v want [!! how]", meta)
	}
}

func TestMetaScanner_PrefixCandidateHelpers(t *testing.T) {
	if !prefixCandidate("!") {
		t.Error(`"!" should be candidate`)
	}
	if !prefixCandidate("!ai-ssh") {
		t.Error(`"!ai-ssh" should be candidate`)
	}
	if !prefixCandidate("!ai-ssh ai") {
		t.Error(`"!ai-ssh ai" should be candidate`)
	}
	if prefixCandidate("!ai-ssh ai ") {
		t.Error(`exact match "!ai-ssh ai " should not be reported as candidate`)
	}
	if prefixCandidate("!zzz") {
		t.Error(`"!zzz" should not be candidate`)
	}
	if _, ok := prefixMatch("!! "); !ok {
		t.Error(`"!! " should match`)
	}
	if _, ok := prefixMatch("!ai-ssh remote\n"); !ok {
		t.Error(`"!ai-ssh remote\n" should match`)
	}
}

func TestAIPromptFromLine(t *testing.T) {
	tests := []struct {
		line   string
		prompt string
		isAI   bool
	}{
		{"!! how to grep", "how to grep", true},
		{"!!", "", true},
		{"!ai-ssh ai how to grep", "how to grep", true},
		{"!ai-ssh ai ask hello", "ask hello", true},
		{"!! ask hello", "ask hello", true},
		{"!ai-ssh ai agent find large files", "agent find large files", true},
		{"!! agent find large files", "agent find large files", true},
		{"!ai-ssh ai", "", true},
		{"!ai-ssh suspend", "", false},
		{"!ai-ssh local", "", false},
	}
	for _, tc := range tests {
		p, ok := AIPromptFromLine(tc.line)
		if ok != tc.isAI || p != tc.prompt {
			t.Errorf("AIPromptFromLine(%q) = (%q, %v), want (%q, %v)", tc.line, p, ok, tc.prompt, tc.isAI)
		}
	}
}
