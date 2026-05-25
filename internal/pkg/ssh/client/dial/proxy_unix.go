//go:build !windows

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

package dial

import (
	"fmt"
	"net"
	"os/exec"
)

// ProxyCommand runs proxyCommand under /bin/sh -c and returns a StdioConn wrapping
// its stdin/stdout. The targetAddr argument is accepted for parity with the
// Windows variant but ignored on POSIX (the proxy command is responsible for the
// upstream connection).
func ProxyCommand(proxyCommand, _ string) (net.Conn, error) {
	cmd := exec.Command("/bin/sh", "-c", proxyCommand)
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
	return &StdioConn{stdin: stdin, stdout: stdout, cmd: cmd}, nil
}
