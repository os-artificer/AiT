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

package sshclient

import (
	"strings"
	"testing"
)

// drive feeds the whole input through a fresh scanner and returns the
// concatenated forwarded bytes plus all completed meta lines.
func drive(input string) (forwarded string, metaLines []string) {
	s := newMetaScanner()
	var fwd strings.Builder
	for i := 0; i < len(input); i++ {
		out, line, ok := s.Feed(input[i])
		fwd.Write(out)
		if ok {
			metaLines = append(metaLines, line)
		}
	}
	return fwd.String(), metaLines
}

func TestMetaScanner_BashHistoryWithArgsPassthrough(t *testing.T) {
	// `!!ls` is bash's history-and-append shortcut and must NOT be meta.
	fwd, meta := drive("!!ls\n")
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta lines: %v", meta)
	}
}

func TestMetaScanner_BashHistoryColonModifierPassthrough(t *testing.T) {
	fwd, meta := drive("!!:gs/x/y/\n")
	if fwd != "!!:gs/x/y/\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BangBangAloneIsMeta(t *testing.T) {
	// `!!<Enter>` enters meta (empty prompt). The bytes are NOT forwarded.
	fwd, meta := drive("!!\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!!" {
		t.Errorf("meta=%v want [!!]", meta)
	}
}

func TestMetaScanner_BangBangSpaceQuery(t *testing.T) {
	fwd, meta := drive("!! how to grep\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty (meta capture)", fwd)
	}
	if len(meta) != 1 || meta[0] != "!! how to grep" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_BackslashAStartedWordPassthrough(t *testing.T) {
	// `\artifact` must not be eaten. The opener is `\ait ` (with space).
	fwd, meta := drive("\\artifact\n")
	if fwd != "\\artifact\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_AitSuspend(t *testing.T) {
	fwd, meta := drive("\\ait suspend\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != `\ait suspend` {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_MidVimBackslashAfterNewline(t *testing.T) {
	// Inside vim insert mode, user types Enter then types `\a` then `bc<Enter>`.
	// Because we only track LOCAL atLineStart, this DOES still enter holding;
	// but since `\abc<Enter>` is not a strict prefix of any opener, it must
	// be flushed and forwarded verbatim before any meta dispatch happens.
	fwd, meta := drive("\\abc\n")
	if fwd != "\\abc\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_NotAtLineStart(t *testing.T) {
	// Typing `x!!y` (after `x`, atLineStart=false) must pass through.
	fwd, meta := drive("x!!y\n")
	if fwd != "x!!y\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_RunawayMetaLineBailsOut(t *testing.T) {
	// A meta line longer than metaMaxLine is bailed out: bytes are flushed
	// and no meta event fires.
	var sb strings.Builder
	sb.WriteString("!! ")
	for i := 0; i < metaMaxLine+50; i++ {
		sb.WriteByte('x')
	}
	sb.WriteByte('\n')
	fwd, meta := drive(sb.String())
	if len(meta) != 0 {
		t.Errorf("expected no meta dispatch for runaway line, got %v", meta)
	}
	if !strings.HasPrefix(fwd, "!! ") || !strings.HasSuffix(fwd, "\n") {
		t.Errorf("runaway line should be flushed verbatim, got len=%d", len(fwd))
	}
}

func TestMetaScanner_TwoMetasInARow(t *testing.T) {
	fwd, meta := drive("!! one\n!! two\n")
	if fwd != "" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 2 || meta[0] != "!! one" || meta[1] != "!! two" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_PrefixCandidateHelpers(t *testing.T) {
	if !metaPrefixCandidate("\\") {
		t.Error(`"\\" should be candidate`)
	}
	if !metaPrefixCandidate("\\ait") {
		t.Error(`"\\ait" should be candidate`)
	}
	if metaPrefixCandidate("\\ait ") {
		t.Error(`exact match should not be reported as candidate`)
	}
	if metaPrefixCandidate("\\zzz") {
		t.Error(`"\\zzz" should not be candidate`)
	}
	if _, ok := metaPrefixMatch("!! "); !ok {
		t.Error(`"!! " should match`)
	}
	if _, ok := metaPrefixMatch("\\ait"); ok {
		t.Error(`"\\ait" (no trailer) should not match`)
	}
}
