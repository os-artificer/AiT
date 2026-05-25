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

// Package dial holds receiver-free helpers shared by the SSH dialer in the
// parent ssh/client package: jump-host parsing, ProxyCommand stdio wiring,
// and SSH private-key loading. The Client construction and HostKeyCallback
// setup stay in ssh/client because they need access to Client internals and
// to the package-private test hooks (hostKeyConfirm, isInteractiveTTY).
//
// Strategy(dial/proxy): ProxyJump vs ProxyCommand share the StdioConn shape
// (net.Conn over a child process's stdio); the parent package's DialChain
// selects between them. Build-tagged proxy_unix.go / proxy_windows.go are
// platform variants, not multiple strategies.
package dial

import (
	"io"
	"net"
	"os/exec"
	"time"
)

// StdioConn is a net.Conn that wraps a child process's stdio (used by ProxyCommand).
type StdioConn struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

// Read implements net.Conn; reads from the ProxyCommand process stdout.
func (c *StdioConn) Read(b []byte) (int, error) { return c.stdout.Read(b) }

// Write implements net.Conn; writes to the ProxyCommand process stdin.
func (c *StdioConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }

// Close implements net.Conn; closes stdio pipes and waits on the proxy command.
func (c *StdioConn) Close() error {
	_ = c.stdin.Close()
	_ = c.stdout.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return c.cmd.Wait()
}

// LocalAddr implements net.Conn; always nil for proxy stdio (no socket address).
func (c *StdioConn) LocalAddr() net.Addr { return nil }

// RemoteAddr implements net.Conn; always nil for proxy stdio (no socket address).
func (c *StdioConn) RemoteAddr() net.Addr { return nil }

// SetDeadline implements net.Conn; no-op on proxy stdio.
func (c *StdioConn) SetDeadline(time.Time) error { return nil }

// SetReadDeadline implements net.Conn; no-op on proxy stdio.
func (c *StdioConn) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline implements net.Conn; no-op on proxy stdio.
func (c *StdioConn) SetWriteDeadline(time.Time) error { return nil }
