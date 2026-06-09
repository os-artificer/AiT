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
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

const proxyJumpSep = ","

// formatProxyJump serializes jump hosts for SQLite storage.
func formatProxyJump(jump []string) string {
	if len(jump) == 0 {
		return ""
	}
	return strings.Join(jump, proxyJumpSep)
}

// parseProxyJump deserializes stored proxy_jump text.
func parseProxyJump(stored string) []string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return nil
	}
	parts := strings.Split(stored, proxyJumpSep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// sessionDialDestination returns the ssh -G destination for load/switch re-resolve.
func sessionDialDestination(sess *store.Session) string {
	if sess.LoginDestination != "" {
		return sess.LoginDestination
	}
	if sess.Alias != "" {
		return sess.Alias
	}
	if sess.Username != "" {
		return sess.Username + "@" + sess.Host
	}
	return sess.Host
}
