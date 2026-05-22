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

package ollama

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const confirmYes = "y"

// ConfirmFunc reads a line after prompt and returns whether the user agreed.
type ConfirmFunc func(prompt string) (bool, error)

// DefaultConfirmFunc prompts on out and reads from in (TTY-style [y/N]).
func DefaultConfirmFunc(in io.Reader, out io.Writer) ConfirmFunc {
	return func(prompt string) (bool, error) {
		_, _ = fmt.Fprint(out, prompt)
		line, err := readConfirmLine(in)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(strings.TrimSpace(line), confirmYes), nil
	}
}

func readConfirmLine(r io.Reader) (string, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return line, nil
}

// ResolveInstallAllowed applies --yes, --no-install, and optional interactive confirm.
func ResolveInstallAllowed(binaryMissing bool, opts Options, confirm ConfirmFunc) (bool, error) {
	if !binaryMissing {
		return false, nil
	}
	if opts.AssumeYes {
		return true, nil
	}
	if opts.NoInstall {
		return false, nil
	}
	if !isTerminalIn(os.Stdin) {
		_, _ = fmt.Fprintln(opts.Err, "AiT: non-interactive terminal; use --yes to allow Ollama install or --no-install to skip")
		return false, nil
	}
	if confirm == nil {
		confirm = DefaultConfirmFunc(os.Stdin, opts.Out)
	}
	return confirm(installPromptSummary())
}

func isTerminalIn(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
