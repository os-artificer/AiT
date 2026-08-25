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

// Facade aliases re-exporting symbols from the prompt, meta, and pty
// subpackages under their legacy names used throughout the codebase.

package client

import (
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/pty"
)

// BasePrompt is the AiT interactive shell prompt when not connected to a remote host (alias of prompt.Base).
const BasePrompt = prompt.Base

// LocalPrompt is shown in AiT local mode within an SSH session (alias of prompt.Local).
const LocalPrompt = prompt.Local

// metaScanner is an alias for the meta subpackage's Scanner so existing call
// sites in the client package continue to compile without churn.
type metaScanner = meta.Scanner

// stdinByte is the per-byte input record forwarded from the TTY reader.
type stdinByte = pty.StdinByte

// stdinReader is the single-goroutine TTY stdin reader controller.
type stdinReader = pty.Reader

// resizeHandler is implemented by anything that wants window-size updates
// while RunInteractive is running.
type resizeHandler = pty.ResizeHandler

// RemotePrompt returns the AiT display prompt for user@host.
func RemotePrompt(user, host string) string {
	return prompt.Remote(user, host)
}

// RemoteSetupStripFragments returns scrollback/filter fragments for remote PS1 setup leakage.
func RemoteSetupStripFragments(user, host string, dirAbsolute bool) []string {
	return prompt.RemoteSetupStripFragments(user, host, dirAbsolute)
}

// RemotePromptInitialCommands returns shell setup lines to set PS1/PROMPT after the remote shell starts.
func RemotePromptInitialCommands(user, host string, dirAbsolute bool) []string {
	return prompt.RemoteInitialCommands(user, host, dirAbsolute)
}

// RemotePromptRedrawCmd returns a shell one-liner that prints the expanded remote prompt.
func RemotePromptRedrawCmd() string {
	return prompt.RemotePromptRedrawCmd()
}

func remoteSetupPromptMarker(user, host string) string {
	return prompt.RemoteSetupMarker(user, host)
}

func newMetaScanner(echoPrefix string) *metaScanner {
	return meta.New(echoPrefix)
}

func startStdinReader() *stdinReader {
	return pty.Start()
}

func termSize(fd int) (cols, rows int) {
	return pty.TermSize(fd)
}

func watchResize(h resizeHandler, fd int) func() {
	return pty.Watch(h, fd)
}
