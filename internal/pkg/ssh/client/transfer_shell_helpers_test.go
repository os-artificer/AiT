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

func TestBuildSzCmd_Heredoc(t *testing.T) {
	cmd := buildSzCmd("/remote/file.txt")
	if !strings.Contains(cmd, "sz -y") {
		t.Fatalf("missing sz: %q", cmd)
	}
	if !strings.Contains(cmd, shellQuoteArg("/remote/file.txt")) {
		t.Fatalf("missing path: %q", cmd)
	}
	if strings.Contains(cmd, "base64") {
		t.Fatalf("should not use base64: %q", cmd)
	}
}

func TestBuildRzCmd_NoRemoteRedirect(t *testing.T) {
	cmd := buildRzCmd("./subdir/file.txt")
	if !strings.Contains(cmd, "rz -y") {
		t.Fatalf("missing rz: %q", cmd)
	}
	if !strings.Contains(cmd, "cd ") {
		t.Fatalf("expected cd to parent dir: %q", cmd)
	}
	if strings.Contains(cmd, ">>") {
		t.Fatalf("must not shell-redirect to target: %q", cmd)
	}
}

func TestShellQuoteArg(t *testing.T) {
	got := shellQuoteArg("it's")
	want := "'it'\"'\"'s'"
	if got != want {
		t.Fatalf("shellQuoteArg = %q, want %q", got, want)
	}
}

func TestRemoteBaseName(t *testing.T) {
	if got := remoteBaseName("./foo/bar.txt"); got != "bar.txt" {
		t.Fatalf("remoteBaseName = %q", got)
	}
}
