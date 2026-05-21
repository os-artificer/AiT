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

package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestExecuteWritersSilencesStderr(t *testing.T) {
	capture := &bytes.Buffer{}
	err := ExecuteWriters([]string{"not-a-real-subcommand"}, os.Stdout, capture)
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

func TestExecuteWritersErrorFormattedForREPL(t *testing.T) {
	err := ExecuteWriters([]string{"not-a-real-subcommand"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	var buf bytes.Buffer
	shellPrintErrForTest(&buf, err.Error())
	out := buf.String()
	if strings.Count(out, "AiT:") != 1 {
		t.Errorf("want single AiT prefix: %q", out)
	}
}

// shellPrintErrForTest mirrors shell REPL error formatting (package shell keeps helpers unexported).
func shellPrintErrForTest(w io.Writer, msg string) {
	msg = strings.TrimSpace(strings.ReplaceAll(msg, "\n", " "))
	if msg == "" {
		return
	}
	_, _ = io.WriteString(w, "AiT: "+msg+"\r\n")
}
