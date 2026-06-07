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

import (
	"bytes"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

type setupOutputFilter struct {
	pending []byte
}

func (f *setupOutputFilter) filter(p []byte, user, host string) []byte {
	if len(p) == 0 {
		return nil
	}
	f.pending = append(f.pending, p...)
	var out bytes.Buffer
	for {
		idx := bytes.IndexByte(f.pending, '\n')
		if idx < 0 {
			break
		}
		line := string(f.pending[:idx])
		f.pending = f.pending[idx+1:]
		if bytes.IndexByte(f.pending, '\n') == 0 && len(f.pending) > 0 {
			// drop empty line after \n
		}
		if !isSetupLeakLine(line, user, host) {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	return out.Bytes()
}

func (f *setupOutputFilter) flush(user, host string) []byte {
	if len(f.pending) == 0 {
		return nil
	}
	line := string(f.pending)
	f.pending = nil
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	return []byte(line)
}

func isSetupLeakLine(line, user, host string) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}
	if strings.Contains(line, "stty -echo") {
		return true
	}
	if strings.Contains(line, "set +o history") || strings.Contains(line, "set -o history") {
		return true
	}
	if strings.Contains(line, "_AIT_HIST") || strings.Contains(line, "HISTFILE=/dev/null") {
		return true
	}
	prefix := prompt.RemotePrefix(user, host)
	if strings.Contains(line, "export PS1='") && strings.Contains(line, prefix) {
		return true
	}
	if strings.Contains(line, "export PROMPT='") && strings.Contains(line, prefix) {
		return true
	}
	if strings.Contains(line, "printf '\\033[1A") {
		return true
	}
	return false
}

// filterRemoteOutputChunk applies setup suppression and duplicate-prompt dropping before scroll/stdout.
// Returns nil when the chunk should not be forwarded.
func (c *Client) filterRemoteOutputChunk(chunk []byte, user, host, promptMarker string) []byte {
	if c.remoteSetupSuppress.Load() {
		chunk = c.setupFilter.filter(chunk, user, host)
		if strings.Contains(string(chunk), promptMarker) {
			c.remoteSetupSuppress.Store(false)
		}
	}
	if len(chunk) == 0 {
		return nil
	}
	if c.remotePromptShown.Load() && isDuplicatePromptOutput(chunk, user, host) {
		return nil
	}
	return chunk
}

func isDuplicatePromptOutput(p []byte, user, host string) bool {
	if len(p) == 0 {
		return false
	}
	s := strings.TrimSpace(string(p))
	if s == "" {
		return false
	}
	return prompt.IsRemotePromptLine(user, host, s)
}
