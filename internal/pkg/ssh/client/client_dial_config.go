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
	"net"
	"os/user"
	"time"

	"golang.org/x/crypto/ssh"
)

func normalizeDialOptions(opts *Options) error {
	if opts.User == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		opts.User = u.Username
	}

	if opts.Port == 0 {
		opts.Port = 22
	}
	return nil
}

func buildSSHClientConfig(opts Options) (*ssh.ClientConfig, net.Conn, error) {
	auth, agentConn, err := buildAuth(opts)
	if err != nil {
		return nil, agentConn, err
	}

	hkc, err := hostKeyCallbackFor(opts)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, nil, err
	}

	timeout := opts.ConnectTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	return &ssh.ClientConfig{
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: hkc,
		Timeout:         timeout,
	}, agentConn, nil
}

func abortSSHSetup(agentConn net.Conn) {
	if agentConn != nil {
		_ = agentConn.Close()
	}
}
