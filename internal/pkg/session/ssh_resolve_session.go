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
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// loginSpecFromSession builds an openssh.Spec for ssh -G re-resolve on load/switch.
func loginSpecFromSession(sess *store.Session, password string) openssh.Spec {
	spec := openssh.Spec{
		Destination: sessionDialDestination(sess),
		User:        sess.Username,
		Port:        sess.Port,
		Identity:    sess.IdentityPath,
		Password:    password,
		ProxyJump:   parseProxyJump(sess.ProxyJump),
	}
	return spec
}

// resolveSSHOptionsFromSession re-runs ssh -G for the saved session and returns dial options.
// Stored ProxyJump and ProxyCommand override empty resolve results when set at login.
func (m *Manager) resolveSSHOptionsFromSession(sess *store.Session, password string) (sshclient.Options, error) {
	spec := loginSpecFromSession(sess, password)

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return sshclient.Options{}, err
	}

	opts := resolved.ClientOptions()
	if len(parseProxyJump(sess.ProxyJump)) > 0 {
		opts.ProxyJump = parseProxyJump(sess.ProxyJump)
	}
	if sess.ProxyCommand != "" {
		opts.ProxyCommand = sess.ProxyCommand
	}
	opts.Password = password
	return opts, nil
}
