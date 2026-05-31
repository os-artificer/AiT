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

package lrzsz

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// ConfirmFunc reads a line after prompt and returns whether the user agreed.
type ConfirmFunc func(prompt string) (bool, error)

// InstallOptions configures lrzsz auto-install confirmation.
type InstallOptions struct {
	AssumeYes  bool
	NoInstall  bool
	ActiveHost string
	Out        io.Writer
	Err        io.Writer
	In         io.Reader
}

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

// ResolveInstallAllowed applies AssumeYes, NoInstall, and optional interactive confirm.
func ResolveInstallAllowed(missing bool, opts InstallOptions, confirm ConfirmFunc) (bool, error) {
	if !missing {
		return false, nil
	}
	if opts.AssumeYes {
		return true, nil
	}
	if opts.NoInstall {
		return false, nil
	}
	in := opts.In
	if in == nil {
		in = os.Stdin
	}
	if !isTerminalIn(in) {
		return false, ErrNonInteractiveInstall
	}
	if confirm == nil {
		out := opts.Out
		if out == nil {
			out = os.Stderr
		}
		confirm = DefaultConfirmFunc(in, out)
	}
	return confirm(InstallPrompt(opts.ActiveHost))
}

// InstallPrompt returns the interactive install confirmation line.
func InstallPrompt(activeHost string) string {
	host := strings.TrimSpace(activeHost)
	if host == "" {
		host = "active shell host"
	}
	return fmt.Sprintf(
		"Install lrzsz on %s? This may run sudo on the remote shell. [y/N] ",
		host,
	)
}

func isTerminalIn(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok || f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// WrapInstallDeclined returns an errhint for declined install.
func WrapInstallDeclined(activeHost string) error {
	host := strings.TrimSpace(activeHost)
	if host == "" {
		host = "active shell host"
	}
	return errhint.WithTry(
		ErrInstallDeclined,
		fmt.Sprintf("Chain-mode file transfer requires lrzsz on %s.", host),
		ManualTryLines(TargetUnknown)...,
	)
}

// WrapInstallFailed returns an errhint when auto-install did not succeed.
func WrapInstallFailed(cause error, target Target) error {
	tries := ManualTryLines(target)
	if len(tries) == 0 {
		tries = ManualTryLines(TargetUnknown)
	}
	return errhint.WithTry(
		fmt.Errorf("%w: %v", ErrInstallFailed, cause),
		"Automatic lrzsz install did not succeed (sudo may be required).",
		tries...,
	)
}

// WrapNonInteractive returns an errhint when install needs TTY confirm.
func WrapNonInteractive(target Target) error {
	return errhint.WithTry(
		ErrNonInteractiveInstall,
		"Chain-mode file transfer requires lrzsz; auto-install needs an interactive terminal.",
		ManualTryLines(target)...,
	)
}

// WrapMissingChain returns an errhint when lrzsz is missing in chain mode before install prompt.
func WrapMissingChain(activeHost string, target Target) error {
	host := strings.TrimSpace(activeHost)
	if host == "" {
		host = "active shell host"
	}
	tries := ManualTryLines(target)
	if len(tries) == 0 {
		tries = ManualTryLines(TargetUnknown)
	}
	return errhint.WithTry(
		ErrMissing,
		fmt.Sprintf("Chain-mode file transfer uses ZMODEM (lrzsz) over the PTY on %s.", host),
		tries...,
	)
}
