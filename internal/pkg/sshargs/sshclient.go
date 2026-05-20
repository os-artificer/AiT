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

package sshargs

import "github.com/os-artificer/AiT/internal/pkg/sshclient"

// SSHClientOptions maps Resolved to sshclient.Options.
func (r Resolved) SSHClientOptions() sshclient.Options {
	return sshclient.Options{
		Host:                  r.Host,
		Port:                  r.Port,
		User:                  r.User,
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
