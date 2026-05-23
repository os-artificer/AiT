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

package login

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpContainsFlags(t *testing.T) {
	if !helpContains(
		"-p port",
		"-J destination",
		"-o key=value",
		"--save",
		"--alias name",
		"Not supported",
		"-L, -R, -D",
		"Examples:",
		"OpenSSH-compatible flags:",
		"AiT flags:",
	) {
		t.Fatal("login help missing expected sections")
	}
}

func TestWriteHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ait login [flags]") {
		t.Errorf("missing usage line: %q", out[:min(200, len(out))])
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
