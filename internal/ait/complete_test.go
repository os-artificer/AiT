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

package ait

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
)

func testNamesFourPlugins() []string {
	return []string{
		aitplugin.NameHost,
		aitplugin.NameImg,
		aitplugin.NameLog,
		aitplugin.NameSSH,
		aitplugin.CmdClear,
		aitplugin.CmdExit,
		aitplugin.CmdHelp,
		aitplugin.CmdHistory,
		aitplugin.CmdQuit,
		aitplugin.CmdVersion,
	}
}

func TestCompleteFirstToken_UniquePlugin(t *testing.T) {
	// AT30a
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "ais", 3)
	if display != "" {
		t.Fatalf("display = %q, want empty", display)
	}
	if line != "aissh " || cur != 6 {
		t.Fatalf("got line=%q cur=%d, want aissh /6", line, cur)
	}
}

func TestCompleteFirstToken_AIPrefixListsFour(t *testing.T) {
	// AT30b
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "ai", 2)
	if line != "ai" || cur != 2 {
		t.Fatalf("line should be unchanged, got %q cur=%d", line, cur)
	}
	want := "aihost\naiimg\nailog\naissh"
	if display != want {
		t.Fatalf("display = %q, want %q", display, want)
	}
}

func TestCompleteFirstToken_SecondTokenNoop(t *testing.T) {
	// AT30d
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "aissh ", 6)
	if line != "aissh " || cur != 6 || display != "" {
		t.Fatalf("want noop, got line=%q cur=%d display=%q", line, cur, display)
	}
}

func TestCompleteFirstToken_UniqueHelp(t *testing.T) {
	// AT30e
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "he", 2)
	if display != "" {
		t.Fatalf("display = %q", display)
	}
	if line != "help " || cur != 5 {
		t.Fatalf("got line=%q cur=%d, want help /5", line, cur)
	}
}

func TestCompleteFirstToken_HListsHelpHistory(t *testing.T) {
	// AT30f
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "h", 1)
	if line != "h" || cur != 1 {
		t.Fatalf("line unchanged want, got %q cur=%d", line, cur)
	}
	want := "help\nhistory"
	if display != want {
		t.Fatalf("display = %q, want %q", display, want)
	}
}

func TestCompleteFirstToken_LeadingSpacesPreserved(t *testing.T) {
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "  ais", 5)
	if display != "" {
		t.Fatalf("display = %q", display)
	}
	if line != "  aissh " {
		t.Fatalf("line = %q, want \"  aissh \"", line)
	}
	if cur != 8 {
		t.Fatalf("cur = %d, want 8", cur)
	}
}

func TestCompleteFirstToken_MidTokenWithArgs(t *testing.T) {
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "ais foo", 3)
	if display != "" {
		t.Fatalf("display = %q", display)
	}
	if line != "aissh foo" {
		t.Fatalf("line = %q, want aissh foo", line)
	}
	if cur != 5 {
		t.Fatalf("cur = %d, want 5", cur)
	}
}

func TestCompleteFirstToken_QuotedNoop(t *testing.T) {
	line, cur, display := completeFirstToken(testNamesFourPlugins(), `"ai`, 3)
	if line != `"ai` || cur != 3 || display != "" {
		t.Fatalf("want noop, got line=%q cur=%d display=%q", line, cur, display)
	}
}

func TestCompleteFirstToken_ZeroMatches(t *testing.T) {
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "xyz", 3)
	if line != "xyz" || cur != 3 || display != "" {
		t.Fatalf("want noop, got line=%q cur=%d display=%q", line, cur, display)
	}
}

func TestCompleteFirstToken_EmptyListsAll(t *testing.T) {
	_, _, display := completeFirstToken(testNamesFourPlugins(), "", 0)
	if display == "" {
		t.Fatal("empty line should list candidates")
	}
	if !strings.Contains(display, "aissh") || !strings.Contains(display, "help") {
		t.Fatalf("display missing expected names: %q", display)
	}
}

func TestCompleteFirstToken_InjectedThirdName(t *testing.T) {
	names := append(testNamesFourPlugins(), "aixtra")
	_, _, display := completeFirstToken(names, "ai", 2)
	if !strings.Contains(display, "aixtra") {
		t.Fatalf("expected injected name in display: %q", display)
	}
}

func TestCompleteFirstToken_MidCursorPrefix(t *testing.T) {
	// ais|sh — prefix before cursor is "ais" → unique aissh
	line, cur, display := completeFirstToken(testNamesFourPlugins(), "aissh", 3)
	if display != "" {
		t.Fatalf("display = %q", display)
	}
	if line != "aissh " || cur != 6 {
		t.Fatalf("got line=%q cur=%d", line, cur)
	}
}

func TestCompleteREPLLine_EmptyPluginsStillReserved(t *testing.T) {
	aitplugin.ResetDiscoverForTest()
	aitplugin.RestoreDiscoverHooksForTest()
	t.Cleanup(func() {
		aitplugin.ResetDiscoverForTest()
		aitplugin.RestoreDiscoverHooksForTest()
	})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(io.Discard)
	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
	aitplugin.ResetDiscoverForTest()

	line, cur, display := completeREPLLine("ai", 2)
	if line != "ai" || cur != 2 {
		t.Fatalf("line unchanged want, got %q cur=%d", line, cur)
	}
	// No plugins: prefix "ai" matches nothing among reserved (help/exit/...).
	if display != "" {
		t.Fatalf("display = %q, want empty (no reserved starts with ai)", display)
	}

	line, cur, display = completeREPLLine("he", 2)
	if line != "help " || cur != 5 || display != "" {
		t.Fatalf("reserved help: line=%q cur=%d display=%q", line, cur, display)
	}
}

func TestReservedCommandNames_Sorted(t *testing.T) {
	names := aitplugin.ReservedCommandNames()
	if len(names) != 9 {
		t.Fatalf("len = %d, want 9", len(names))
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Fatalf("not sorted: %v", names)
		}
	}
}
