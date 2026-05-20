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

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OpenSSHPath returns the path to the OpenSSH ssh client binary.
func OpenSSHPath() (string, error) {
	if p, err := LookPath("ssh"); err == nil {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		sysRoot := os.Getenv("SystemRoot")
		if sysRoot == "" {
			sysRoot = `C:\Windows`
		}
		fallback := filepath.Join(sysRoot, "System32", "OpenSSH", "ssh.exe")
		if _, err := os.Stat(fallback); err == nil {
			return fallback, nil
		}
	}
	return "", openSSHNotFoundError()
}

func openSSHNotFoundError() error {
	switch runtime.GOOS {
	case "windows":
		return fmt.Errorf("OpenSSH client not found: enable \"OpenSSH Client\" in Settings → Apps → Optional features, or run: winget install Microsoft.OpenSSH.Client")
	case "darwin":
		return fmt.Errorf("OpenSSH client not found: install Xcode command line tools or ensure `ssh` is on PATH (try: which ssh)")
	default:
		return fmt.Errorf("OpenSSH client not found: install openssh-client (e.g. apt install openssh-client)")
	}
}

// SSHConfigDir returns the user's OpenSSH config directory (~/.ssh).
func SSHConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

// ExpandSSHPath expands ~ and Windows %VAR% prefixes in OpenSSH config paths.
func ExpandSSHPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "none" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(p)
	}
	if strings.HasPrefix(p, "~") {
		if len(p) == 1 || p[1] == '/' || p[1] == '\\' {
			rest := strings.TrimPrefix(strings.TrimPrefix(p[1:], "/"), `\`)
			return filepath.Clean(filepath.Join(home, rest))
		}
	}
	if runtime.GOOS == "windows" {
		p = expandWindowsEnv(p)
	}
	return filepath.Clean(p)
}

func expandWindowsEnv(s string) string {
	for _, key := range []string{"USERPROFILE", "HOMEDRIVE", "HOMEPATH", "HOME"} {
		if v := os.Getenv(key); v != "" {
			s = strings.ReplaceAll(s, "%"+key+"%", v)
		}
	}
	return s
}

// OpenSSHInstallHint returns a short hint for installing the ssh client.
func OpenSSHInstallHint() string {
	return openSSHNotFoundError().Error()
}

// ExecOpenSSH runs ssh with the given arguments (config resolution, etc.).
func ExecOpenSSH(args ...string) ([]byte, error) {
	sshPath, err := OpenSSHPath()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(sshPath, args...)
	cmd.Env = os.Environ()
	return cmd.Output()
}
