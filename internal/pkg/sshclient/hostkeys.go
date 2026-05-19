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
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// KnownHostsPath returns the OpenSSH known_hosts file path.
func KnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}

func ensureKnownHostsFile(path string) error {
	if path == "" {
		return fmt.Errorf("cannot resolve home directory for known_hosts")
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, nil, 0o600)
}

func hostKeyCallback() (ssh.HostKeyCallback, error) {
	if os.Getenv("AIT_SSH_INSECURE") == "1" {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // explicit opt-in for dev
	}
	path := KnownHostsPath()
	if err := ensureKnownHostsFile(path); err != nil {
		return nil, fmt.Errorf("prepare %s: %w (set AIT_SSH_INSECURE=1 to skip host key check)", path, err)
	}
	cb, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w (set AIT_SSH_INSECURE=1 to skip host key check)", path, err)
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
			host, _, _ := net.SplitHostPort(remote.String())
			if host == "" {
				host = hostname
			}
			return fmt.Errorf(
				"unknown host key for %s — add it with: ssh-keyscan -H %s >> %q",
				host, host, path,
			)
		}
		return err
	}, nil
}
