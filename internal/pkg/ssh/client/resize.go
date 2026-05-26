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

package client

import "golang.org/x/crypto/ssh"

type sessionResize struct {
	sess   *ssh.Session
	client *Client
}

// OnResize implements pty.ResizeHandler; forwards new dimensions to the local
// shell (when active) and the SSH session window.
func (r *sessionResize) OnResize(cols, rows int) {
	if r.client.localShell.Active() {
		_ = r.client.localShell.Resize(cols, rows)
	}

	if cols > 0 && rows > 0 {
		_ = r.sess.WindowChange(rows, cols)
	}
}

func newSessionResize(sess *ssh.Session, c *Client) *sessionResize {
	return &sessionResize{sess: sess, client: c}
}
