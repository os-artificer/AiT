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

package sshclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

// Options for SSH dial.
type Options struct {
	Host         string
	Port         int
	User         string
	IdentityPath string
	Password     string
	UseAgent     bool
}

// Client is an SSH connection with optional interactive shell.
type Client struct {
	conn      *ssh.Client
	agentConn net.Conn
	opts      Options
	scroll    *terminal.Buffer
	remoteCwd string
}

// Dial connects to the remote host.
func Dial(opts Options) (*Client, error) {
	if opts.User == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		opts.User = u.Username
	}
	if opts.Port == 0 {
		opts.Port = 22
	}
	auth, agentConn, err := buildAuth(opts)
	if err != nil {
		return nil, err
	}
	hkc, err := hostKeyCallback()
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: hkc,
		Timeout:         15 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	conn, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, err
	}
	return &Client{conn: conn, agentConn: agentConn, opts: opts, scroll: &terminal.Buffer{}}, nil
}

func buildAuth(opts Options) ([]ssh.AuthMethod, net.Conn, error) {
	var methods []ssh.AuthMethod
	var agentConn net.Conn

	if opts.Password != "" {
		methods = append(methods, ssh.Password(opts.Password))
	}

	if opts.IdentityPath != "" {
		signer, err := loadSigner(opts.IdentityPath, opts.Password)
		if err != nil {
			return nil, agentConn, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	} else {
		home, _ := os.UserHomeDir()
		for _, name := range []string{"id_ed25519", "id_rsa"} {
			p := filepath.Join(home, ".ssh", name)
			if _, err := os.Stat(p); err == nil {
				if signer, err := loadSigner(p, opts.Password); err == nil {
					methods = append(methods, ssh.PublicKeys(signer))
					break
				}
			}
		}
	}

	if opts.UseAgent {
		if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
			conn, err := net.Dial("unix", sock)
			if err == nil {
				agentConn = conn
				ag := agent.NewClient(conn)
				methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
			}
		}
	}

	if len(methods) == 0 {
		return nil, agentConn, fmt.Errorf("no authentication method available")
	}
	return methods, agentConn, nil
}

func loadSigner(path, passphrase string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(b, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(b)
}

// Conn returns underlying ssh.Client.
func (c *Client) Conn() *ssh.Client {
	return c.conn
}

// Scrollback returns the output buffer.
func (c *Client) Scrollback() *terminal.Buffer {
	return c.scroll
}

// RemoteCwd returns last known remote working directory.
func (c *Client) RemoteCwd() string {
	return c.remoteCwd
}

// RunCommand runs a non-interactive remote command and returns combined output.
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
	AI              *ai.Client
	OnSuspend       func(scroll *terminal.Buffer, cwd string, cols, rows int) error
}

// RunInteractive runs shell with meta lines (\ait suspend) and the !! AI prefix.
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

	cols, rows, _ := term.GetSize(fd)
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
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

	if opts.TmuxEnabled {
		if err := tmux.Ensure(c.conn, opts.TmuxName); err != nil {
			sess.Close()
			return err
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

	stopResize := watchResize(sess, fd)
	defer stopResize()

	// Brief pause so the remote shell is ready before injecting cd commands.
	time.Sleep(150 * time.Millisecond)

	for _, line := range opts.InitialCommands {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		if _, err := stdin.Write([]byte(line + "\n")); err != nil {
			sess.Close()
			return err
		}
	}

	var wg sync.WaitGroup
	copyOut := func(r io.Reader) {
		buf := make([]byte, 32*1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				_, _ = c.scroll.Write(buf[:n])
				_, _ = os.Stdout.Write(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}

	stdinCh, stopStdin := startStdinReader()
	shutdown := func() {
		stopStdin()
		_ = sess.Close()
		wg.Wait()
	}
	defer shutdown()

	wg.Add(2)
	go func() { defer wg.Done(); copyOut(stdout) }()
	go func() { defer wg.Done(); copyOut(stderr) }()

	waitCh := make(chan error, 1)
	go func() { waitCh <- sess.Wait() }()

	done := make(chan error, 1)
	go func() { done <- c.forwardStdin(ctx, stdin, stdinCh, opts) }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-waitCh:
		if err != nil && err != io.EOF {
			return err
		}
		return nil
	case err := <-done:
		if err == io.EOF {
			return nil
		}
		return err
	}
}

// stdinByte is one byte plus a sticky error from os.Stdin.
type stdinByte struct {
	b   byte
	err error
}

// startStdinReader spawns a goroutine reading os.Stdin one byte at a time.
// The returned channel is closed when stdin reaches EOF or an error occurs.
// stop() asks the goroutine to exit; the goroutine may still be blocked in
// the read syscall until the next byte arrives or the process exits.
func startStdinReader() (<-chan stdinByte, func()) {
	out := make(chan stdinByte, 32)
	stop := make(chan struct{})
	go func() {
		defer close(out)
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				select {
				case out <- stdinByte{b: buf[0]}:
				case <-stop:
					return
				}
			}
			if err != nil {
				select {
				case out <- stdinByte{err: err}:
				case <-stop:
				}
				return
			}
		}
	}()
	return out, func() { close(stop) }
}

// forwardStdin pipes stdin to the remote shell, deferring meta-line handling
// to a metaScanner. Meta openers are only recognized at the start of a line
// AND require a strict full prefix match (see metaPrefixes), so bash !!
// history expansion, words like "\\artifact", or anything typed inside vim
// after a newline are left untouched.
func (c *Client) forwardStdin(ctx context.Context, stdin io.Writer, stdinCh <-chan stdinByte, opts InteractiveOpts) error {
	scan := newMetaScanner()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sb, ok := <-stdinCh:
			if !ok {
				return io.EOF
			}
			if sb.err != nil {
				return sb.err
			}
			fwd, line, hasLine := scan.Feed(sb.b)
			if len(fwd) > 0 {
				if _, err := stdin.Write(fwd); err != nil {
					return err
				}
			}
			if hasLine {
				if err := c.handleMetaLine(ctx, stdin, line, opts); err != nil {
					return err
				}
			}
		}
	}
}

func (c *Client) handleMetaLine(ctx context.Context, stdin io.Writer, line string, opts InteractiveOpts) error {
	line = strings.TrimSpace(line)
	switch {
	case line == `\ait suspend`:
		if opts.OnSuspend != nil {
			fd := int(os.Stdin.Fd())
			cols, rows, _ := term.GetSize(fd)
			if opts.TmuxEnabled {
				_ = tmux.Detach(c.conn, opts.TmuxName)
			}
			cwd := c.captureRemoteCwd(stdin, opts)
			if cwd != "" && pathsafe.RemotePath(cwd) {
				c.remoteCwd = cwd
			}
			c.scroll.StripSubstrings(".__AITPWD__.", ".__AITPWDEND__")
			if err := opts.OnSuspend(c.scroll, c.remoteCwd, cols, rows); err != nil {
				return err
			}
		}
		_ = c.conn.Close()
		return io.EOF
	case strings.HasPrefix(line, "!!"):
		prompt := strings.TrimSpace(strings.TrimPrefix(line, "!!"))
		if prompt == "" {
			fmt.Fprintln(os.Stdout, "usage: !! <question>")
			return nil
		}
		if opts.AI == nil {
			fmt.Fprintln(os.Stdout, "AI not configured (ait config set deepseek.api_key <key>)")
			return nil
		}
		promptCopy := prompt
		go func() {
			ans, err := opts.AI.Assist(ctx, promptCopy, c.scroll.RecentLines(50), c.remoteCwd)
			if err != nil {
				fmt.Fprintf(os.Stdout, "\nAI error: %v\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "\n--- AiT ---\n%s\n-----------\n", ans)
			}
		}()
		return nil
	default:
		fmt.Fprintf(os.Stdout, "AiT: unknown meta command %q (try \\ait suspend or !! <question>)\n", line)
		_, _ = stdin.Write([]byte(line + "\n"))
		return nil
	}
}

// captureRemoteCwd reads cwd from the interactive shell (tmux pane path or pwd marker).
func (c *Client) captureRemoteCwd(stdin io.Writer, opts InteractiveOpts) string {
	if opts.TmuxEnabled {
		if p, err := tmux.PanePath(c.conn, opts.TmuxName); err == nil {
			return strings.TrimSpace(p)
		}
	}
	const begin, end = ".__AITPWD__.", ".__AITPWDEND__."
	snap := c.scroll.String()
	cmd := fmt.Sprintf("printf '%%s%%s%%s\\n' '%s' \"$(pwd -P 2>/dev/null || pwd)\" '%s'\n", begin, end)
	_, _ = stdin.Write([]byte(cmd))
	for i := 0; i < 100; i++ {
		time.Sleep(20 * time.Millisecond)
		if p := parsePwdMarker(c.scroll.String(), snap, begin, end); p != "" {
			return p
		}
	}
	return ""
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

// Close closes the connection.
func (c *Client) Close() error {
	if c.agentConn != nil {
		_ = c.agentConn.Close()
		c.agentConn = nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DetectAuthMethod returns auth method label after successful dial options.
func DetectAuthMethod(opts Options) string {
	if opts.Password != "" {
		return store.AuthMethodPassword
	}
	if opts.IdentityPath != "" {
		return store.AuthMethodKey
	}
	return store.AuthMethodAgent
}
