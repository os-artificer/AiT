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

package lineedit

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal/history"
	"github.com/os-artificer/AiT/internal/pkg/terminal/screen"
)

func TestReadLineRawASCII(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("help\r")))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "help" {
		t.Errorf("line = %q, want help", line)
	}
}

func TestReadLineRawCRLF(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("list\r\n")))
	line, err := ReadLineRaw(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "list" {
		t.Errorf("line = %q, want list", line)
	}
}

func TestReadLineRawUTF8Chinese(t *testing.T) {
	input := append([]byte("帮助"), '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "帮助" {
		t.Errorf("line = %q, want 帮助", line)
	}
	if !strings.Contains(out.String(), "帮助") {
		t.Errorf("echo missing UTF-8, got %q", out.String())
	}
}

func TestReadLineRawBackspaceUTF8(t *testing.T) {
	var input []byte
	input = append(input, []byte("a你")...)
	input = append(input, 127)
	input = append(input, '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLineRaw(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "a" {
		t.Errorf("line = %q, want a", line)
	}
}

func TestReadLineRawIgnoresUnknownEscape(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("hi\x1b[Z\r")))
	line, err := ReadLineRaw(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "hi" {
		t.Errorf("line = %q, want hi", line)
	}
}

func TestReadLineRawCtrlCEmpty(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{3}))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "" {
		t.Errorf("line = %q, want empty", line)
	}
	if out.String() != "^C\r\n" {
		t.Errorf("output = %q, want ^C\\r\\n", out.String())
	}
}

func TestReadLineRawCtrlDEmpty(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{4}))
	_, err := ReadLineRaw(br, io.Discard, "")
	if err != io.EOF {
		t.Fatalf("err = %v, want EOF", err)
	}
}

func TestReadLineRawEmptyCRIgnored(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{'\r', 'x', '\r'}))
	line, err := ReadLineRaw(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "x" {
		t.Errorf("line = %q, want x", line)
	}
}

func TestReadLineRawCtrlLRedrawsLine(t *testing.T) {
	const prompt = "AiT> "
	input := []byte{'a', 'b', ctrlL, '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out, prompt)
	if err != nil {
		t.Fatal(err)
	}
	if line != "ab" {
		t.Errorf("line = %q, want ab", line)
	}
	got := out.String()
	if !strings.Contains(got, screen.ClearSeq()) {
		t.Errorf("output missing clear seq, got %q", got)
	}
	if !strings.Contains(got, prompt+"ab") {
		t.Errorf("output missing redrawn prompt+line, got %q", got)
	}
}

func TestReadLineRawCtrlLNoPrompt(t *testing.T) {
	input := []byte{'x', ctrlL, '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out, "")
	if err != nil {
		t.Fatal(err)
	}
	if line != "x" {
		t.Errorf("line = %q, want x", line)
	}
	if !strings.Contains(out.String(), screen.ClearSeq()) {
		t.Errorf("output missing clear seq, got %q", out.String())
	}
	if strings.Contains(out.String(), "AiT>") {
		t.Errorf("unexpected prompt redraw, got %q", out.String())
	}
}

func TestReadLine_CtrlAE(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{'h', 'i', ctrlE, ctrlA, 'x', '\r'}))
	line, err := ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "xhi" {
		t.Errorf("line = %q, want xhi", line)
	}
}

func TestReadLine_ArrowKeysUTF8(t *testing.T) {
	input := []byte("a")
	input = append(input, []byte("你")...)
	input = append(input, '\x1b', '[', 'D')
	input = append(input, 'b', '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "ab你" {
		t.Errorf("line = %q, want ab你", line)
	}
}

func TestReadLine_InsertMidLineDisplay(t *testing.T) {
	prompt := "P> "
	input := append([]byte("ab"), '\x1b', '[', 'D', 'X', '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	var out bytes.Buffer
	line, err := ReadLine(br, &out, Options{Prompt: prompt})
	if err != nil {
		t.Fatal(err)
	}
	if line != "aXb" {
		t.Errorf("line = %q, want aXb", line)
	}
	if !strings.Contains(out.String(), prompt+"aXb") {
		t.Errorf("out = %q, want redraw containing %qaXb", out.String(), prompt)
	}
}

func TestReadLine_HistoryUpDown(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("first")
	hist.Add("second")

	br := bufio.NewReader(bytes.NewReader([]byte{'\x1b', '[', 'A', '\x1b', '[', 'A', '\r'}))
	line, err := ReadLine(br, io.Discard, Options{History: hist})
	if err != nil {
		t.Fatal(err)
	}
	if line != "first" {
		t.Errorf("line = %q, want first", line)
	}
}

func TestReadLine_HistorySkipEmpty(t *testing.T) {
	hist := history.NewLineHistory()
	br := bufio.NewReader(bytes.NewReader([]byte{'\r', 'x', '\r'}))
	_, err := ReadLine(br, io.Discard, Options{History: hist, RecordHistory: true})
	if err != nil {
		t.Fatal(err)
	}
	if hist.Len() != 1 || hist.At(0) != "x" {
		t.Fatalf("history = [%q], want [x]", hist.At(0))
	}

	br = bufio.NewReader(bytes.NewReader([]byte{ctrlC}))
	_, err = ReadLine(br, io.Discard, Options{History: hist, RecordHistory: true})
	if err != nil {
		t.Fatal(err)
	}
	if hist.Len() != 1 {
		t.Fatalf("history len = %d, want 1 after Ctrl+C", hist.Len())
	}
}

func TestReadLine_KillKeys(t *testing.T) {
	input := append([]byte("hello world"), ctrlU, 'x', '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "x" {
		t.Errorf("Ctrl+U line = %q, want x", line)
	}

	input = append([]byte("hello world"), '\x1b', '[', 'H')
	for i := 0; i < 6; i++ {
		input = append(input, '\x1b', '[', 'C')
	}
	input = append(input, ctrlK, '\r')
	br = bufio.NewReader(bytes.NewReader(input))
	line, err = ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "hello " {
		t.Errorf("Ctrl+K line = %q, want hello ", line)
	}

	input = append([]byte("hello world"), ctrlW, '\r')
	br = bufio.NewReader(bytes.NewReader(input))
	line, err = ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "hello " {
		t.Errorf("Ctrl+W line = %q, want hello ", line)
	}
}

func TestLineHistory_Dedup(t *testing.T) {
	h := history.NewLineHistory()
	h.Add("list")
	h.Add("list")
	if h.Len() != 1 {
		t.Fatalf("len = %d, want 1", h.Len())
	}
}

func TestReadLine_CtrlRSearch(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("list")
	hist.Add("help")

	br := bufio.NewReader(bytes.NewReader([]byte{ctrlR, '\r'}))
	line, err := ReadLine(br, io.Discard, Options{History: hist})
	if err != nil {
		t.Fatal(err)
	}
	if line != "help" {
		t.Errorf("line = %q, want help (newest match)", line)
	}
}

func TestReadLine_CtrlRCycle(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("alpha")
	hist.Add("beta-list")
	hist.Add("gamma-list")

	input := []byte{ctrlR, ctrlR, ctrlR, '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{History: hist})
	if err != nil {
		t.Fatal(err)
	}
	if line != "alpha" {
		t.Errorf("line = %q, want alpha (older match after two Ctrl+R)", line)
	}
}

func TestReadLine_CtrlRAbort(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("list")

	input := []byte{'x', ctrlR, ctrlG, 'y', '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{History: hist})
	if err != nil {
		t.Fatal(err)
	}
	if line != "xy" {
		t.Errorf("line = %q, want xy (search aborted, prior buffer kept)", line)
	}
}

func TestReadLine_CtrlRAcceptRight(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("list")

	input := []byte{ctrlR, '\x1b', '[', 'C', '!', '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{History: hist})
	if err != nil {
		t.Fatal(err)
	}
	if line != "list!" {
		t.Errorf("line = %q, want list!", line)
	}
}

func TestShouldRecordLine(t *testing.T) {
	if shouldRecordLine("") || shouldRecordLine("   ") {
		t.Fatal("empty should not record")
	}
	if !shouldRecordLine("list") {
		t.Fatal("list should record")
	}
	if shouldRecordLine("export AIT_LOCAL_MODE=1") {
		t.Fatal("AIT env export should not record")
	}
	if shouldRecordLine("PS1=$ ") {
		t.Fatal("PS1 should not record")
	}
}

func TestReadLine_NilCompleterInsertsTab(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{'a', '\t', 'b', '\r'}))
	line, err := ReadLine(br, io.Discard, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if line != "a\tb" {
		t.Errorf("line = %q, want a\\tb", line)
	}
}

func TestReadLine_CompleterUniqueExpand(t *testing.T) {
	completer := func(line string, cursor int) (string, int, string) {
		if line != "ais" || cursor != 3 {
			t.Fatalf("completer got line=%q cursor=%d", line, cursor)
		}
		return "aissh ", 6, ""
	}
	br := bufio.NewReader(bytes.NewReader([]byte{'a', 'i', 's', '\t', '\r'}))
	line, err := ReadLine(br, io.Discard, Options{Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if line != "aissh " {
		t.Errorf("line = %q, want aissh ", line)
	}
}

func TestReadLine_CompleterLCPStretch(t *testing.T) {
	completer := func(line string, cursor int) (string, int, string) {
		return "hel", 3, ""
	}
	br := bufio.NewReader(bytes.NewReader([]byte{'h', '\t', '\r'}))
	line, err := ReadLine(br, io.Discard, Options{Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if line != "hel" {
		t.Errorf("line = %q, want hel", line)
	}
}

func TestReadLine_CompleterListDisplayCRLF(t *testing.T) {
	completer := func(line string, cursor int) (string, int, string) {
		return line, cursor, "help\nhistory"
	}
	br := bufio.NewReader(bytes.NewReader([]byte{'h', '\t', '\r'}))
	var out bytes.Buffer
	line, err := ReadLine(br, &out, Options{Prompt: "ait> ", Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if line != "h" {
		t.Errorf("line = %q, want h", line)
	}
	s := out.String()
	if !strings.Contains(s, "help\r\nhistory\r\n") {
		t.Errorf("missing CRLF list, got %q", s)
	}
	if strings.Contains(s, "\r\r\n") {
		t.Errorf("double CR in output: %q", s)
	}
}

func TestReadLine_CompleterDisplayNoTrailingLF(t *testing.T) {
	completer := func(line string, cursor int) (string, int, string) {
		return line, cursor, "help\nhistory"
	}
	br := bufio.NewReader(bytes.NewReader([]byte{'h', '\t', 'x', '\r'}))
	var out bytes.Buffer
	line, err := ReadLine(br, &out, Options{Prompt: "P> ", Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if line != "hx" {
		t.Errorf("line = %q, want hx", line)
	}
	s := out.String()
	// After list, redraw must leave candidates; typed x must appear after redrawn prompt.
	if !strings.Contains(s, "help\r\nhistory\r\n") {
		t.Errorf("list not preserved before redraw, got %q", s)
	}
}

func TestReadLine_CompleterNoopNoRedraw(t *testing.T) {
	calls := 0
	completer := func(line string, cursor int) (string, int, string) {
		calls++
		return line, cursor, ""
	}
	br := bufio.NewReader(bytes.NewReader([]byte{'a', '\t', '\r'}))
	var out bytes.Buffer
	line, err := ReadLine(br, &out, Options{Prompt: "P> ", Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if line != "a" {
		t.Errorf("line = %q, want a", line)
	}
	if calls != 1 {
		t.Fatalf("completer calls = %d, want 1", calls)
	}
	// Echo of 'a' only; no redraw CR after Tab noop.
	if strings.Count(out.String(), "\r") != 0 {
		// Enter writes \r\n; filter that out.
		trimmed := strings.TrimSuffix(out.String(), "\r\n")
		if strings.Contains(trimmed, "\r") {
			t.Errorf("unexpected redraw CR on noop: %q", out.String())
		}
	}
}

func TestReadLine_CtrlRIgnoresCompleter(t *testing.T) {
	hist := history.NewLineHistory()
	hist.Add("list")

	calls := 0
	completer := func(line string, cursor int) (string, int, string) {
		calls++
		return "BAD", 3, ""
	}
	input := []byte{ctrlR, '\t', '\r'}
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLine(br, io.Discard, Options{History: hist, Completer: completer})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("completer called %d times during reverse search", calls)
	}
	if line != "list" {
		t.Errorf("line = %q, want list", line)
	}
}
