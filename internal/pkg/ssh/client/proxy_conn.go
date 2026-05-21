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
	"io"
	"net"
	"os/exec"
	"time"
)

type proxyStdioConn struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

// Read implements net.Conn; reads from the ProxyCommand process stdout.
// Returns the read count and any io error from stdout.Read.
func (c *proxyStdioConn) Read(b []byte) (int, error) { return c.stdout.Read(b) }

// Write implements net.Conn; writes to the ProxyCommand process stdin.
// Returns the write count and any io error from stdin.Write.
func (c *proxyStdioConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }

// Close implements net.Conn; closes stdio pipes and waits on the proxy command.
// Returns cmd.Wait error after killing the process when needed.
func (c *proxyStdioConn) Close() error {
	_ = c.stdin.Close()
	_ = c.stdout.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return c.cmd.Wait()
}

// LocalAddr implements net.Conn; always nil for proxy stdio (no socket address).
func (c *proxyStdioConn) LocalAddr() net.Addr { return nil }

// RemoteAddr implements net.Conn; always nil for proxy stdio (no socket address).
func (c *proxyStdioConn) RemoteAddr() net.Addr { return nil }

// SetDeadline implements net.Conn; no-op on proxy stdio (always returns nil).
func (c *proxyStdioConn) SetDeadline(time.Time) error { return nil }

// SetReadDeadline implements net.Conn; no-op on proxy stdio (always returns nil).
func (c *proxyStdioConn) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline implements net.Conn; no-op on proxy stdio (always returns nil).
func (c *proxyStdioConn) SetWriteDeadline(time.Time) error { return nil }
