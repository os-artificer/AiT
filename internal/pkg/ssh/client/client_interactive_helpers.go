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
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	xterm "golang.org/x/term"
)

// termEOL is the line ending for messages written while stdin is in raw mode.
// LF-only would leave the cursor at the previous column ("staircase" layout).
const termEOL = msgCRLF

const termEraseEOL = "\x1b[K"

// forwardStdinAction describes how one stdin byte should be handled.
type forwardStdinAction int

const (
	fwdStdinContinue forwardStdinAction = iota
	fwdStdinReturn
)

type forwardStdinResult struct {
	action forwardStdinAction
	err    error
}

func termPrintln(w io.Writer, s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", termEOL)
	if !strings.HasSuffix(s, termEOL) {
		s += termEOL
	}
	_, _ = io.WriteString(w, s)
}

// termPrintInline overwrites the current terminal line without advancing (TTY progress).
func termPrintInline(w io.Writer, s string) {
	_, _ = io.WriteString(w, "\r"+s+termEraseEOL)
}

func termPrintf(w io.Writer, format string, args ...any) {
	termPrintln(w, fmt.Sprintf(format, args...))
}

// writeInteractiveStdout writes remote PTY bytes to the interactive stdout sink.
// During RunInteractive raw TTY, interactiveOut applies CRLF translation (K23).
func (c *Client) writeInteractiveStdout(p []byte) (int, error) {
	w := io.Writer(os.Stdout)
	if c.interactiveOut != nil {
		w = c.interactiveOut
	}
	return w.Write(p)
}

func (c *Client) refreshRemotePrompt(stdin io.Writer) {
	if stdin == nil || c.localShell.Active() {
		return
	}

	wasPaused := c.pauseRemoteOut.Load()
	if !wasPaused {
		c.pauseRemoteOut.Store(true)
	}

	_ = c.Inject(stdin, RemotePromptRedrawCmd(), InjectOpts{SuppressEcho: true})

	if !wasPaused {
		c.pauseRemoteOut.Store(false)
	}
}

// aitModeNotice prints a multi-line AiT status block on stderr. It also emits
// a newline on stdout so a partial remote prompt or meta echo does not share a
// line with the local shell's init output (e.g. nvm).
func aitModeNotice(lines ...string) {
	_, _ = os.Stdout.Write([]byte(termEOL))

	var b strings.Builder
	b.WriteString(termEOL)
	for _, line := range lines {
		b.WriteString(errhint.PrefixAiT)
		b.WriteString(line)
		b.WriteString(termEOL)
	}
	b.WriteString(termEOL)
	_, _ = os.Stderr.WriteString(b.String())
}

func (c *Client) processForwardStdinByte(
	ctx context.Context,
	stdin io.Writer,
	sb stdinByte,
	scan *metaScanner,
	opts InteractiveOpts,
) forwardStdinResult {
	if sb.Err != nil {
		return forwardStdinResult{action: fwdStdinReturn, err: sb.Err}
	}

	scan.SyncFrom(c.localShell.Active(), c.metaLineEchoPrefix())

	fwd, echo, line, hasLine := scan.Feed(sb.B)
	if len(echo) > 0 {
		_, _ = c.writeInteractiveStdout(echo)
		if c.setupKeystrokeGraceUntil.Load() != 0 {
			c.setupKeystrokeGraceUntil.Store(0)
		}
	}

	if hasLine {
		err := c.handleMetaLine(ctx, stdin, line, opts)
		if err != nil {
			return forwardStdinResult{action: fwdStdinReturn, err: err}
		}
		return forwardStdinResult{action: fwdStdinContinue}
	}

	if len(fwd) == 0 {
		return forwardStdinResult{action: fwdStdinContinue}
	}

	if c.localShell.Active() {
		if c.setupKeystrokeGraceUntil.Load() != 0 {
			c.setupKeystrokeGraceUntil.Store(0)
		}
		err := c.localShell.Write(fwd)
		if err != nil {
			return forwardStdinResult{action: fwdStdinReturn, err: err}
		}
		return forwardStdinResult{action: fwdStdinContinue}
	}

	// End keystroke grace before forwarding user input so echo is visible; leak
	// filtering continues until setupLeakGraceMax regardless of keystrokes.
	if c.setupKeystrokeGraceUntil.Load() != 0 {
		c.setupKeystrokeGraceUntil.Store(0)
	}
	if _, err := stdin.Write(fwd); err != nil {
		return forwardStdinResult{action: fwdStdinReturn, err: err}
	}
	c.noteInteractiveInput(fwd)
	return forwardStdinResult{action: fwdStdinContinue}
}

func (c *Client) handleSuspendMeta(stdin io.Writer, opts InteractiveOpts) error {
	if opts.OnSuspend == nil {
		return nil
	}

	fd := int(os.Stdin.Fd())
	cols, rows, _ := xterm.GetSize(fd)

	if opts.TmuxEnabled {
		_ = c.tmuxDetach(stdin, opts.TmuxName)
	}

	cwd := c.captureRemoteCwd(stdin, opts)
	if cwd != "" && pathsafe.RemotePath(cwd) {
		c.remoteCwd = cwd
	}

	// Filter leaked setup/probe echo lines before the scrollback is saved so
	// the next load/switch replay never shows them.
	user, host := c.promptUser(), c.promptHost()
	c.scroll.ReplaceWith(func(s string) string {
		return StripSetupLeakLines(s, user, host)
	})
	c.scroll.StripSubstrings(".__AITPWD__.", ".__AITPWDEND__")
	return opts.OnSuspend(c.scroll, c.remoteCwd, cols, rows)
}

func parsePwdMarker(full, before, begin, end string) string {
	if len(full) <= len(before) {
		return ""
	}
	chunk := full[len(before):]
	i := strings.Index(chunk, begin)
	if i < 0 {
		return ""
	}
	j := strings.Index(chunk[i+len(begin):], end)
	if j < 0 {
		return ""
	}
	return strings.TrimSpace(chunk[i+len(begin) : i+len(begin)+j])
}

// handleChainMeta handles the !ait chain meta command.
// !ait chain       — show current chain mode status
// !ait chain on    — enable chain mode (PTY-based transfer)
// !ait chain off   — disable chain mode (SFTP-based transfer)
func (c *Client) handleChainMeta(stdin io.Writer, line string) error {
	arg := strings.TrimPrefix(line, metaLineChain)
	arg = strings.TrimSpace(arg)

	switch arg {
	case "on":
		c.SetChainMode(true)
		aitModeNotice("chain mode enabled",
			"File transfer and tmux will use PTY-pipe injection.",
		)
	case "off":
		c.SetChainMode(false)
		aitModeNotice("chain mode disabled",
			"File transfer and tmux will use SFTP/direct sessions.",
			"Use "+toolname.MetaLine("chain on")+" to re-enable.",
		)
	default:
		// "status" or bare "!ait chain"
		status := "off"
		if c.ChainMode() {
			status = "on"
		}
		aitModeNotice(
			"chain mode: "+status,
			"Use "+toolname.MetaLine("chain on")+" to enable, "+toolname.MetaLine("chain off")+" to disable.",
			"Chain mode uses PTY-pipe injection for file transfer and tmux,",
			"supporting any number of SSH hops without extra tools.",
		)
	}

	c.finishMetaInteraction(stdin)
	return nil
}

// handleRemoteOutput processes one chunk of remote PTY output.
// When ZMODEM is active, output is fed to the ZMODEM session instead.
// Filtered output is appended to scrollback and written to stdout unless paused.
func (c *Client) handleRemoteOutput(buf []byte) {
	c.touchRemoteOut()

	if c.zmodemActive.Load() {
		c.feedZModemOutput(buf)
		return
	}

	c.maybeInjectNestedPrompt(buf)

	user := c.promptUser()
	host := c.promptHost()

	chunkStr := string(buf)
	closedHost := parseSSHConnectionClosedHost(chunkStr)

	chunk, repaintPrompt := c.filterRemoteOutputChunkRepaint(buf, user, host)
	nestedClosed := false
	if closedHost != "" && c.onNestedSSHSessionClosed(closedHost) {
		nestedClosed = true
		repaintPrompt = true
		user = c.promptUser()
		host = c.promptHost()
	}
	if chunk == nil {
		if UseRemoteSystemPrompt() && !c.remotePromptShown.Load() {
			c.tryEmitNativePromptIfNeeded(user, host)
		}
		if repaintPrompt {
			if nestedClosed {
				c.repaintAfterNestedSSHClose(user, host)
			} else {
				c.repaintInteractivePrompt(user, host)
			}
		}
		return
	}

	// While paused (setup phase) raw output is buffered in scrollback only;
	// finishRemoteSetup needs the unfiltered prompt there.
	if c.pauseRemoteOut.Load() {
		if containsPromptReadySentinel(string(chunk)) {
			c.remotePromptReadySeen.Store(true)
		}
		_, _ = c.scroll.Write(chunk)
		return
	}

	// Scroll and stdout share the grace-filtered bytes so leaked setup echo
	// never reaches the saved scrollback either (replayed on next load/switch).
	stdoutChunk := c.applySetupGrace(chunk, user, host)
	if len(stdoutChunk) == 0 {
		if UseRemoteSystemPrompt() && !c.remotePromptShown.Load() {
			c.tryEmitNativePromptIfNeeded(user, host)
		}
		if repaintPrompt {
			if nestedClosed {
				c.repaintAfterNestedSSHClose(user, host)
			} else {
				c.repaintInteractivePrompt(user, host)
			}
		}
		return
	}
	if c.metaRecoverWhitespaceUntil.Load() > time.Now().UnixNano() && isWhitespaceOnlyPTYOutput(stdoutChunk) {
		return
	}
	_, _ = c.scroll.Write(stdoutChunk)
	_, _ = c.writeInteractiveStdout(stdoutChunk)
	if !UseRemoteSystemPrompt() && !c.remotePromptShown.Load() && isDuplicatePromptOutput(stdoutChunk, user, host) {
		c.remotePromptShown.Store(true)
	}
	if UseRemoteSystemPrompt() && !c.remotePromptShown.Load() {
		if chunkContainsNativePrompt(stdoutChunk, user, host) {
			c.remotePromptShown.Store(true)
		} else {
			c.tryEmitNativePromptIfNeeded(user, host)
		}
	}
	if repaintPrompt {
		if nestedClosed {
			c.repaintAfterNestedSSHClose(user, host)
		} else {
			c.repaintInteractivePrompt(user, host)
		}
	}
}

// applySetupGrace filters setup-leak lines from stdout-bound output while the
// post-setup leak-filter window is active. The window ends after setupLeakGraceMax;
// remotePromptShown and the first keystroke do not end leak filtering early.
func (c *Client) applySetupGrace(chunk []byte, user, host string) []byte {
	until := c.setupLeakGraceUntil.Load()
	if until == 0 {
		return chunk
	}

	if time.Now().UnixNano() >= until {
		c.setupLeakGraceUntil.Store(0)
		if tail := c.setupGraceStripper.endGrace(user, host); len(tail) > 0 {
			return append(tail, chunk...)
		}
		return chunk
	}

	if c.remotePromptShown.Load() && isWhitespaceOnlyPTYOutput(chunk) {
		return nil
	}

	if c.metaRecoverWhitespaceUntil.Load() > time.Now().UnixNano() && c.shouldStripNativePrompt() {
		var stripped bool
		chunk, stripped = stripUnwantedPromptLines(
			chunk, user, host, c.shouldStripNativePrompt(), c.remotePromptShown.Load(),
		)
		if stripped && len(chunk) == 0 {
			return nil
		}
	}

	// After the first keystroke, pass single-byte echo through; strip leak lines from larger chunks.
	if c.setupKeystrokeGraceUntil.Load() == 0 {
		if containsClearTokens(string(chunk)) {
			return chunk
		}
		if len(chunk) == 1 {
			return chunk
		}
		return stripSetupLeakStdout(chunk, user, host)
	}

	allowNative := !c.remotePromptShown.Load()
	return c.setupGraceStripper.filterChunk(chunk, user, host, allowNative)
}
