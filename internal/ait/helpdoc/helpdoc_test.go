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

package helpdoc

import (
	"bytes"
	"strings"
	"testing"
)

func helpMustContain(t *testing.T, out string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(out, s) {
			t.Fatalf("help missing %q in:\n%s", s, out)
		}
	}
}

func TestLoginHelpMatchesLegacyGolden(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLoginHelp(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != LoginHelpFull {
		t.Fatalf("login help drifted from golden (%d vs %d bytes)", len(buf.String()), len(LoginHelpFull))
	}
}

func TestOpenSSHBlockInLoginGolden(t *testing.T) {
	if !strings.Contains(LoginHelpFull, SectionOpenSSHFlags) {
		t.Fatal("LoginHelpFull must embed SectionOpenSSHFlags verbatim")
	}
}

func TestHelpRequested(t *testing.T) {
	if !HelpRequested([]string{"-h"}) || !HelpRequested([]string{"--help"}) {
		t.Error("expected help requested")
	}
	if HelpRequested([]string{"host"}) || HelpRequested([]string{"-h", "host"}) {
		t.Error("unexpected help requested")
	}
}

func TestWriteLoginHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLoginHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(),
		"Usage:",
		"Examples:",
		"OpenSSH-compatible flags:",
		"help for login",
		"-p port",
		"Not supported",
	)
}

func TestWriteTransferHelpSections(t *testing.T) {
	for _, cmd := range []string{"get", "put"} {
		var buf bytes.Buffer
		if err := WriteTransferHelp(&buf, cmd); err != nil {
			t.Fatal(err)
		}
		helpMustContain(t, buf.String(),
			"Usage:",
			"Examples:",
			"OpenSSH-compatible flags:",
			"Not supported",
			"help for "+cmd,
			SectionOpenSSHFlags[:20],
		)
	}
}

func TestWriteListHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteListHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "help for list")
}

func TestWriteRootHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Commands:", "help for ait", "login", "ai")
}

func TestWriteLoadHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLoadHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "Arguments:", "AiT flags:", "help for load", "--tmux")
}

func TestWriteShellHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteShellHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "REPL commands", "help for shell")
}

func TestWriteShellREPLHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteShellREPLHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "Commands:", "help for help", "-p port")
}

func TestWriteSuspendHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteSuspendHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "help for suspend", "!ait suspend")
}

func TestWriteConfigHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteConfigHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "Subcommands:", "help for config")
}

func TestWriteConfigSetGetHelpSections(t *testing.T) {
	var setBuf, getBuf bytes.Buffer
	if err := WriteConfigSetHelp(&setBuf); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfigGetHelp(&getBuf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, setBuf.String(), "Usage:", "Examples:", "help for set", "session.auth_ttl", "ai.default")
	helpMustContain(t, getBuf.String(), "Usage:", "Examples:", "help for get", "ai.deepseek.api_key", "minimax.api_key")
}

func TestWriteSessionHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteSessionHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "Subcommands:", "help for session")
}

func TestWriteSessionSwitchDeleteHelpSections(t *testing.T) {
	var swBuf, delBuf bytes.Buffer
	if err := WriteSessionSwitchHelp(&swBuf); err != nil {
		t.Fatal(err)
	}
	if err := WriteSessionDeleteHelp(&delBuf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, swBuf.String(), "Usage:", "Examples:", "help for session switch", "--force")
	helpMustContain(t, delBuf.String(), "Usage:", "Examples:", "help for session delete", "--force")
}

func TestWriteMetaHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteMetaHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "Commands:", "help for help", "!ait list", "ait shell")
}
