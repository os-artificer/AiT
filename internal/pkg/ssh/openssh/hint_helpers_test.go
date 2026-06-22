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

package openssh

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
)

func TestDestinationRequiredErr_Format(t *testing.T) {
	err := destinationRequiredErr()
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, clihints.TryConnectHelp) {
		t.Fatalf("missing connect help try line: %q", s)
	}
}

func TestTmuxWithRemoteCommandErr_Format(t *testing.T) {
	err := TmuxWithRemoteCommandErr()
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, "cannot use --tmux with a remote command") {
		t.Fatalf("missing summary: %q", s)
	}
}

func TestTransferPathCountErr_Format(t *testing.T) {
	err := transferPathCountErr("get")
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, clihints.TryGetHelp) {
		t.Fatalf("missing get help try line: %q", s)
	}

	errPut := transferPathCountErr("put")
	sPut := errPut.Error()
	if !strings.Contains(sPut, "1 to 3 path arguments for ai-ssh put") {
		t.Fatalf("missing put path count hint: %q", sPut)
	}
	if !strings.Contains(sPut, clihints.TryPutHelp) {
		t.Fatalf("missing put help try line: %q", sPut)
	}
}
