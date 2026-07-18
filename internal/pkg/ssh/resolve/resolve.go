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
	"strconv"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/stricthostkey"
)

// Resolved is the effective SSH connection parameters after ssh -G and CLI overrides.
type Resolved struct {
	Host         string
	Port         int
	User         string
	IdentityPath string
	// IdentityExplicit is true when the user passed -i on the CLI.
	IdentityExplicit bool
	Password         string

	ProxyJump    []string
	ProxyCommand string

	KnownHostsPath        string
	StrictHostKeyChecking string // yes, no, ask, accept-new
	ConnectTimeout        time.Duration
	BatchMode             bool

	RemoteCommand []string
	Alias         string
	// Destination is the CLI/OpenSSH destination before hostname resolution (e.g. devhost).
	Destination string
	Save        bool
	Tmux        bool
}

// Resolve runs ssh -G for spec.Destination and merges the parsed config with CLI flags in spec.
// Returns effective connection parameters for dial. Errors: platform.OpenSSHPath; ssh -G failure;
// identity file not found; could not resolve hostname; cannot combine --tmux with remote command.
func Resolve(spec openssh.Spec) (Resolved, error) {
	sshPath, err := platform.OpenSSHPath()
	if err != nil {
		return Resolved{}, err
	}

	args := []string{"-G"}
	if spec.ConfigFile != "" {
		args = append(args, "-F", spec.ConfigFile)
	}
	for _, j := range spec.ProxyJump {
		args = append(args, "-J", j)
	}
	if spec.Port != 0 {
		args = append(args, "-p", strconv.Itoa(spec.Port))
	}
	if spec.User != "" {
		args = append(args, "-l", spec.User)
	}
	if spec.Identity != "" {
		args = append(args, "-i", spec.Identity)
	}
	for k, v := range spec.Options {
		args = append(args, "-o", k+"="+v)
	}
	args = append(args, spec.Destination)

	out, err := sshGOutput(sshPath, args)
	if err != nil {
		return Resolved{}, err
	}

	cfg := parseSSHGOutput(out)

	r := Resolved{
		Host:                  firstNonEmpty(cfg.single["hostname"], spec.Destination),
		Port:                  parsePort(cfg.single["port"]),
		User:                  cfg.single["user"],
		ProxyCommand:          cfg.single["proxycommand"],
		KnownHostsPath:        platform.ExpandSSHPath(cfg.single["userknownhostsfile"]),
		StrictHostKeyChecking: stricthostkey.NormalizeStrictHostKeyChecking(cfg.single["stricthostkeychecking"]),
		BatchMode:             strings.EqualFold(cfg.single["batchmode"], "yes"),
		RemoteCommand:         spec.RemoteCommand,
		Alias:                 spec.Alias,
		Destination:           spec.Destination,
		Save:                  spec.Save,
		Tmux:                  spec.Tmux,
		Password:              spec.Password,
		IdentityExplicit:      spec.Identity != "",
	}

	if r.Port == 0 {
		r.Port = 22
	}
	if spec.Port != 0 {
		r.Port = spec.Port
	}
	if spec.User != "" {
		r.User = spec.User
	}

	identityFiles := expandIdentityPaths(cfg.multi["identityfile"])
	if spec.Identity != "" {
		r.IdentityPath = platform.ExpandSSHPath(spec.Identity)
		r.IdentityExplicit = true
		if _, err := os.Stat(r.IdentityPath); err != nil {
			return Resolved{}, fmt.Errorf("identity file not found: %s", r.IdentityPath)
		}
	} else if spec.Options["identityfile"] != "" {
		r.IdentityPath = platform.ExpandSSHPath(spec.Options["identityfile"])
		if _, err := os.Stat(r.IdentityPath); err != nil {
			return Resolved{}, fmt.Errorf("identity file not found: %s", r.IdentityPath)
		}
	} else {
		r.IdentityPath = firstExistingIdentity(identityFiles)
	}

	if spec.ProxyJump != nil {
		r.ProxyJump = append([]string(nil), spec.ProxyJump...)
	} else if pj := cfg.single["proxyjump"]; pj != "" && !strings.EqualFold(pj, "none") {
		for _, j := range strings.Split(pj, ",") {
			j = strings.TrimSpace(j)
			if j != "" {
				r.ProxyJump = append(r.ProxyJump, j)
			}
		}
	}

	if len(r.ProxyJump) > 0 {
		expanded, err := expandProxyJumpList(sshPath, spec, r.ProxyJump)
		if err != nil {
			return Resolved{}, err
		}
		r.ProxyJump = expanded
	}

	if pc := cfg.single["proxycommand"]; pc != "" && !strings.EqualFold(pc, "none") {
		r.ProxyCommand = pc
	}

	if spec.Options["port"] != "" {
		if p, err := strconv.Atoi(spec.Options["port"]); err == nil {
			r.Port = p
		}
	}
	if spec.Options["user"] != "" {
		r.User = spec.Options["user"]
	}
	if kh := spec.Options["userknownhostsfile"]; kh != "" {
		r.KnownHostsPath = platform.ExpandSSHPath(kh)
	}
	if spec.Options["stricthostkeychecking"] != "" {
		r.StrictHostKeyChecking = stricthostkey.NormalizeStrictHostKeyChecking(spec.Options["stricthostkeychecking"])
	}
	if to := cfg.single["connecttimeout"]; to != "" && to != "none" && to != "0" {
		if sec, err := strconv.Atoi(to); err == nil && sec > 0 {
			r.ConnectTimeout = time.Duration(sec) * time.Second
		}
	}

	// Parse destination for embedded user@host if hostname still looks like alias
	if u, h, p, err := openssh.ParseDestination(spec.Destination); err == nil && h != "" {
		if r.Host == spec.Destination || r.Host == "" {
			r.Host = h
		}
		if u != "" && r.User == "" {
			r.User = u
		}
		if p != 0 && spec.Port == 0 {
			r.Port = p
		}
	}

	if r.Host == "" {
		return Resolved{}, fmt.Errorf("could not resolve hostname from %q", spec.Destination)
	}

	if len(spec.RemoteCommand) > 0 && spec.Tmux {
		return Resolved{}, openssh.TmuxWithRemoteCommandErr()
	}

	return r, nil
}

// ClientOptions maps r to client.Options for dialing (UseAgent is always true).
// Returns a value copy suitable for client.Dial or DialChain.
func (r Resolved) ClientOptions() client.Options {
	return client.Options{
		Host:                  r.Host,
		Port:                  r.Port,
		User:                  r.User,
		LoginDestination:      r.Destination,
		IdentityPath:          r.IdentityPath,
		IdentityExplicit:      r.IdentityExplicit,
		Password:              r.Password,
		UseAgent:              true,
		ProxyJump:             append([]string(nil), r.ProxyJump...),
		ProxyCommand:          r.ProxyCommand,
		KnownHostsPath:        r.KnownHostsPath,
		StrictHostKeyChecking: r.StrictHostKeyChecking,
		ConnectTimeout:        r.ConnectTimeout,
		BatchMode:             r.BatchMode,
	}
}

// ToSSHOptions returns the Resolved fields as separate return values for legacy call sites.
// The returned slices and strings are copies or primitives from r; this method never errors.
func (r Resolved) ToSSHOptions() (
	host string,
	port int,
	user, identity, password string,
	proxyJump []string,
	proxyCommand string,
	knownHosts string,
	timeout time.Duration,
	batch bool,
) {
	return r.Host, r.Port, r.User, r.IdentityPath, r.Password,
		append([]string(nil), r.ProxyJump...), r.ProxyCommand,
		r.KnownHostsPath, r.ConnectTimeout, r.BatchMode
}
