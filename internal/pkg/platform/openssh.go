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

const errOpenSSHWindows = "OpenSSH client not found: enable \"OpenSSH Client\" " +
	"in Settings → Apps → Optional features, or run: winget install Microsoft.OpenSSH.Client"

// OpenSSHPath locates the OpenSSH ssh client on PATH, with a Windows System32 fallback.
// Returns the executable path, or an OS-specific install-hint error when ssh is missing.
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
		return fmt.Errorf("%s", errOpenSSHWindows)
	case "darwin":
		return fmt.Errorf("OpenSSH client not found: install Xcode command line tools or ensure `ssh` is on PATH (try: which ssh)")
	default:
		return fmt.Errorf("OpenSSH client not found: install openssh-client (e.g. apt install openssh-client)")
	}
}

// SSHConfigDir returns the user's OpenSSH configuration directory (typically ~/.ssh).
// Returns the path, or os.UserHomeDir error when the home directory cannot be resolved.
func SSHConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

// ExpandSSHPath expands leading ~ and Windows %VAR% tokens in OpenSSH config paths.
// Empty or "none" returns ""; unexpandable paths are filepath.Clean'd only.
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

// OpenSSHInstallHint returns the same install guidance message as OpenSSHPath when ssh is absent.
// The string is suitable for printing to the user; it is not an error value.
func OpenSSHInstallHint() string {
	return openSSHNotFoundError().Error()
}

// ExecOpenSSH runs the OpenSSH ssh binary with args and returns combined stdout/stderr.
// Returns OpenSSHPath errors, or exec exit/IO errors from cmd.Output.
func ExecOpenSSH(args ...string) ([]byte, error) {
	sshPath, err := OpenSSHPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(sshPath, args...)
	cmd.Env = os.Environ()
	return cmd.Output()
}
