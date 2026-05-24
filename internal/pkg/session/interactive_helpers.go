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

package session

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// ErrNonInteractivePick is returned when session pick requires a TTY on stdin.
var ErrNonInteractivePick = errhint.NewSentinel("non-interactive terminal")

// CheckInteractiveStdin returns ErrNonInteractivePick when stdin is not a terminal.
// Used before interactive session pick for ait session --load or --switch without a target.
func CheckInteractiveStdin() error {
	if stdinIsTerminal() {
		return nil
	}
	return errhint.WithTry(ErrNonInteractivePick,
		"Interactive session pick requires a TTY.",
		tryLoadByAlias,
		tryLoadByFullID,
		tryLoadInteractive,
		trySessionHelp,
	)
}

func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// readPickLine reads one selection line; uses the line editor when stdin is a TTY.
func readPickLine(in io.Reader, out io.Writer) (string, error) {
	if stdinIsTerminal() && in == os.Stdin {
		br := bufio.NewReader(in)
		return terminal.ReadLine(br, out, terminal.LineEditOptions{Prompt: ""})
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
