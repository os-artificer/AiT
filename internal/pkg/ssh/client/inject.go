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
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
)

const (
	injectCaptureRetryAttempts      = 3
	injectCaptureRetryAttemptsChain = 5
	injectCaptureRetryDelay         = 200 * time.Millisecond
	injectCaptureRetryDelayChain    = 300 * time.Millisecond
	transferChainCaptureQuiet       = 500 * time.Millisecond
)

// injectCaptureForRetry is the capture backend used by injectCaptureRetry (overridable in tests).
var injectCaptureForRetry = func(c *Client, stdin io.Writer, cmd string) (string, error) {
	return c.InjectCapture(stdin, cmd)
}

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
		c.waitRemoteOutputQuietAfter(injectAt, injectQuietForClient(c, opts), scrollback.CaptureMaxWait)
	}

	return nil
}

// InjectZModemCommand starts sz/rz on the remote shell without stty wrap or post-inject quiet wait.
// ZMODEM peers manage the TTY themselves; caller must have BeginZModemSession active.
func (c *Client) InjectZModemCommand(stdin io.Writer, cmd string) error {
	if stdin == nil {
		return io.ErrClosedPipe
	}

	cmd = strings.TrimRight(cmd, "\r\n")
	if cmd == "" {
		return nil
	}

	line := injectSpacePrefix + cmd + "\n"
	_, err := io.WriteString(stdin, line)
	return err
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
	if !captureMarkersPresent(delta) {
		full := c.scroll.String()
		if captureMarkersPresent(full) {
			delta = full
		}
	}
	c.scroll.StripSubstrings(injectStripFragments()...)
	if !captureMarkersPresent(delta) {
		return "", fmt.Errorf("%w", ErrInjectCaptureNoMarkers)
	}
	return parseCapturedOutput(delta), nil
}

// captureQuietForClient returns the post-inject idle wait for capture commands.
// Chain mode uses a longer quiet period to tolerate multi-hop PTY latency.
func captureQuietForClient(c *Client) time.Duration {
	if c != nil && c.ChainMode() {
		return transferChainCaptureQuiet
	}
	return scrollback.CaptureQuiet
}

// injectQuietForClient returns the post-inject idle wait for Inject calls.
func injectQuietForClient(c *Client, opts InjectOpts) time.Duration {
	if opts.CaptureOutput {
		return captureQuietForClient(c)
	}
	if opts.SuppressEcho && c != nil && c.ChainMode() {
		return transferChainCaptureQuiet
	}
	return scrollback.CaptureQuiet
}

// injectCaptureRetry runs InjectCapture up to injectCaptureRetryAttempts times when markers are missing.
func (c *Client) injectCaptureRetry(stdin io.Writer, cmd string) (string, error) {
	attempts := injectCaptureRetryAttempts
	delay := injectCaptureRetryDelay
	if c.ChainMode() {
		attempts = injectCaptureRetryAttemptsChain
		delay = injectCaptureRetryDelayChain
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if c.scroll != nil {
				c.scroll.StripSubstrings(injectStripFragments()...)
			}
			time.Sleep(delay)
		}

		out, err := injectCaptureForRetry(c, stdin, cmd)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !errors.Is(err, ErrInjectCaptureNoMarkers) {
			return "", err
		}
	}
	return "", lastErr
}
