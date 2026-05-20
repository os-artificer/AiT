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
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// hostKeyConfirmFunc returns true when the user accepts an unknown host key.
// Tests may replace hostKeyConfirm and isInteractiveTTY.
var hostKeyConfirm = defaultHostKeyConfirm

var isInteractiveTTY = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func defaultHostKeyConfirm(prompt string) bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	fmt.Fprint(os.Stderr, prompt)
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return false
	}
	ans := strings.TrimSpace(strings.ToLower(sc.Text()))
	return ans == "yes" || ans == "y"
}

// KnownHostsPath returns the OpenSSH known_hosts file path.
func KnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}

// parseKnownHostsFiles splits an OpenSSH UserKnownHostsFile value (space-separated paths).
func parseKnownHostsFiles(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	var files []string
	for _, p := range strings.Fields(path) {
		p = strings.TrimSpace(p)
		if p != "" {
			files = append(files, p)
		}
	}
	return files
}

func primaryKnownHostsFile(files []string) string {
	if len(files) == 0 {
		return KnownHostsPath()
	}
	return files[0]
}

func ensureKnownHostsFile(path string) error {
	if path == "" {
		return fmt.Errorf("cannot resolve home directory for known_hosts")
	}
	st, err := os.Stat(path)
	if err == nil {
		if st.IsDir() {
			return fmt.Errorf("%s is a directory, not a file (remove it and retry)", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, nil, 0o600)
}

func hostKeyCallbackFor(opts Options) (ssh.HostKeyCallback, error) {
	if os.Getenv("AIT_SSH_INSECURE") == "1" {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // explicit opt-in for dev
	}
	mode := normalizeStrictHostKeyChecking(opts.StrictHostKeyChecking)
	if mode == "no" {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // StrictHostKeyChecking=no
	}
	path := opts.KnownHostsPath
	if path == "" {
		path = KnownHostsPath()
	}
	return hostKeyCallbackAt(path, opts.Host, opts.Port, mode, opts.BatchMode)
}

func hostKeyCallback() (ssh.HostKeyCallback, error) {
	return hostKeyCallbackAt(KnownHostsPath(), "", 0, "ask", false)
}

func hostKeyCallbackAt(path, host string, port int, mode string, batch bool) (ssh.HostKeyCallback, error) {
	if os.Getenv("AIT_SSH_INSECURE") == "1" {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // explicit opt-in for dev
	}
	files := parseKnownHostsFiles(path)
	if len(files) == 0 {
		files = []string{KnownHostsPath()}
	}
	primary := primaryKnownHostsFile(files)
	if err := ensureKnownHostsFile(primary); err != nil {
		return nil, fmt.Errorf("prepare %s: %w (set AIT_SSH_INSECURE=1 to skip host key check)", primary, err)
	}
	cb, err := openKnownHostsDB(files)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w (set AIT_SSH_INSECURE=1 to skip host key check)", strings.Join(files, " "), err)
	}
	displayHost := host
	if displayHost == "" {
		displayHost = "host"
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}
		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) {
			return err
		}
		if len(keyErr.Want) > 0 {
			h := hostFromAddr(hostname, remote)
			return fmt.Errorf(
				"host key for %s has changed — possible MITM (refusing to connect); remove the old entry from %q",
				h, primary,
			)
		}
		// Unknown host key.
		h := hostFromAddr(hostname, remote)
		if h != "" && displayHost == "host" {
			displayHost = h
		}
		p := port
		if _, portStr, splitErr := net.SplitHostPort(remote.String()); splitErr == nil {
			if pp, perr := parsePortString(portStr); perr == nil && pp != 0 {
				p = pp
			}
		}
		if p == 0 {
			p = port
		}
		if p == 0 {
			p = 22
		}

		effective := mode
		if mode == "ask" && (batch || !isInteractiveTTY()) {
			effective = "yes"
		}

		switch effective {
		case "accept-new":
			return appendKnownHost(primary, displayHost, p, key)
		case "ask":
			fp := ssh.FingerprintSHA256(key)
			keyType := key.Type()
			portNote := ""
			if p != 22 {
				portNote = fmt.Sprintf(" (port %d)", p)
			}
			prompt := fmt.Sprintf(
				"The authenticity of host '%s'%s can't be established.\n"+
					"%s key fingerprint is %s.\n"+
					"Are you sure you want to continue connecting (yes/no)? ",
				displayHost, portNote, strings.ToUpper(keyType), fp,
			)
			if !hostKeyConfirm(prompt) {
				return fmt.Errorf("host key verification failed")
			}
			return appendKnownHost(primary, displayHost, p, key)
		default: // yes
			return hostKeyUnknownError(displayHost, p, primary)
		}
	}, nil
}

func hostFromAddr(hostname string, remote net.Addr) string {
	host, _, _ := net.SplitHostPort(remote.String())
	if host == "" {
		return hostname
	}
	return host
}

func hostKeyUnknownError(host string, port int, path string) error {
	if port != 0 && port != 22 {
		return fmt.Errorf(
			"unknown host key for %s — add it with: ssh-keyscan -p %d -H %s >> %q",
			host, port, host, path,
		)
	}
	return fmt.Errorf(
		"unknown host key for %s — add it with: ssh-keyscan -H %s >> %q",
		host, host, path,
	)
}

func knownHostsAddresses(host string, port int) []string {
	if port == 0 || port == 22 {
		return []string{host}
	}
	return []string{fmt.Sprintf("[%s]:%d", host, port)}
}

func openKnownHostsDB(files []string) (ssh.HostKeyCallback, error) {
	var existing []string
	for _, fn := range files {
		if _, err := os.Stat(fn); err == nil {
			existing = append(existing, fn)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if len(existing) == 0 {
		existing = []string{primaryKnownHostsFile(files)}
	}
	return knownhosts.New(existing...)
}

func appendKnownHost(path, host string, port int, key ssh.PublicKey) error {
	line := knownhosts.Line(knownHostsAddresses(host, port), key) + "\n"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("append to %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func normalizeStrictHostKeyChecking(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "ask":
		return "ask"
	case "yes", "true", "on":
		return "yes"
	case "no", "false", "off":
		return "no"
	case "accept-new":
		return "accept-new"
	default:
		return "ask"
	}
}
