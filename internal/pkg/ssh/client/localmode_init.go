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
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func defaultLocalShell() string {
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

// localShellEnv appends AiT local-mode markers and default prompt env vars.
func localShellEnv(base []string) []string {
	out := append([]string{}, base...)
	out = append(out,
		"AIT_LOCAL_MODE=1",
		"PS1="+localPromptPS1(),
		"PROMPT="+LocalPrompt,
	)
	if runtime.GOOS == "windows" {
		out = append(out, "PROMPT="+LocalPrompt+"$P$G")
	}
	return out
}

func shellBase(shellPath string) string {
	return strings.ToLower(filepath.Base(shellPath))
}

func isZshShell(shellPath string) bool {
	b := shellBase(shellPath)
	return b == "zsh" || strings.HasSuffix(b, "-zsh")
}

func isBashShell(shellPath string) bool {
	b := shellBase(shellPath)
	return b == "bash" || strings.HasSuffix(b, "-bash")
}

// localShellInit holds a temp dir with rc files sourced at shell start (no PTY echo).
type localShellInit struct {
	dir string
}

func (i *localShellInit) remove() {
	if i != nil && i.dir != "" {
		_ = os.RemoveAll(i.dir)
	}
}

// prepareLocalShellInit creates silent init scripts for shells that override PS1/PROMPT from env.
// Returns extra argv for the shell binary and extra env entries (e.g. ZDOTDIR).
func prepareLocalShellInit(shellPath string) (*localShellInit, []string, []string, error) {
	if runtime.GOOS == "windows" {
		return nil, nil, nil, nil
	}

	switch {
	case isZshShell(shellPath):
		return prepareZshLocalInit()
	case isBashShell(shellPath):
		return prepareBashLocalInit()
	default:
		return nil, nil, nil, nil
	}
}
