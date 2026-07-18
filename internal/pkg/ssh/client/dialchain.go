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
	"net"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/crypto/ssh"
)

// RemoteExitError is returned when a remote command exits non-zero.
type RemoteExitError struct {
	Code int
}

// Error implements error; the message includes the non-zero remote exit Code.
func (e *RemoteExitError) Error() string {
	return fmt.Sprintf("remote command exited with status %d", e.Code)
}

// DialChain connects to opts.Host, using ProxyJump or ProxyCommand when configured, otherwise Dial.
// opts.User defaults to the current OS user; opts.Port defaults to 22.
// Returns a Client, or errors from user lookup, proxy dial, buildAuth, host key setup, or ssh.NewClientConn.
func DialChain(opts Options) (*Client, error) {
	if err := normalizeDialOptions(&opts); err != nil {
		return nil, err
	}

	if len(opts.ProxyJump) == 0 && opts.ProxyCommand == "" {
		return Dial(opts)
	}

	targetAddr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)

	var conn net.Conn
	var via []*Client
	var err error

	if opts.ProxyCommand != "" && !strings.EqualFold(opts.ProxyCommand, "none") {
		conn, err = dialProxyCommand(opts.ProxyCommand, targetAddr)
	} else if len(opts.ProxyJump) > 0 {
		conn, via, err = dialProxyJumpChain(opts, opts.ProxyJump, targetAddr)
	} else {
		return Dial(opts)
	}

	if err != nil {
		return nil, err
	}

	cfg, agentConn, err := buildSSHClientConfig(opts)
	if err != nil {
		_ = conn.Close()
		closeJumpClients(via)
		return nil, err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, cfg)
	if err != nil {
		abortSSHSetup(agentConn)
		_ = conn.Close()
		closeJumpClients(via)
		return nil, err
	}
	client := ssh.NewClient(ncc, chans, reqs)

	c := &Client{conn: client, agentConn: agentConn, opts: opts, scroll: &terminal.Buffer{}, via: via}
	return c, nil
}

// RunRemoteCommand runs cmdline on the remote host without a PTY, streaming to os.Stdout/Stderr.
// ctx cancellation closes the session and returns ctx.Err(). Non-zero exit returns *RemoteExitError.
// Other errors come from NewSession or sess.Run.
func (c *Client) RunRemoteCommand(ctx context.Context, cmdline string) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	done := make(chan error, 1)
	go func() { done <- sess.Run(cmdline) }()

	select {
	case <-ctx.Done():
		_ = sess.Close()
		return ctx.Err()
	case err := <-done:
		if err == nil {
			return nil
		}
		if ee, ok := err.(*ssh.ExitError); ok {
			return &RemoteExitError{Code: ee.ExitStatus()}
		}
		return err
	}
}

// BuildRemoteCommandLine joins args into a POSIX shell command line with single-quote escaping.
// args may be empty; returns an empty string when len(args) is zero.
func BuildRemoteCommandLine(args []string) string {
	var b strings.Builder
	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		if strings.ContainsAny(a, " \t\"'$\\") {
			b.WriteByte('\'')
			b.WriteString(strings.ReplaceAll(a, "'", `'"'"'`))
			b.WriteByte('\'')
		} else {
			b.WriteString(a)
		}
	}
	return b.String()
}
