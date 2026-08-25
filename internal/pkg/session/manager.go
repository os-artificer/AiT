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

package session

import (
	"time"

	"github.com/os-artificer/AiT/internal/pkg/config"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// LoginOpts for new connection.
type LoginOpts struct {
	Alias string
	Save  bool
	Host  string
	// LoginDestination is the ssh config Host alias or CLI destination before hostname resolution.
	LoginDestination string
	Port             int
	User             string
	IdentityPath     string
	// IdentityExplicit is true when the user passed -i on the CLI.
	IdentityExplicit bool
	Password         string
	Tmux             bool

	RemoteCommand []string

	ProxyJump             []string
	ProxyCommand          string
	KnownHostsPath        string
	StrictHostKeyChecking string
	ConnectTimeout        time.Duration
	BatchMode             bool
}

// Manager orchestrates session lifecycle.
type Manager struct {
	Store  *store.Store
	Config *config.Config
	// MetaCLI runs aissh subcommands from in-session !ait lines (set by internal/ssh).
	MetaCLI func(args []string) error
}

// DialSSH connects with DialChain and prompts for a password when authentication fails and no password was supplied.
// hostLabel is used only in the password prompt (e.g. user@host).
func (m *Manager) DialSSH(sshOpts sshclient.Options, passwordOverride, hostLabel string) (*sshclient.Client, error) {
	return m.dialLogin(sshOpts, passwordOverride, hostLabel)
}

func (m *Manager) dialLogin(sshOpts sshclient.Options, password, host string) (*sshclient.Client, error) {
	client, err := sshclient.DialChain(sshOpts)
	if err == nil {
		return client, nil
	}

	return retryDialAfterPasswordPrompt(sshOpts, err, password, host)
}
