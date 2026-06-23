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

package mode

import (
	"bytes"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

func TestWriteREPLHelp_Local(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteREPLHelp(&buf, ai.ScopeLocal); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		pkghelpdoc.SectionCommands,
		pkghelpdoc.SectionKeys,
		"ai-ssh list",
		"Ctrl+R",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("local REPL help missing %q:\n%s", want, out)
		}
	}
}

func TestWriteREPLHelp_Remote(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteREPLHelp(&buf, ai.ScopeRemote); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		pkghelpdoc.SectionCommands,
		pkghelpdoc.SectionKeys,
		"Ctrl+R",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("remote REPL help missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "ai-ssh list") {
		t.Fatalf("remote REPL help must not include local-only ai-ssh list extra:\n%s", out)
	}
}
