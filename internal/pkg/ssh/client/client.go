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
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/localmode"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const (
	shellReadyDelay           = 280 * time.Millisecond
	initialCommandSettleDelay = 120 * time.Millisecond
)

// Options for SSH dial.
type Options struct {
	Host string
	Port int
	User string
	// LoginDestination is the ssh config Host alias or CLI destination (e.g. devhost)
	// before hostname resolution; used to match the active PTY shell for transfer routing.
	LoginDestination string
	IdentityPath     string
	// IdentityExplicit is true when the user passed -i on the CLI.
	IdentityExplicit bool
	Password         string
	UseAgent         bool

	ProxyJump             []string
	ProxyCommand          string
	KnownHostsPath        string
	StrictHostKeyChecking string
	ConnectTimeout        time.Duration
	BatchMode             bool
}

// Client is an SSH connection with optional interactive shell.
type Client struct {
	conn                       *ssh.Client
	agentConn                  net.Conn
	opts                       Options
	scroll                     *terminal.Buffer
	remoteCwd                  string
	localShell                 localmode.Shell
	pauseRemoteOut             atomic.Bool
	lastRemoteOut              atomic.Int64
	remoteSetupSuppress        atomic.Bool
	setupInjectDone            atomic.Bool
	setupLeakGraceUntil        atomic.Int64
	setupKeystrokeGraceUntil   atomic.Int64
	metaRecoverWhitespaceUntil atomic.Int64
	remotePromptShown          atomic.Bool
	remotePromptReadySeen      atomic.Bool
	remotePromptDirAbs         bool
	initialRemoteCwd           string
	setupFilter                setupOutputFilter
	setupGraceStripper         setupStdoutStripper
	stdinReader                *stdinReader // set for RunInteractive; used during remote AI REPL
	chainMode                  atomic.Bool  // false by default; PTY-pipe injection when chain mode is on
	zmodemActive               atomic.Bool
	zmodemSession              *zmodem.Session
	transferMu                 sync.Mutex
	transferCancel             context.CancelFunc
	transferActive             atomic.Bool

	// via holds ProxyJump hop clients that must stay open for the lifetime of
	// this client's forwarded TCP channel. Closed after c.conn in Close.
	via []*Client

	activePromptUser      string
	activePromptHost      string
	forwardLineBuf        []byte
	interactiveStdin      io.Writer
	interactiveOut        io.Writer // non-nil during RunInteractive raw TTY; CRLF-safe stdout
	nestedMu              sync.Mutex
	nestedState           nestedPromptState
	pendingNestedUser     string
	pendingNestedHost     string
	nestedAwaitingSince   time.Time
	nestedMOTDSeen        bool
	nestedMOTDAt          time.Time
	nestedShellReadySeen  bool
	nestedPasswordPending bool
	nestedRemoteHost      string
}

// Conn returns the underlying golang.org/x/crypto/ssh client connection.
func (c *Client) Conn() *ssh.Client {
	return c.conn
}

// Scrollback returns the in-memory terminal output buffer for this session.
func (c *Client) Scrollback() *terminal.Buffer {
	return c.scroll
}

// RemoteCwd returns the last captured remote working directory (from suspend or pwd probe).
// May be empty until the interactive loop updates it.
func (c *Client) RemoteCwd() string {
	return c.remoteCwd
}

// ChainMode reports whether chain mode is active.
// When true, SFTP and conn.NewSession-based operations are replaced with PTY-pipe
// alternatives that traverse nested SSH hops in the interactive shell.
func (c *Client) ChainMode() bool {
	return c.chainMode.Load()
}

// SetChainMode enables or disables chain mode. When enabled, file transfer and tmux
// operations use PTY-pipe injection instead of direct SFTP/session.
func (c *Client) SetChainMode(enabled bool) {
	c.chainMode.Store(enabled)
}

// RunCommand runs cmd without a PTY and returns combined stdout/stderr as a string.
// Returns NewSession errors, or the remote command exit/error from CombinedOutput.
func (c *Client) RunCommand(cmd string) (string, error) {
	sess, err := c.conn.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}

// InteractiveOpts configures interactive session.
type InteractiveOpts struct {
	SessionID   string
	TmuxEnabled bool
	TmuxName    string
	// InitialCommands are sent into the interactive shell stdin right after
	// the shell starts (e.g. "cd /restored/path"). Each entry is sent as a
	// separate line.
	InitialCommands []string
	// RemotePromptDirAbsolute selects absolute vs short (basename) cwd in the remote PS1.
	RemotePromptDirAbsolute bool
	OnSuspend               func(scroll *terminal.Buffer, cwd string, cols, rows int) error
	// OnMetaCLI runs aissh CLI subcommands from !ait <args> (args exclude the prefix).
	OnMetaCLI func(args []string) error
	// AliasLookup resolves saved session aliases for chain connect meta lines (!ait prod).
	AliasLookup cliroute.AliasLookup
	// InitialRemoteCwd is the saved remote cwd for client-painted first prompt on load/switch.
	InitialRemoteCwd string
	// ReplayScrollback is sanitized scrollback from a prior suspend, seeded into the
	// in-memory scroll buffer after setup (load/switch). Not written to stdout.
	// Empty skips seeding.
	ReplayScrollback string
}

// RunInteractive runs a PTY shell with !ait meta handling, optional tmux, and local mode.
// ctx cancels the session loop. Requires a TTY on stdin. Returns terminal, session, tmux, or OnSuspend errors.
func (c *Client) RunInteractive(ctx context.Context, opts InteractiveOpts) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		sess.Close()
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		sess.Close()
		return fmt.Errorf("terminal: %w", err)
	}
	defer func() {
		if oldState != nil {
			_ = term.Restore(fd, oldState)
		}
	}()

	c.interactiveOut = terminal.NewCRLFWriter(os.Stdout)
	defer func() { c.interactiveOut = nil }()

	pendingReplay := opts.ReplayScrollback

	cols, rows, _ := term.GetSize(fd)

	echo := uint32(1)
	if len(opts.InitialCommands) > 0 {
		echo = 0 // restored by stty echo in setup command
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          echo,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sess.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		sess.Close()
		return err
	}
	if cols > 0 && rows > 0 {
		_ = sess.WindowChange(rows, cols)
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		sess.Close()
		return err
	}

	c.initActivePrompt()
	c.interactiveStdin = stdin
	c.nestedState = nestedPromptIdle

	setupNeeded := len(opts.InitialCommands) > 0
	postAttachCmds := opts.InitialCommands
	if setupNeeded {
		c.beginRemotePromptSetup(opts.RemotePromptDirAbsolute, opts.InitialRemoteCwd)
	}

	var wg sync.WaitGroup

	copyOut := func(r io.Reader) {
		buf := make([]byte, ioCopyBufSize)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				c.handleRemoteOutput(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}

	wg.Add(2)
	go func() { defer wg.Done(); copyOut(stdout) }()
	go func() { defer wg.Done(); copyOut(stderr) }()

	if opts.TmuxEnabled {
		if err := c.tmuxEnsure(stdin, opts.TmuxName); err != nil {
			sess.Close()
			return err
		}
		if err := c.ensureTmuxSessionReady(stdin, opts.TmuxName); err != nil {
			sess.Close()
			return err
		}
		if setupNeeded {
			setup, after := partitionInitialCommands(opts.InitialCommands, c.opts.User, c.opts.Host)
			if err := c.tmuxPreAttachSetup(opts.TmuxName, setup); err != nil {
				sess.Close()
				return wrapTmuxPreAttachErr(opts.TmuxName, err)
			}
			postAttachCmds = after
		}
		if err := sess.Start(tmux.AttachCommand(opts.TmuxName)); err != nil {
			sess.Close()
			return err
		}
	} else {
		if err := sess.Shell(); err != nil {
			sess.Close()
			return err
		}
	}

	stopResize := watchResize(newSessionResize(sess, c), fd)

	defer stopResize()

	shutdown := func() {
		if c.stdinReader != nil {
			c.stdinReader.Stop()
		}
		_ = sess.Close()
		wg.Wait()
	}
	defer shutdown()

	// Wait for login shell / .bashrc to finish while output is suppressed.
	attachAt := time.Now()
	if opts.TmuxEnabled {
		time.Sleep(shellReadyDelay)
	} else {
		c.waitForShellReadyBeforeInject(attachAt, c.promptUser(), c.promptHost(), shellReadyMaxWait)
	}

	if len(postAttachCmds) > 0 {
		if err := c.injectInitialCommands(stdin, postAttachCmds); err != nil {
			return err
		}
	} else if setupNeeded {
		c.setupInjectDone.Store(true)
	}

	if setupNeeded && !UseRemoteSystemPrompt() {
		c.runRemotePromptSetup(stdin, c.opts.User, c.opts.Host, opts.TmuxEnabled)
	}

	if UseRemoteSystemPrompt() && !c.remotePromptShown.Load() {
		c.finishNativePromptAfterAttach()
	}

	if pendingReplay != "" {
		c.seedReplayScrollback(pendingReplay, c.promptUser(), c.promptHost())
	}

	drainInteractiveStdin()
	c.stdinReader = startStdinReader()
	stdinCh := c.stdinReader.Ch

	waitCh := make(chan error, 1)

	go func() { waitCh <- sess.Wait() }()

	done := make(chan error, 1)
	go func() { done <- c.forwardStdin(ctx, stdin, stdinCh, opts) }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-waitCh:
		if err != nil && err != io.EOF {
			return wrapInteractiveWaitErr(opts.TmuxEnabled, opts.TmuxName, err)
		}
		return nil
	case err := <-done:
		if err == io.EOF {
			return nil
		}
		return err
	}
}

// injectInitialCommands sends setup lines (PS1 export, cd restore) while
// suppressing their echo on the local terminal.
func (c *Client) injectInitialCommands(stdin io.Writer, cmds []string) error {
	if len(cmds) == 0 {
		return nil
	}

	graceEnd := time.Now().Add(setupLeakGraceMax)
	c.setupLeakGraceUntil.Store(graceEnd.UnixNano())
	c.setupKeystrokeGraceUntil.Store(graceEnd.UnixNano())
	c.setupGraceStripper.reset()

	injectAt := time.Now()

	for _, line := range cmds {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		if err := c.Inject(stdin, line, InjectOpts{SuppressEcho: true}); err != nil {
			return err
		}
	}

	time.Sleep(initialCommandSettleDelay)
	c.waitRemoteOutputQuietAfter(injectAt, remoteOutputQuietPeriod, remoteOutputQuietMaxWait)
	c.waitRemoteOutputQuietSince(injectAt, remoteOutputQuietPeriod, remoteOutputQuietMaxWait)
	c.setupInjectDone.Store(true)
	return nil
}

// forwardStdin pipes stdin to the remote shell, deferring !ait meta-line handling
// to a metaScanner. Line-start !! is forwarded to the remote shell (bash history).
func (c *Client) forwardStdin(ctx context.Context, stdin io.Writer, stdinCh <-chan stdinByte, opts InteractiveOpts) error {
	scan := newMetaScanner(c.metaLineEchoPrefix())
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sb, ok := <-stdinCh:
			if !ok {
				return io.EOF
			}
			res := c.processForwardStdinByte(ctx, stdin, sb, scan, opts)
			if res.action == fwdStdinReturn {
				return res.err
			}
		}
	}
}

func (c *Client) handleMetaLine(ctx context.Context, stdin io.Writer, line string, opts InteractiveOpts) error {
	line = strings.TrimSpace(line)
	switch {
	case line == metaLineSuspend:
		if err := c.handleSuspendMeta(stdin, opts); err != nil {
			return err
		}
		_ = c.Close()
		return io.EOF
	case line == metaLineLocal:
		return c.enterLocalMode()
	case line == metaLineRemote:
		return c.exitLocalMode()
	case line == metaLineChain || strings.HasPrefix(line, metaLineChain+" "):
		return c.handleChainMeta(stdin, line)

	default:

		if chain, ok := ParseMetaChainLine(line, opts.AliasLookup); ok {
			if err := c.handleSuspendMeta(stdin, opts); err != nil {
				return err
			}
			_ = c.Close()
			return chain
		}

		if line == metaLineBare || line == metaLineHelp {
			return c.runMetaCLI(stdin, opts, nil)
		}
		if cmd, args, isXfer := transferArgsFromMetaLine(line); isXfer {
			switch cmd {
			case "get":
				return c.handleTransferGet(ctx, stdin, args)
			case "put":
				return c.handleTransferPut(ctx, stdin, args)
			}
		}
		if strings.HasPrefix(line, metaPrefixAitSpace) {
			args := strings.Fields(strings.TrimPrefix(line, metaPrefixAitSpace))
			return c.runMetaCLI(stdin, opts, args)
		}
		termPrintln(os.Stdout, "AiT: unknown meta command (try "+toolname.MetaLine("help")+")")
		c.finishMetaInteraction(stdin)
		return nil
	}
}

func (c *Client) runMetaCLI(stdin io.Writer, opts InteractiveOpts, args []string) error {
	if opts.OnMetaCLI == nil {
		termPrintln(os.Stdout, "AiT: meta CLI not available (try "+toolname.MetaLine("help")+")")
		c.finishMetaInteraction(stdin)
		return nil
	}

	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	if err := opts.OnMetaCLI(args); err != nil {
		errhint.Print(os.Stderr, errhint.SurfaceTTY, err)
	}
	c.finishMetaInteraction(stdin)
	return nil
}

func (c *Client) enterLocalMode() error {
	if c.localShell.Running() {
		aitModeNotice("already in local mode (type " + toolname.MetaLine("remote to return") + ")")
		return nil
	}

	if c.localShell.Active() {
		c.localShell.Stop()
	}

	fd := int(os.Stdin.Fd())
	cols, rows := termSize(fd)

	c.pauseRemoteOut.Store(true)
	if err := c.localShell.Start(cols, rows, c.onLocalShellExit); err != nil {
		c.pauseRemoteOut.Store(false)
		return fmt.Errorf("local shell: %w", err)
	}

	aitModeNotice(
		"local mode",
		"Prompt is \""+LocalPrompt+"\" (commands run on this machine; SSH stays connected).",
		"Type "+toolname.MetaLine("remote to return")+".",
	)
	return nil
}

func (c *Client) onLocalShellExit() {
	c.pauseRemoteOut.Store(false)
	aitModeNotice("local shell exited — remote mode restored.")
}

func (c *Client) exitLocalMode() error {
	if c.localShell.Running() {
		c.localShell.Stop()
		c.pauseRemoteOut.Store(false)
		aitModeNotice("remote mode restored.")
		return nil
	}

	if c.localShell.Active() {
		c.localShell.Stop()
	}

	aitModeNotice("not in local mode (type " + toolname.MetaLine("local to switch") + ")")
	return nil
}

// captureRemoteCwd reads cwd from the interactive shell (tmux pane path or pwd marker).
func (c *Client) captureRemoteCwd(stdin io.Writer, opts InteractiveOpts) string {
	if opts.TmuxEnabled {
		if p, err := c.tmuxPanePath(stdin, opts.TmuxName); err == nil {
			return strings.TrimSpace(p)
		}
	}

	pwd, err := c.probeShellPwd(stdin)
	if err != nil {
		return ""
	}
	return pwd
}

// Close closes the SSH agent connection (if any), the underlying ssh.Client,
// then any ProxyJump hop clients retained in via (innermost hop first).
// Returns the ssh.Client Close error when a connection was open.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	var err error
	if c.agentConn != nil {
		_ = c.agentConn.Close()
		c.agentConn = nil
	}
	if c.conn != nil {
		err = c.conn.Close()
		c.conn = nil
	}
	for i := len(c.via) - 1; i >= 0; i-- {
		_ = c.via[i].Close()
	}
	c.via = nil
	return err
}
