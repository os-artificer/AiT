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

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
	"github.com/os-artificer/AiT/internal/pkg/tmux"

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

// finishMetaInteraction ends local meta output and asks the remote shell to redraw PS1.
func (c *Client) finishMetaInteraction(stdin io.Writer) {
	_, _ = os.Stdout.Write([]byte(termEOL))
	c.refreshRemotePrompt(stdin)
}

func (c *Client) refreshRemotePrompt(stdin io.Writer) {
	if stdin == nil || c.localShell.Active() {
		return
	}
	_, _ = stdin.Write([]byte("\r"))
}

// aitModeNotice prints a multi-line AiT status block on stderr. It also emits
// a newline on stdout so a partial remote prompt or meta echo does not share a
// line with the local shell's init output (e.g. nvm).
func aitModeNotice(lines ...string) {
	_, _ = os.Stdout.Write([]byte(termEOL))

	var b strings.Builder
	b.WriteString(termEOL)
	for _, line := range lines {
		b.WriteString("AiT: ")
		b.WriteString(line)
		b.WriteString(termEOL)
	}
	b.WriteString(termEOL)
	_, _ = os.Stderr.WriteString(b.String())
}

func aiPromptFromMetaLine(line string) (prompt string, isAI bool) {
	return meta.AIPromptFromLine(line)
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

	fwd, echo, line, hasLine := scan.Feed(sb.B)
	if len(echo) > 0 {
		_, _ = os.Stdout.Write(echo)
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
		err := c.localShell.Write(fwd)
		if err != nil {
			return forwardStdinResult{action: fwdStdinReturn, err: err}
		}
		return forwardStdinResult{action: fwdStdinContinue}
	}

	if _, err := stdin.Write(fwd); err != nil {
		return forwardStdinResult{action: fwdStdinReturn, err: err}
	}
	c.noteInteractiveInput(fwd)
	return forwardStdinResult{action: fwdStdinContinue}
}

func (c *Client) handleAIMetaLine(ctx context.Context, stdin io.Writer, initialQuestion string, opts InteractiveOpts) error {
	if opts.AI == nil {
		errhint.Print(os.Stderr, errhint.SurfaceTTY, ai.NotConfiguredErr())
		c.finishMetaInteraction(stdin)
		return nil
	}

	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	c.finishMetaInteraction(stdin)

	if opts.OnRemoteAI == nil {
		termPrintln(os.Stderr, "AiT: remote AI mode not available")
		return nil
	}

	var replIn io.Reader
	if c.stdinReader != nil {
		replIn = c.stdinReader.BeginREPL()
		defer c.stdinReader.EndREPL()
	}

	err := opts.OnRemoteAI(RemoteAIParams{
		Ctx:             ctx,
		Stdin:           stdin,
		InitialQuestion: initialQuestion,
		AI:              opts.AI,
		HostLabel:       c.opts.User + "@" + c.opts.Host,
		RemoteCwd:       c.remoteCwd,
		Scrollback:      func() string { return c.scroll.RecentLines(50) },
		ReplIn:          replIn,
		RunCapture: func(cmd string) (string, error) {
			return c.RunPTYCapture(stdin, cmd)
		},
		MaxSteps: opts.AgentMaxSteps,
	})

	c.finishMetaInteraction(stdin)
	return err
}

func (c *Client) handleSuspendMeta(stdin io.Writer, opts InteractiveOpts) error {
	if opts.OnSuspend == nil {
		return nil
	}

	fd := int(os.Stdin.Fd())
	cols, rows, _ := xterm.GetSize(fd)

	if opts.TmuxEnabled {
		_ = tmux.Detach(c.conn, opts.TmuxName)
	}

	cwd := c.captureRemoteCwd(stdin, opts)
	if cwd != "" && pathsafe.RemotePath(cwd) {
		c.remoteCwd = cwd
	}

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
			"Use !ait chain on to re-enable.",
		)
	default:
		// "status" or bare "!ait chain"
		status := "off"
		if c.ChainMode() {
			status = "on"
		}
		aitModeNotice(
			"chain mode: "+status,
			"Use !ait chain on to enable, !ait chain off to disable.",
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
	promptMarker := c.promptMarker()

	chunk := c.filterRemoteOutputChunk(buf, user, host, promptMarker)
	if chunk == nil {
		return
	}

	_, _ = c.scroll.Write(chunk)
	if !c.pauseRemoteOut.Load() {
		_, _ = os.Stdout.Write(chunk)
		if !c.remotePromptShown.Load() && isDuplicatePromptOutput(chunk, user, host) {
			c.remotePromptShown.Store(true)
		}
	}
}
