//go:build windows

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
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"
)

func dialProxyCommand(proxyCommand, _ string) (net.Conn, error) {
	cmd := exec.Command("cmd.exe", "/c", proxyCommand)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("proxycommand: %w", err)
	}
	return &proxyStdioConn{stdin: stdin, stdout: stdout, cmd: cmd}, nil
}

type proxyStdioConn struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (c *proxyStdioConn) Read(b []byte) (int, error)  { return c.stdout.Read(b) }
func (c *proxyStdioConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }
func (c *proxyStdioConn) Close() error {
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return c.cmd.Wait()
}
func (c *proxyStdioConn) LocalAddr() net.Addr              { return nil }
func (c *proxyStdioConn) RemoteAddr() net.Addr             { return nil }
func (c *proxyStdioConn) SetDeadline(time.Time) error      { return nil }
func (c *proxyStdioConn) SetReadDeadline(time.Time) error  { return nil }
func (c *proxyStdioConn) SetWriteDeadline(time.Time) error { return nil }
