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

package shell

import (
	"bytes"
	"strings"
	"testing"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

func TestWriteShellHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteShellHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:",
		pkghelpdoc.SectionREPLCommands,
		pkghelpdoc.SectionKeys,
		"Ctrl+R",
		"Examples:",
		"help for shell",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("shell help missing %q:\n%s", want, out)
		}
	}
}

func TestWriteShellREPLHelpSections(t *testing.T) {
	t.Run("noActiveSession", func(t *testing.T) {
		var buf bytes.Buffer
		if err := WriteShellREPLHelp(&buf, false); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		for _, want := range []string{
			pkghelpdoc.SectionCommands,
			"[flags] destination",
			pkghelpdoc.SectionKeys,
			"Ctrl+R",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("shell REPL help missing %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, shellSectionActiveSession) {
			t.Fatalf("unexpected active session section:\n%s", out)
		}
	})
	t.Run("activeSession", func(t *testing.T) {
		var buf bytes.Buffer
		if err := WriteShellREPLHelp(&buf, true); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if strings.Contains(out, "[flags] destination") {
			t.Fatalf("active session REPL help must omit connect command:\n%s", out)
		}
		if !strings.Contains(out, shellREPLHelpUsageNoLoad) {
			t.Fatalf("active session REPL help missing no-login usage:\n%s", out)
		}
		for _, want := range []string{
			shellSectionActiveSession,
			pkghelpdoc.SectionCommands,
			pkghelpdoc.SectionKeys,
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("shell REPL help missing %q:\n%s", want, out)
			}
		}
	})
}
