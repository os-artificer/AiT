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
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"golang.org/x/crypto/ssh"
)

// DialChain connects to the target host, optionally via ProxyJump or ProxyCommand.
func DialChain(opts Options) (*Client, error) {
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

	if len(opts.ProxyJump) == 0 && opts.ProxyCommand == "" {
		return Dial(opts)
	}

	targetAddr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	var conn net.Conn
	var err error

	if opts.ProxyCommand != "" && !strings.EqualFold(opts.ProxyCommand, "none") {
		conn, err = dialProxyCommand(opts.ProxyCommand, targetAddr)
	} else if len(opts.ProxyJump) > 0 {
		conn, err = dialProxyJumpChain(opts, opts.ProxyJump, targetAddr)
	} else {
		return Dial(opts)
	}
	if err != nil {
		return nil, err
	}

	auth, agentConn, err := buildAuth(opts)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	hkc, err := hostKeyCallbackFor(opts)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		_ = conn.Close()
		return nil, err
	}
	timeout := opts.ConnectTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	cfg := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: hkc,
		Timeout:         timeout,
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, cfg)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		_ = conn.Close()
		return nil, err
	}
	client := ssh.NewClient(ncc, chans, reqs)
	return &Client{conn: client, agentConn: agentConn, opts: opts, scroll: &terminal.Buffer{}}, nil
}

func dialProxyJumpChain(targetOpts Options, jumps []string, targetAddr string) (net.Conn, error) {
	if len(jumps) == 0 {
		return nil, fmt.Errorf("empty proxy jump chain")
	}
	jumpDest := jumps[0]
	jumpUser, jumpHost, jumpPort, err := parseJumpDestination(jumpDest)
	if err != nil {
		return nil, err
	}
	jumpOpts := targetOpts
	jumpOpts.Host = jumpHost
	jumpOpts.Port = jumpPort
	if jumpPort == 0 {
		jumpOpts.Port = 22
	}
	if jumpUser != "" {
		jumpOpts.User = jumpUser
	}
	jumpOpts.ProxyJump = jumps[1:]
	jumpOpts.ProxyCommand = ""

	jumpClient, err := DialChain(jumpOpts)
	if err != nil {
		return nil, fmt.Errorf("proxy jump %s: %w", jumpDest, err)
	}
	defer jumpClient.Close()

	conn, err := jumpClient.conn.Dial("tcp", targetAddr)
	if err != nil {
		return nil, fmt.Errorf("forward to %s via %s: %w", targetAddr, jumpDest, err)
	}
	return conn, nil
}

func parseJumpDestination(dest string) (user, host string, port int, err error) {
	dest = strings.TrimSpace(dest)
	if at := strings.LastIndex(dest, "@"); at >= 0 {
		user = dest[:at]
		dest = dest[at+1:]
	}
	if strings.HasPrefix(dest, "[") {
		end := strings.Index(dest, "]")
		if end < 0 {
			return "", "", 0, fmt.Errorf("invalid jump host %q", dest)
		}
		host = dest[1:end]
		rest := dest[end+1:]
		if strings.HasPrefix(rest, ":") {
			fmt.Sscanf(rest[1:], "%d", &port)
		}
		return user, host, port, nil
	}
	if i := strings.LastIndex(dest, ":"); i > 0 {
		if p, perr := parsePortString(dest[i+1:]); perr == nil {
			host = dest[:i]
			port = p
			return user, host, port, nil
		}
	}
	host = dest
	return user, host, 0, nil
}

func parsePortString(s string) (int, error) {
	var p int
	_, err := fmt.Sscanf(s, "%d", &p)
	return p, err
}

// RemoteExitError is returned when a remote command exits non-zero.
type RemoteExitError struct {
	Code int
}

func (e *RemoteExitError) Error() string {
	return fmt.Sprintf("remote command exited with status %d", e.Code)
}

// RunRemoteCommand runs a remote command without a PTY, streaming to os stdout/stderr.
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

// BuildRemoteCommandLine joins command arguments for the remote shell.
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
