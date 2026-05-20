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
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRunCLISilencesStderr(t *testing.T) {
	capture := &bytes.Buffer{}
	oldErr := shellCLIErr
	shellCLIErr = capture
	defer func() { shellCLIErr = oldErr }()

	err := runCLI([]string{"not-a-real-subcommand"})
	if err == nil {
		t.Fatal("expected error")
	}
	if capture.Len() != 0 {
		t.Errorf("cobra should not write stderr, got %q", capture.String())
	}
	if strings.Contains(err.Error(), "\n") {
		t.Errorf("error should be one line from caller's view: %q", err.Error())
	}
}

func TestRunCLIErrorFormattedForREPL(t *testing.T) {
	err := runCLI([]string{"not-a-real-subcommand"})
	if err == nil {
		t.Fatal("expected error")
	}
	var buf bytes.Buffer
	shellPrintErr(&buf, err.Error())
	out := buf.String()
	if strings.Count(out, "AiT:") != 1 {
		t.Errorf("want single AiT prefix: %q", out)
	}
}

func TestExecuteShellLineStripsTrailingSemicolon(t *testing.T) {
	args, err := parseShellLine(normalizeShellLine("list;"))
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "list" {
		t.Errorf("args = %v want [list]", args)
	}
}

// Ensure shellCLIErr default does not break non-shell Run path.
func TestShellCLIErrDefault(t *testing.T) {
	if shellCLIErr == nil {
		t.Fatal("shellCLIErr should be set")
	}
	_ = io.Discard
}
