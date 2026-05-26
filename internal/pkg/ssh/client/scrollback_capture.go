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
	injectAt := time.Now()

	if _, err := io.WriteString(stdin, cmd+"\n"); err != nil {
		return "", err
	}

	c.waitRemoteOutputQuietAfter(injectAt, scrollback.CaptureQuiet, scrollback.CaptureMaxWait)
	return scrollback.Delta(snap, c.scroll.String()), nil
}

// scrollDelta is retained as a private wrapper for the existing client-package tests.
func scrollDelta(before, after string) string {
	return scrollback.Delta(before, after)
}
