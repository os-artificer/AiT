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
	"net"
	"path/filepath"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Server is a loopback SSH+SFTP daemon for automated acceptance tests.
type Server struct {
	cfg          Config
	ln           net.Listener
	hostSigner   ssh.Signer
	clientSigner ssh.Signer
	identityPath string
	done         chan struct{}
	closeOnce    sync.Once
}

// Close stops the accept loop. It is safe to call more than once.
func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.ln.Close()
		<-s.done
	})
	return err
}

// Addr returns the bound host:port string.
func (s *Server) Addr() string {
	return s.ln.Addr().String()
}

// Host returns the listen hostname (always 127.0.0.1 when using DefaultAddr).
func (s *Server) Host() string {
	host, _, err := net.SplitHostPort(s.Addr())
	if err != nil {
		return "127.0.0.1"
	}
	return host
}

// Port returns the ephemeral listen port.
func (s *Server) Port() int {
	_, port, err := net.SplitHostPort(s.Addr())
	if err != nil {
		return 0
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return 0
	}
	return p
}

// User returns the configured username.
func (s *Server) User() string {
	return s.cfg.User
}

// Password returns the configured password (empty when RejectPassword is set).
func (s *Server) Password() string {
	if s.cfg.RejectPassword {
		return ""
	}
	return s.cfg.Password
}

// RootDir returns the mock filesystem root.
func (s *Server) RootDir() string {
	return s.cfg.RootDir
}

// IdentityPath returns the generated client private-key path, or empty.
func (s *Server) IdentityPath() string {
	return s.identityPath
}

// RemotePath joins name under the mock root using the OS path separator.
func (s *Server) RemotePath(name string) string {
	return filepath.Join(s.cfg.RootDir, name)
}

// Info returns fixture metadata for pytest / cmd/testenv-mock.
func (s *Server) Info() Info {
	return Info{
		Host:         s.Host(),
		Port:         s.Port(),
		User:         s.User(),
		Password:     s.Password(),
		Root:         s.RootDir(),
		IdentityPath: s.IdentityPath(),
		Welcome:      s.RemotePath(WelcomeName),
	}
}
