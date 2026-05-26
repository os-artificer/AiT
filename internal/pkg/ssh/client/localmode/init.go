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

package localmode

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

// Init holds a temp dir with rc files sourced at shell start (no PTY echo).
type Init struct {
	dir string
}

// Dir returns the temp directory holding the rc files.
func (i *Init) Dir() string {
	if i == nil {
		return ""
	}
	return i.dir
}

func (i *Init) remove() {
	if i != nil && i.dir != "" {
		_ = os.RemoveAll(i.dir)
	}
}

// Default returns the user's preferred shell binary path
// (ComSpec on Windows, $SHELL or /bin/bash elsewhere).
func Default() string {
	if runtime.GOOS == "windows" {
		if c := os.Getenv("ComSpec"); c != "" {
			return c
		}
		return "cmd.exe"
	}
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/bash"
}

// Env appends AiT local-mode markers and default prompt env vars to base.
// The result includes AIT_LOCAL_MODE=1, a PS1 set to the local prompt, and
// a PROMPT export (zsh / Windows cmd).
func Env(base []string) []string {
	out := append([]string{}, base...)
	out = append(out,
		"AIT_LOCAL_MODE=1",
		"PS1="+prompt.LocalPS1(),
		"PROMPT="+prompt.Local,
	)
	if runtime.GOOS == "windows" {
		out = append(out, "PROMPT="+prompt.Local+"$P$G")
	}
	return out
}

// Prepare creates silent init scripts for shells that override PS1/PROMPT from env.
// Returns (init, extra argv for the shell binary, extra env, error).
// For non-bash/zsh shells (or Windows) Prepare returns no init dir and no extra args.
func Prepare(shellPath string) (*Init, []string, []string, error) {
	if runtime.GOOS == "windows" {
		return nil, nil, nil, nil
	}

	switch {
	case isZsh(shellPath):
		return prepareZsh()
	case isBash(shellPath):
		return prepareBash()
	default:
		return nil, nil, nil, nil
	}
}

func shellBase(shellPath string) string {
	return strings.ToLower(filepath.Base(shellPath))
}

func isZsh(shellPath string) bool {
	b := shellBase(shellPath)
	return b == "zsh" || strings.HasSuffix(b, "-zsh")
}

func isBash(shellPath string) bool {
	b := shellBase(shellPath)
	return b == "bash" || strings.HasSuffix(b, "-bash")
}
