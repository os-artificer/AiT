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

package testssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Start listens on an ephemeral loopback port, seeds RootDir, and registers t.Cleanup.
func Start(t *testing.T, cfg Config) *Server {
	t.Helper()

	if cfg.RootDir == "" {
		cfg.RootDir = t.TempDir()
	}
	if cfg.IdentityDir == "" {
		cfg.IdentityDir = t.TempDir()
	}
	if err := seedRoot(cfg.RootDir); err != nil {
		t.Fatal(err)
	}

	srv, err := Listen(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

// Listen starts the mock sshd. The caller must Close the returned Server.
func Listen(cfg Config) (*Server, error) {
	cfg = applyConfigDefaults(cfg)
	if cfg.RootDir == "" {
		return nil, fmt.Errorf("RootDir is required")
	}
	if err := seedRoot(cfg.RootDir); err != nil {
		return nil, err
	}

	hostSigner, err := generateSigner()
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}

	clientSigner, identityPath, err := maybeWriteIdentity(cfg.IdentityDir)
	if err != nil {
		return nil, err
	}

	addr := cfg.Addr
	if addr == "" {
		addr = DefaultAddr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	if cfg.promptListenHost == "" {
		if host, _, splitErr := net.SplitHostPort(ln.Addr().String()); splitErr == nil {
			cfg.promptListenHost = host
		}
	}

	srv := &Server{
		cfg:          cfg,
		ln:           ln,
		hostSigner:   hostSigner,
		clientSigner: clientSigner,
		identityPath: identityPath,
		done:         make(chan struct{}),
	}
	go acceptLoop(srv)
	return srv, nil
}

func applyConfigDefaults(cfg Config) Config {
	if cfg.User == "" {
		cfg.User = DefaultUser
	}
	if cfg.Password == "" {
		cfg.Password = DefaultPassword
	}
	if cfg.Hostname == "" {
		cfg.Hostname = DefaultHostname
	}
	return cfg
}

func maybeWriteIdentity(dir string) (signer ssh.Signer, path string, err error) {
	if dir == "" {
		return nil, "", nil
	}
	signer, pemBytes, err := generateIdentityPEM()
	if err != nil {
		return nil, "", err
	}
	path = filepath.Join(dir, identityFileName)
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		return nil, "", err
	}
	return signer, path, nil
}
