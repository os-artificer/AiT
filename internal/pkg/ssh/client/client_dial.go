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
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/dial"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/hostkeys"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Dial opens a direct TCP SSH connection to opts.Host:opts.Port.
// opts.User defaults to the current OS user; opts.Port defaults to 22.
// Returns a Client, or errors from user lookup, buildAuth, host key setup, or ssh.Dial.
func Dial(opts Options) (*Client, error) {
	if err := normalizeDialOptions(&opts); err != nil {
		return nil, err
	}

	cfg, agentConn, err := buildSSHClientConfig(opts)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	conn, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		abortSSHSetup(agentConn)
		return nil, err
	}
	c := &Client{conn: conn, agentConn: agentConn, opts: opts, scroll: &terminal.Buffer{}}
	return c, nil
}

// DetectAuthMethod returns a store.AuthMethod* label inferred from opts (password, key, or agent).
// opts.Password and opts.IdentityPath determine the label; empty both yields AuthMethodAgent.
func DetectAuthMethod(opts Options) string {
	if opts.Password != "" {
		return store.AuthMethodPassword
	}
	if opts.IdentityPath != "" {
		return store.AuthMethodKey
	}
	return store.AuthMethodAgent
}

func buildAuth(opts Options) ([]ssh.AuthMethod, net.Conn, error) {
	var methods []ssh.AuthMethod
	var agentConn net.Conn

	if opts.Password != "" {
		methods = append(methods, ssh.Password(opts.Password))
	}

	tryDefaultKeys := func() {
		home, _ := os.UserHomeDir()
		for _, name := range []string{"id_ed25519", "id_rsa"} {
			p := filepath.Join(home, ".ssh", name)
			if opts.IdentityPath != "" && p == opts.IdentityPath {
				continue
			}
			if _, err := os.Stat(p); err == nil {
				if signer, err := dial.LoadSigner(p, opts.Password); err == nil {
					methods = append(methods, ssh.PublicKeys(signer))
					return
				}
			}
		}
	}

	if opts.IdentityPath != "" {
		if _, err := os.Stat(opts.IdentityPath); err != nil {
			if opts.IdentityExplicit {
				return nil, agentConn, fmt.Errorf("identity file not found: %s", opts.IdentityPath)
			}
			tryDefaultKeys()
		} else if signer, err := dial.LoadSigner(opts.IdentityPath, opts.Password); err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		} else if opts.IdentityExplicit {
			return nil, agentConn, err
		} else {
			tryDefaultKeys()
		}
	} else {
		tryDefaultKeys()
	}

	if opts.UseAgent {
		if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
			conn, err := net.Dial("unix", sock)
			if err == nil {
				agentConn = conn
				ag := agent.NewClient(conn)
				methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
			}
		}
	}

	if len(methods) == 0 {
		return nil, agentConn, fmt.Errorf("no authentication method available")
	}
	return methods, agentConn, nil
}

// loadSigner is a small wrapper around dial.LoadSigner kept for tests that
// reference the unexported symbol.
func loadSigner(path, passphrase string) (ssh.Signer, error) {
	return dial.LoadSigner(path, passphrase)
}

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

	hkc, err := hostkeys.Callback(hostkeys.Options{
		Host:                  opts.Host,
		Port:                  opts.Port,
		KnownHostsPath:        opts.KnownHostsPath,
		StrictHostKeyChecking: opts.StrictHostKeyChecking,
		BatchMode:             opts.BatchMode,
	})
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
