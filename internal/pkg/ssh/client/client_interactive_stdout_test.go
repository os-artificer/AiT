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
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestHandleRemoteOutput_InteractiveStdoutCRLF (TM11a) translates bare LF to CRLF on interactive stdout.
func TestHandleRemoteOutput_InteractiveStdoutCRLF(t *testing.T) {
	user, host := "u", "h"
	c := testClientWithPrompt(user, host)

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.pauseRemoteOut.Store(false)

	c.handleRemoteOutput([]byte("AiT u@h:/tmp> \n"))
	got := buf.String()
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("want CRLF in stdout, got %q", got)
	}
	if strings.Contains(got, "\r\r\n") {
		t.Fatalf("double CR: %q", got)
	}
}

// TestHandleRemoteOutput_GraceStripUsesCRLF (TM11a) grace-path stdout still goes through CRLF writer.
func TestHandleRemoteOutput_GraceStripUsesCRLF(t *testing.T) {
	user, host := "u", "h"
	c := testClientWithPrompt(user, host)

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.pauseRemoteOut.Store(false)
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	leak := "stty -echo 2>/dev/null\r\n"
	prompt := "AiT u@h:/tmp> \n"
	c.handleRemoteOutput([]byte(leak + prompt))

	got := buf.String()
	if strings.Contains(got, "stty -echo") {
		t.Fatalf("leak should be stripped: %q", got)
	}
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("want CRLF after strip, got %q", got)
	}
}
