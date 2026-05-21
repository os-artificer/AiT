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
	"fmt"
	"net"
	"strings"
)

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
