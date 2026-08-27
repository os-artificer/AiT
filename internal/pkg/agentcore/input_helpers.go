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

package agentcore

import (
	"bufio"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// scannerLineReader returns a line reader over in that writes prompt to out
// first. One Scanner is retained by the closure so input buffered by an earlier
// call is not lost. Returns io.EOF once in is exhausted.
// For cooked-mode input only; a raw terminal must use ttyLineReader, whose
// Enter key arrives as a carriage return the Scanner would never split on.
func scannerLineReader(in io.Reader, out io.Writer) func(prompt string) (string, error) {
	if in == nil {
		return func(string) (string, error) { return "", io.EOF }
	}

	sc := bufio.NewScanner(in)
	return func(prompt string) (string, error) {
		if prompt != "" && out != nil {
			_, _ = io.WriteString(out, prompt)
		}
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return sc.Text(), nil
	}
}

// ttyLineReader returns a line reader driven by the raw-mode line editor, so
// carriage-return line endings and editing keys work while the loop holds the
// terminal in raw mode. Lines are not recorded in the REPL history.
func ttyLineReader(br *bufio.Reader, out io.Writer) func(prompt string) (string, error) {
	return func(prompt string) (string, error) {
		// The line editor only redraws the prompt after an edit, so paint it here.
		if prompt != "" && out != nil {
			agentWrite(out, prompt)
		}
		return terminal.ReadTTYLineNoHistory(br, out, prompt)
	}
}
