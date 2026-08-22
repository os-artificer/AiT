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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// ErrNonInteractivePick is returned when session pick requires a TTY on stdin.
var ErrNonInteractivePick = errhint.NewSentinel(errNonInteractiveTerminal)

// ErrPickCancelled is returned when the user cancels an interactive session picker (Ctrl+C).
var ErrPickCancelled = errhint.NewSentinel(errPickCancelled)

// CheckInteractiveStdin returns ErrNonInteractivePick when stdin is not a terminal.
// Used before interactive session pick for aissh session --load or --switch without a target.
func CheckInteractiveStdin() error {
	if stdinIsTerminal() {
		return nil
	}
	return errhint.WithTry(ErrNonInteractivePick,
		hintNonInteractivePickExplain,
		tryLoadByAlias,
		tryLoadByFullID,
		tryLoadInteractive,
		trySessionHelp,
	)
}

// StdinIsTerminal reports whether os.Stdin is a TTY.
func StdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func stdinIsTerminal() bool {
	return StdinIsTerminal()
}

// PickSessionInteractiveTTY lists sessions and reads a selection with raw TTY line editing.
// Used after SSH chain actions and standalone aissh session --load/--switch pickers.
// Shell REPL callers should use PickSessionFrom directly while already in raw mode.
// Returns ErrNonInteractivePick when stdin is not a TTY, ErrPickCancelled on Ctrl+C,
// or invalid-selection / store errors from the picker.
func (m *Manager) PickSessionInteractiveTTY() (string, error) {
	if err := CheckInteractiveStdin(); err != nil {
		return "", err
	}

	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, old) }()

	out := terminal.NewCRLFWriter(os.Stdout)
	id, err := m.PickSessionFrom(os.Stdin, out)
	if err != nil {
		return "", err
	}
	terminal.DrainStdin()
	return id, nil
}

// readPickLine reads one selection line; uses the line editor when stdin is a TTY.
func readPickLine(in io.Reader, out io.Writer) (string, error) {
	if stdinIsTerminal() && in == os.Stdin {
		br := bufio.NewReader(in)
		line, err := terminal.ReadLine(br, out, terminal.LineEditOptions{Prompt: ""})
		if err != nil {
			return "", err
		}
		if line == "" {
			return "", pickCancelledErr()
		}
		return line, nil
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
