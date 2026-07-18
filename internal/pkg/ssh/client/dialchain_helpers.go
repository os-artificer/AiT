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

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/dial"
)

// dialProxyJumpChain stays in the client package because it recursively calls
// DialChain, which depends on private Client construction.
//
// The returned via slice must stay open for as long as conn is used; the
// caller attaches it to the final Client and closes hops in Client.Close.
func dialProxyJumpChain(targetOpts Options, jumps []string, targetAddr string) (net.Conn, []*Client, error) {
	if len(jumps) == 0 {
		return nil, nil, fmt.Errorf("empty proxy jump chain")
	}
	jumpDest := jumps[0]

	jumpOpts, err := jumpOptionsForHop(jumpDest, targetOpts)
	if err != nil {
		return nil, nil, err
	}
	jumpOpts.ProxyJump = jumps[1:]
	jumpOpts.ProxyCommand = ""

	jumpClient, err := DialChain(jumpOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("proxy jump %s: %w", jumpDest, err)
	}

	conn, err := jumpClient.conn.Dial("tcp", targetAddr)
	if err != nil {
		_ = jumpClient.Close()
		return nil, nil, fmt.Errorf("forward to %s via %s: %w", targetAddr, jumpDest, err)
	}
	return conn, []*Client{jumpClient}, nil
}

func jumpOptionsForHop(jumpDest string, targetOpts Options) (Options, error) {
	if JumpDialResolver != nil {
		opts, err := JumpDialResolver(jumpDest, targetOpts)
		if err != nil {
			return Options{}, err
		}
		return opts, nil
	}

	jumpUser, jumpHost, jumpPort, err := dial.ParseJumpDestination(jumpDest)
	if err != nil {
		return Options{}, err
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
	return jumpOpts, nil
}

func parsePortString(s string) (int, error) {
	return dial.ParsePortString(s)
}

func dialProxyCommand(proxyCommand, targetAddr string) (net.Conn, error) {
	return dial.ProxyCommand(proxyCommand, targetAddr)
}

func closeJumpClients(via []*Client) {
	for i := len(via) - 1; i >= 0; i-- {
		_ = via[i].Close()
	}
}
