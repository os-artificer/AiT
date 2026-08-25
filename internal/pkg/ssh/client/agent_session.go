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
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// AgentSessionHandles exposes remote PTY capture for an Agent REPL (no shell stdin loop).
type AgentSessionHandles struct {
	Stdin      io.Writer
	RunCapture func(cmd string) (string, error)
	Scrollback func() string
	HostLabel  string
}

// AgentSessionOpts configures RunAgentSession.
// Unlike InteractiveOpts, there is no meta scanner, !! handling, or stdin forward shell loop.
type AgentSessionOpts struct {
	HostLabel string
	// OnReady runs after the remote shell is ready. When it returns, the session tears down.
	OnReady func(ctx context.Context, h AgentSessionHandles) error
}

// agentSessionUsesMeta is false: RunAgentSession never installs meta line scanning (OA16).
const agentSessionUsesMeta = false

// AgentPTY is a live remote shell PTY without meta scanning (for Ops AI connect).
type AgentPTY struct {
	Handles    AgentSessionHandles
	sess       *ssh.Session
	wg         sync.WaitGroup
	stopResize func()
}

// Close tears down the remote shell session.
func (p *AgentPTY) Close() error {
	if p == nil {
		return nil
	}
	if p.stopResize != nil {
		p.stopResize()
		p.stopResize = nil
	}
	var err error
	if p.sess != nil {
		err = p.sess.Close()
		p.sess = nil
	}
	p.wg.Wait()
	return err
}

// StartAgentPTY opens a remote PTY shell without meta/!! or stdin forwarding.
// Caller must Close the returned AgentPTY. Capture via Handles.RunCapture.
func (c *Client) StartAgentPTY(hostLabel string) (*AgentPTY, error) {
	_ = agentSessionUsesMeta

	sess, err := c.conn.NewSession()
	if err != nil {
		return nil, err
	}

	fd := int(os.Stdin.Fd())
	cols, rows := 80, 24
	if term.IsTerminal(fd) {
		cols, rows, _ = term.GetSize(fd)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sess.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		sess.Close()
		return nil, err
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}

	c.interactiveStdin = stdin
	// Keep remote PTY output in scrollback only (no interactive shell display).
	c.pauseRemoteOut.Store(true)

	pty := &AgentPTY{sess: sess}
	copyOut := func(r io.Reader) {
		buf := make([]byte, ioCopyBufSize)
		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				c.handleRemoteOutput(buf[:n])
			}
			if readErr != nil {
				return
			}
		}
	}
	pty.wg.Add(2)
	go func() { defer pty.wg.Done(); copyOut(stdout) }()
	go func() { defer pty.wg.Done(); copyOut(stderr) }()

	if err := sess.Shell(); err != nil {
		sess.Close()
		pty.wg.Wait()
		return nil, err
	}

	if term.IsTerminal(fd) {
		pty.stopResize = watchResize(newSessionResize(sess, c), fd)
	}

	attachAt := time.Now()
	c.waitForShellReadyBeforeInject(attachAt, c.promptUser(), c.promptHost(), shellReadyMaxWait)
	c.pauseRemoteOut.Store(true)

	if hostLabel == "" {
		hostLabel = c.opts.User + "@" + c.opts.Host
	}
	pty.Handles = AgentSessionHandles{
		Stdin: stdin,
		RunCapture: func(cmd string) (string, error) {
			return c.RunPTYCapture(stdin, cmd)
		},
		Scrollback: func() string { return c.scroll.String() },
		HostLabel:  hostLabel,
	}
	return pty, nil
}

// RunAgentSession opens a remote PTY shell without meta/!! scanning or interactive stdin forwarding.
// Remote output is retained in scrollback for RunPTYCapture. Caller drives the Agent REPL via OnReady.
func (c *Client) RunAgentSession(ctx context.Context, opts AgentSessionOpts) error {
	if opts.OnReady == nil {
		return fmt.Errorf("AgentSessionOpts.OnReady is required")
	}

	pty, err := c.StartAgentPTY(opts.HostLabel)
	if err != nil {
		return err
	}
	defer pty.Close()

	done := make(chan error, 1)
	go func() { done <- opts.OnReady(ctx, pty.Handles) }()

	select {
	case <-ctx.Done():
		_ = pty.Close()
		return ctx.Err()
	case err := <-done:
		return err
	}
}
