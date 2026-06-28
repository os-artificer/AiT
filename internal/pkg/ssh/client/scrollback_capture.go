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
	"io"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
)

// RunPTYCapture sends cmd to the remote shell via stdin and returns new scrollback output.
// The command is space-prefixed for ignorespace history suppression so it does not appear
// in bash/zsh history or arrow-key navigation.
// Returns io.ErrClosedPipe when stdin is nil, or empty string when cmd is empty after trimming.
func (c *Client) RunPTYCapture(stdin io.Writer, cmd string) (string, error) {
	if stdin == nil {
		return "", io.ErrClosedPipe
	}
	cmd = strings.TrimRight(cmd, "\r\n")
	if cmd == "" {
		return "", nil
	}

	snap := c.scroll.String()

	// Use space prefix for history suppression, but do NOT suppress echo
	// because the caller (typically AI mode) expects the user to see the
	// command and its output in the terminal.
	injectAt := time.Now()
	if err := c.Inject(stdin, cmd, InjectOpts{SuppressEcho: false}); err != nil {
		return "", err
	}

	c.waitRemoteOutputQuietAfter(injectAt, captureQuietForClient(c), scrollback.CaptureMaxWait)

	return scrollback.Delta(snap, c.scroll.String()), nil
}

// scrollDelta is retained as a private wrapper for the existing client-package tests.
func scrollDelta(before, after string) string {
	return scrollback.Delta(before, after)
}
