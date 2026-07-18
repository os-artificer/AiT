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

package resolve

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/dial"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/stricthostkey"
)

// sshGOutput runs `ssh -G` for destination and returns stdout.
// Tests may replace this to avoid a live OpenSSH dependency.
var sshGOutput = defaultSSHGOutput

func defaultSSHGOutput(sshPath string, args []string) ([]byte, error) {
	cmd := exec.Command(sshPath, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("ssh -G: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("ssh -G: %w", err)
	}
	return out, nil
}

func init() {
	client.JumpDialResolver = ResolveJumpDialOptions
}

// expandProxyJumpList resolves each ProxyJump hop through ssh -G so Host aliases
// (e.g. "devhost") become user@hostname:port before DialChain (PJ2).
func expandProxyJumpList(sshPath string, spec openssh.Spec, jumps []string) ([]string, error) {
	if len(jumps) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(jumps))
	for _, j := range jumps {
		expanded, err := expandOneProxyJump(sshPath, spec, j)
		if err != nil {
			return nil, err
		}
		out = append(out, expanded)
	}
	return out, nil
}

func expandOneProxyJump(sshPath string, base openssh.Spec, jump string) (string, error) {
	opts, err := resolveJumpHop(sshPath, base, jump)
	if err != nil {
		return "", err
	}
	return formatJumpDestination(opts.User, opts.Host, opts.Port), nil
}

// ResolveJumpDialOptions resolves a single ProxyJump hop via ssh -G into dial Options.
// Remaining ProxyJump hops must be set by the caller. Used by DialChain (PJ2).
func ResolveJumpDialOptions(jump string, target client.Options) (client.Options, error) {
	sshPath, err := platform.OpenSSHPath()
	if err != nil {
		return client.Options{}, err
	}
	base := openssh.Spec{}
	opts, err := resolveJumpHop(sshPath, base, jump)
	if err != nil {
		return client.Options{}, err
	}

	out := target
	out.Host = opts.Host
	out.Port = opts.Port
	out.User = opts.User
	if opts.IdentityPath != "" {
		out.IdentityPath = opts.IdentityPath
		out.IdentityExplicit = opts.IdentityExplicit
	}
	out.ProxyCommand = ""
	return out, nil
}

func resolveJumpHop(sshPath string, base openssh.Spec, jump string) (client.Options, error) {
	jumpUser, jumpHost, jumpPort, err := dial.ParseJumpDestination(jump)
	if err != nil {
		return client.Options{}, err
	}
	if jumpHost == "" {
		return client.Options{}, fmt.Errorf("empty proxy jump host in %q", jump)
	}

	args := []string{"-G"}
	if base.ConfigFile != "" {
		args = append(args, "-F", base.ConfigFile)
	}
	if jumpPort != 0 {
		args = append(args, "-p", strconv.Itoa(jumpPort))
	}
	if jumpUser != "" {
		args = append(args, "-l", jumpUser)
	}
	args = append(args, jumpHost)

	out, err := sshGOutput(sshPath, args)
	if err != nil {
		return client.Options{}, fmt.Errorf("proxy jump %s: %w", jump, err)
	}
	cfg := parseSSHGOutput(out)

	host := firstNonEmpty(cfg.single["hostname"], jumpHost)
	port := parsePort(cfg.single["port"])
	if jumpPort != 0 {
		port = jumpPort
	}
	if port == 0 {
		port = 22
	}
	user := cfg.single["user"]
	if jumpUser != "" {
		user = jumpUser
	}

	identity := firstExistingIdentity(expandIdentityPaths(cfg.multi["identityfile"]))
	opts := client.Options{
		Host:             host,
		Port:             port,
		User:             user,
		IdentityPath:     identity,
		IdentityExplicit: false,
		UseAgent:         true,
		KnownHostsPath:   platform.ExpandSSHPath(cfg.single["userknownhostsfile"]),
		StrictHostKeyChecking: stricthostkey.NormalizeStrictHostKeyChecking(
			cfg.single["stricthostkeychecking"],
		),
		BatchMode: strings.EqualFold(cfg.single["batchmode"], "yes"),
	}
	if opts.Host == "" {
		return client.Options{}, fmt.Errorf("proxy jump %s: could not resolve hostname", jump)
	}
	if opts.IdentityPath != "" {
		if _, err := os.Stat(opts.IdentityPath); err != nil {
			opts.IdentityPath = ""
		}
	}
	return opts, nil
}

func formatJumpDestination(user, host string, port int) string {
	if port == 0 {
		port = 22
	}
	if user != "" {
		return fmt.Sprintf("%s@%s:%d", user, host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
