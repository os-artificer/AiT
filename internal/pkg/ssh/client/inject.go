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
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
)

// InjectOpts configures silent PTY command injection behavior.
type InjectOpts struct {
	// CaptureOutput, when true, wraps the command with unique delimiters
	// for output capture from scrollback.
	CaptureOutput bool
	// SuppressEcho, when true, wraps the command with stty -echo/echo
	// so the injected command and its output are not shown to the user.
	// Set to false when the user should see the output (e.g. AI mode
	// RunCapture where the user initiated the command).
	SuppressEcho bool
}

// Inject sends cmd through the PTY stdin pipe with history and echo suppression.
// Commands are space-prefixed (ignorespace) so they do not appear in bash/zsh
// history. When SuppressEcho is true, the command is wrapped with stty -echo/echo.
// When CaptureOutput is true, unique delimiters are injected before/after the
// command output for capture from scrollback.
// Returns io.ErrClosedPipe when stdin is nil, or the write error from stdin.
func (c *Client) Inject(stdin io.Writer, cmd string, opts InjectOpts) error {
	if stdin == nil {
		return io.ErrClosedPipe
	}

	cmd = strings.TrimRight(cmd, "\r\n")
	if cmd == "" {
		return nil
	}

	parts := buildInjectParts(cmd, opts)
	combined := strings.Join(parts, "; ") + "\n"

	if _, err := io.WriteString(stdin, combined); err != nil {
		return err
	}

	if opts.CaptureOutput || opts.SuppressEcho {
		injectAt := time.Now()
		c.waitRemoteOutputQuietAfter(injectAt, scrollback.CaptureQuiet, scrollback.CaptureMaxWait)
	}

	return nil
}

// InjectCapture sends cmd through the PTY stdin pipe with full suppression and
// output capture, then returns the captured output from scrollback.
// This is a convenience wrapper around Inject with CaptureOutput=true and
// SuppressEcho=true, followed by output extraction between capture delimiters.
// Returns the captured output between delimiters, or io.ErrClosedPipe when
// stdin is nil, or write errors from stdin.
func (c *Client) InjectCapture(stdin io.Writer, cmd string) (string, error) {
	if stdin == nil {
		return "", io.ErrClosedPipe
	}

	cmd = strings.TrimRight(cmd, "\r\n")
	if cmd == "" {
		return "", nil
	}

	var snap string
	if c.scroll != nil {
		snap = c.scroll.String()
	}

	if err := c.Inject(stdin, cmd, InjectOpts{CaptureOutput: true, SuppressEcho: true}); err != nil {
		return "", err
	}

	if c.scroll == nil {
		return "", fmt.Errorf("%w", ErrInjectCaptureNoMarkers)
	}

	delta := scrollback.Delta(snap, c.scroll.String())
	c.scroll.StripSubstrings(injectStripFragments()...)
	if !captureMarkersPresent(delta) {
		return "", fmt.Errorf("%w", ErrInjectCaptureNoMarkers)
	}
	return parseCapturedOutput(delta), nil
}
