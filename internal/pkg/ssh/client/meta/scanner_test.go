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

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// drive feeds input one byte at a time and aggregates the scanner outputs.
func drive(input string, echoPrefix string, localActive bool) (forwarded string, echoed string, metaLines []string) {
	s := New(echoPrefix)
	s.SyncFrom(localActive, echoPrefix)
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

func driveRemote(input string) (forwarded string, echoed string, metaLines []string) {
	return drive(input, "", false)
}

func TestMetaScanner_BashHistoryWithArgsPassthrough(t *testing.T) {
	fwd, _, meta := driveRemote("!!ls\n")
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta lines: %v", meta)
	}
}

// OA21: !!ls always forwards to the remote PTY (bash history); AI no longer owns !!.
func TestMetaScanner_BangBangForwardsPTY(t *testing.T) {
	fwd, _, meta := driveRemote("!!ls\n")
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta lines: %v", meta)
	}
}

func TestMetaScanner_AitPrefixCapturedWhenAIDisabled(t *testing.T) {
	fwd, _, meta := driveRemote("!ait local\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait local" {
		t.Errorf("meta=%v want [!ait local]", meta)
	}
}

// TestScanner_MaxLookCoversLongestPrefix ensures holding buf fits the longest opener.
func TestScanner_MaxLookCoversLongestPrefix(t *testing.T) {
	longest := "!ait remote\n"
	if maxLook < len(longest) {
		t.Fatalf("maxLook=%d shorter than len(%q)", maxLook, longest)
	}
	fwd, _, meta := driveRemote("!ait local\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait local" {
		t.Errorf("meta=%v want [!ait local]", meta)
	}
}

func TestMetaScanner_BashHistoryColonModifierPassthrough(t *testing.T) {
	fwd, _, meta := driveRemote("!!:gs/x/y/\n")
	if fwd != "!!:gs/x/y/\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BangBangAloneForwards(t *testing.T) {
	fwd, _, meta := driveRemote("!!\n")
	if fwd != "!!\n" {
		t.Errorf("forward=%q want %q", fwd, "!!\n")
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BangBangSpaceQueryForwards(t *testing.T) {
	fwd, _, meta := driveRemote("!! how to grep\n")
	if fwd != "!! how to grep\n" {
		t.Errorf("forward=%q want passthrough", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BangBangUTF8QueryForwards(t *testing.T) {
	line := "!! 当前系统是什么名字?\n"
	fwd, _, meta := driveRemote(line)
	if fwd != line {
		t.Errorf("forward=%q want %q", fwd, line)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_AitAIQuery(t *testing.T) {
	fwd, _, meta := driveRemote("!ait ai how to grep\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai how to grep" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitAIAsk(t *testing.T) {
	fwd, _, meta := driveRemote("!ait ai ask ping\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai ask ping" {
		t.Errorf("meta=%v", meta)
	}
}

// RA6: bare !ait ai line is captured as meta and yields empty initial prompt.
func TestMetaScanner_AitAIBareLine(t *testing.T) {
	fwd, _, meta := driveRemote("!ait ai\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai" {
		t.Errorf("meta=%v", meta)
	}
	_, ok := AIPromptFromLine(meta[0])
	if ok {
		t.Error("AIPromptFromLine must always return false")
	}
}

func TestMetaScanner_BangBangAskForwards(t *testing.T) {
	fwd, _, meta := driveRemote("!! ask ping\n")
	if fwd != "!! ask ping\n" {
		t.Errorf("forward=%q want passthrough", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_AitAIAgent(t *testing.T) {
	fwd, _, meta := driveRemote("!ait ai agent find large files\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait ai agent find large files" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_BangBangAgentForwards(t *testing.T) {
	fwd, _, meta := driveRemote("!! agent find large files\n")
	if fwd != "!! agent find large files\n" {
		t.Errorf("forward=%q want passthrough", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_AitLocal(t *testing.T) {
	fwd, _, meta := driveRemote("!ait local\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait local" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitRemote(t *testing.T) {
	fwd, _, meta := driveRemote("!ait remote\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait remote" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_AitSuspend(t *testing.T) {
	fwd, _, meta := driveRemote("!ait suspend\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait suspend" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_HelpLine(t *testing.T) {
	fwd, _, meta := driveRemote("!ait help\n")
	if len(fwd) != 0 {
		t.Errorf("forwarded %q", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait help" {
		t.Errorf("meta=%v want [!ait help]", meta)
	}
}

func TestMetaScanner_LegacyBackslashAitPassthrough(t *testing.T) {
	fwd, _, meta := driveRemote("\\ait suspend\n")
	if fwd != "\\ait suspend\n" {
		t.Errorf("forward=%q want passthrough", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_BackslashArtifactPassthrough(t *testing.T) {
	fwd, _, meta := driveRemote("\\artifact\n")
	if fwd != "\\artifact\n" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_NotAtLineStart(t *testing.T) {
	fwd, _, meta := driveRemote("x!!y\n")
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
	fwd, _, meta := driveRemote(sb.String())
	if len(meta) != 0 {
		t.Errorf("expected no meta dispatch for runaway line, got %v", meta)
	}
	if !strings.HasPrefix(fwd, "!! ") || !strings.HasSuffix(fwd, "\n") {
		t.Errorf("runaway line should be flushed verbatim, got len=%d", len(fwd))
	}
}

func TestMetaScanner_TwoMetasInARow(t *testing.T) {
	fwd, _, meta := driveRemote("!ait local\n!ait remote\n")
	if fwd != "" {
		t.Errorf("forward=%q", fwd)
	}
	if len(meta) != 2 || meta[0] != "!ait local" || meta[1] != "!ait remote" {
		t.Errorf("meta=%v", meta)
	}
}

func TestMetaScanner_MetaLineLocalEcho(t *testing.T) {
	_, echo, meta := driveRemote("!ait local\n")
	if len(meta) != 1 {
		t.Fatalf("meta=%v", meta)
	}
	if !strings.Contains(echo, "!ait") || !strings.Contains(echo, "local") {
		t.Errorf("echo=%q should contain meta line", echo)
	}
}

// TestMetaScanner_RemoteInvalidBangBangNoHoldingEcho (MD5) rejects !!ait\n passthrough
// without duplicating remote shell echo on the newline flush. Typo-prefix redraw
// during holding may appear while typing !!ait before the invalid newline.
func TestMetaScanner_RemoteInvalidBangBangNoHoldingEcho(t *testing.T) {
	fwd, _, meta := driveRemote("!!ait\n")
	if fwd != "!!ait\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ait\n")
	}
	if len(meta) != 0 {
		t.Errorf("meta=%v want none", meta)
	}
}

// TestMetaScanner_BangBangAitTypoHelp treats !!ait as !ait meta (doubled-bang typo).
func TestMetaScanner_BangBangAitTypoHelp(t *testing.T) {
	fwd, echo, meta := driveRemote("!!ait help\n")
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait help" {
		t.Errorf("meta=%v want [!ait help]", meta)
	}
	if !strings.Contains(echo, "help") {
		t.Errorf("echo=%q should show meta line editing", echo)
	}
}

func TestMetaScanner_LocalBangBangAitTypoHelp(t *testing.T) {
	fwd, _, meta := drive("!!ait help\n", "AiT Local> ", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait help" {
		t.Errorf("meta=%v want [!ait help]", meta)
	}
}

// TestMetaScanner_RemoteHoldingEchoTimeline checks incremental redraw during remote holding.
func TestMetaScanner_RemoteHoldingEchoTimeline(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		checkIdx   int
		wantInEcho string
	}{
		{
			name:       "aitPrefix",
			input:      "!ait ",
			checkIdx:   len("!ait") - 1,
			wantInEcho: "!ait",
		},
		{
			name:       "bangAitTypo",
			input:      "!!ait ",
			checkIdx:   len("!!ait") - 1,
			wantInEcho: "!!ait",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := New("AiT u@h> ")
			var echoAtCheck string
			for i, b := range []byte(tc.input) {
				_, echo, _, _ := s.Feed(b)
				if i == tc.checkIdx {
					echoAtCheck = string(echo)
				}
			}
			if echoAtCheck == "" || !strings.Contains(echoAtCheck, tc.wantInEcho) {
				t.Errorf("echo at idx %d=%q want containing %q", tc.checkIdx, echoAtCheck, tc.wantInEcho)
			}
		})
	}
}

// TestMetaScanner_RemoteCtrlUClearsHoldingRedraw ensures Ctrl+U after !! does not leave
// holdingRedraw set so a later !!ls passthrough does not spuriously clear the line.
func TestMetaScanner_RemoteCtrlUClearsHoldingRedraw(t *testing.T) {
	fwd, echo, meta := driveRemote("!!\x15!!ls\n")
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if strings.Contains(echo, "ls") {
		t.Errorf("echo=%q must not locally echo passthrough command", echo)
	}
	if len(meta) != 0 {
		t.Errorf("meta=%v want none", meta)
	}
}

func TestBangAitTypoPrefix_SyncWithMetaPrefix(t *testing.T) {
	want := "!!" + toolname.MetaPrefix[1:]
	if bangAitTypoPrefix != want {
		t.Fatalf("bangAitTypoPrefix=%q want %q", bangAitTypoPrefix, want)
	}
	if !isBangAitTypoPrefix("!!ait") {
		t.Error("!!ait should be typo prefix")
	}
	if isBangAitTypoPrefix("!! how") {
		t.Error("!! how should not be typo prefix")
	}
}

func TestMetaScanner_RemoteInvalidBangBangPrefixNoDoubleEcho(t *testing.T) {
	fwd, echo, meta := driveRemote("!!ls\n")
	if fwd != "!!ls\n" {
		t.Errorf("forward=%q want %q", fwd, "!!ls\n")
	}
	if strings.Contains(echo, "ls") {
		t.Errorf("echo=%q must not locally echo passthrough command", echo)
	}
	if len(meta) != 0 {
		t.Errorf("meta=%v want none", meta)
	}
}

func TestMetaScanner_CtrlUThenBangBang(t *testing.T) {
	fwd, _, meta := driveRemote("\x15!! how\n")
	if fwd != "\x15!! how\n" {
		t.Errorf("forward=%q want ctrl-u then !! how", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}

func TestMetaScanner_PrefixCandidateHelpers(t *testing.T) {
	if !prefixCandidate("!!") {
		t.Error(`"!!" should be candidate (bash history / !!ait typo hold)`)
	}
	if !prefixCandidate("!ait") {
		t.Error(`"!ait" should be candidate`)
	}
	if prefixCandidate("!ait ai") {
		t.Error(`"!ait ai" is past "!ait " opener; not a holding candidate`)
	}
	if prefixCandidate("!zzz") {
		t.Error(`"!zzz" should not be candidate`)
	}
	if _, ok := prefixMatch("!! "); ok {
		t.Error(`"!! " must not match after AI strip`)
	}
	if _, ok := prefixMatch("!ait remote\n"); !ok {
		t.Error(`"!ait remote\n" should match`)
	}
}

func TestMetaScanner_LegacyPrefixNotCaptured(t *testing.T) {
	cases := []string{"!aitssh help\n", "!aitssh ai ask hi\n"}
	for _, input := range cases {
		fwd, _, meta := driveRemote(input)
		if len(meta) != 0 {
			t.Errorf("input %q: meta=%v want none (legacy prefix not recognized)", input, meta)
		}
		trimmed := strings.TrimSuffix(input, "\n")
		if !strings.Contains(fwd, trimmed) {
			t.Errorf("input %q: forward=%q want to contain %q", input, fwd, trimmed)
		}
	}
}

func TestAIPromptFromLine(t *testing.T) {
	for _, line := range []string{
		"!! how to grep",
		"!!",
		"!ait ai how to grep",
		"!ait ai ask hello",
		"!ait suspend",
	} {
		p, ok := AIPromptFromLine(line)
		if ok || p != "" {
			t.Errorf("AIPromptFromLine(%q) = (%q, %v), want (\"\", false)", line, p, ok)
		}
	}
}

const (
	keyLeftArrow = "\x1b[D"
)

func TestMetaScanner_ME1_CtrlKAfterLeft(t *testing.T) {
	input := "!ait hello" + keyLeftArrow + keyLeftArrow + keyLeftArrow + keyLeftArrow + string([]byte{ctrlK}) + "\n"
	_, _, meta := driveRemote(input)
	if len(meta) != 1 || meta[0] != "!ait h" {
		t.Errorf("meta=%v want [!ait h]", meta)
	}
}

func TestMetaScanner_ME2_CtrlW(t *testing.T) {
	input := "!ait hello" + string([]byte{ctrlW}) + "\n"
	_, _, meta := driveRemote(input)
	if len(meta) != 1 || meta[0] != "!ait " {
		t.Errorf("meta=%v want [!ait ]", meta)
	}
}

func TestMetaScanner_ME3_CtrlUClearThenNewLine(t *testing.T) {
	s := New("")
	var fwd strings.Builder
	var meta []string
	seq := "!ait hello" + string([]byte{ctrlU}) + "!ait bar\n"
	for i := 0; i < len(seq); i++ {
		out, _, line, ok := s.Feed(seq[i])
		fwd.Write(out)
		if ok {
			meta = append(meta, line)
		}
	}
	if fwd.Len() != 0 {
		t.Errorf("forward=%q want empty (meta Ctrl+U must not forward)", fwd.String())
	}
	if len(meta) != 1 || meta[0] != "!ait bar" {
		t.Errorf("meta=%v want [!ait bar]", meta)
	}
}

func TestMetaScanner_ME4_CtrlL(t *testing.T) {
	input := "!ait x" + string([]byte{ctrlL})
	_, echo, _ := driveRemote(input)
	if !strings.Contains(echo, "\x1b[2J") {
		t.Errorf("echo=%q want clear-screen sequence", echo)
	}
}

func TestMetaScanner_ME5_BackspaceMidLine(t *testing.T) {
	input := "!ait ab" + keyLeftArrow + "\x7f\n"
	_, _, meta := driveRemote(input)
	if len(meta) != 1 || meta[0] != "!ait b" {
		t.Errorf("meta=%v want [!ait b]", meta)
	}
}

func TestMetaScanner_ME6_CtrlAInsert(t *testing.T) {
	input := "!ait ab" + string([]byte{ctrlA}) + "z\n"
	_, _, meta := driveRemote(input)
	if len(meta) != 1 || meta[0] != "z!ait ab" {
		t.Errorf("meta=%v want [z!ait ab]", meta)
	}
}

func TestMetaScanner_ME7_UTF8Backspace(t *testing.T) {
	input := "!ait \xe4\xb8\xad" + "\x7f\n"
	_, _, meta := driveRemote(input)
	if len(meta) != 1 || meta[0] != "!ait " {
		t.Errorf("meta=%v want [!ait ]", meta)
	}
}

func TestMetaScanner_MD2_LocalBangPassthrough(t *testing.T) {
	fwd, _, meta := drive("!! how\n", "AiT Local> ", true)
	if fwd != "!! how\n" {
		t.Errorf("forward=%q want !! how\\n", fwd)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta lines: %v", meta)
	}
}

func TestMetaScanner_MD2b_LocalAitStillCapture(t *testing.T) {
	fwd, _, meta := drive("!ait remote\n", "AiT Local> ", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait remote" {
		t.Errorf("meta=%v want [!ait remote]", meta)
	}
}

// TestMetaScanner_MD4_LocalBangNoOrphanEcho defers local !! passthrough until a 3rd byte
// so !!ait meta typo can be disambiguated; bare !! must not orphan-echo the first !.
func TestMetaScanner_MD4_LocalBangNoOrphanEcho(t *testing.T) {
	fwd, echo, meta := drive("!!", "AiT Local> ", true)
	if fwd != "" {
		t.Errorf("forward=%q want empty (deferred until 3rd byte)", fwd)
	}
	if echo != "" {
		t.Errorf("echo=%q want empty (no orphan !)", echo)
	}
	if len(meta) != 0 {
		t.Errorf("unexpected meta: %v", meta)
	}
}
