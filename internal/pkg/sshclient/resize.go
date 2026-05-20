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
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// resizeHandler receives terminal dimension changes during an interactive session.
type resizeHandler interface {
	onResize(cols, rows int)
}

type sessionResize struct {
	sess   *ssh.Session
	client *Client
}

func (r *sessionResize) onResize(cols, rows int) {
	if r.client.localShell.active() {
		_ = r.client.localShell.resize(cols, rows)
	}
	if cols > 0 && rows > 0 {
		_ = r.sess.WindowChange(rows, cols)
	}
}

func newSessionResize(sess *ssh.Session, c *Client) *sessionResize {
	return &sessionResize{sess: sess, client: c}
}

func termSize(fd int) (cols, rows int) {
	cols, rows, _ = term.GetSize(fd)
	return cols, rows
}
