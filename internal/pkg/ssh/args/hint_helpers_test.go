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

package args

import (
	"strings"
	"testing"
)

func TestDestinationRequiredErr_Format(t *testing.T) {
	err := destinationRequiredErr()
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, tryLoginHelp) {
		t.Fatalf("missing login help try line: %q", s)
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
	if !strings.Contains(s, tryGetHelp) {
		t.Fatalf("missing get help try line: %q", s)
	}
}
