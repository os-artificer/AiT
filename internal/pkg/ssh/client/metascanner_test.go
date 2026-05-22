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

func drive(input string, aiAvailable bool) (forwarded string, echoed string, metaLines []string) {
	s := newMetaScanner(aiAvailable)
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
	fwd, _, meta := drive("!ait local\n", false)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait local" {
		t.Errorf("meta=%v want [!ait local]", meta)
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

func TestMetaScanner_AitAIQuery(t *testing.T) {
	fwd, _, meta := drive("!ait ai how to grep\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai how to grep" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitAIChat(t *testing.T) {
	fwd, _, meta := drive("!ait ai chat ping\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai chat ping" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_BangBangChat(t *testing.T) {
	fwd, _, meta := drive("!! chat ping\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! chat ping" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitLocal(t *testing.T) {
	fwd, _, meta := drive("!ait local\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait local" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitRemote(t *testing.T) {
	fwd, _, meta := drive("!ait remote\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait remote" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitSuspend(t *testing.T) {
	fwd, _, meta := drive("!ait suspend\n", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait suspend" {
		t.Errorf("meta=%v", meta)
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
	for i := 0; i < metaMaxLine+50; i++ {
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
	if !metaPrefixCandidate("!") {
		t.Error(`"!" should be candidate`)
	}
	if !metaPrefixCandidate("!ait") {
		t.Error(`"!ait" should be candidate`)
	}
	if !metaPrefixCandidate("!ait ai") {
		t.Error(`"!ait ai" should be candidate`)
	}
	if metaPrefixCandidate("!ait ai ") {
		t.Error(`exact match "!ait ai " should not be reported as candidate`)
	}
	if metaPrefixCandidate("!zzz") {
		t.Error(`"!zzz" should not be candidate`)
	}
	if _, ok := metaPrefixMatch("!! "); !ok {
		t.Error(`"!! " should match`)
	}
	if _, ok := metaPrefixMatch("!ait remote\n"); !ok {
		t.Error(`"!ait remote\n" should match`)
	}
}

func TestAIPromptFromMetaLine(t *testing.T) {
	tests := []struct {
		line   string
		prompt string
		isAI   bool
	}{
		{"!! how to grep", "how to grep", true},
		{"!!", "", true},
		{"!ait ai how to grep", "how to grep", true},
		{"!ait ai chat hello", "chat hello", true},
		{"!! chat hello", "chat hello", true},
		{"!ait ai", "", true},
		{"!ait suspend", "", false},
		{"!ait local", "", false},
	}
	for _, tc := range tests {
		p, ok := aiPromptFromMetaLine(tc.line)
		if ok != tc.isAI || p != tc.prompt {
			t.Errorf("aiPromptFromMetaLine(%q) = (%q, %v), want (%q, %v)", tc.line, p, ok, tc.prompt, tc.isAI)
		}
	}
}
