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

func nextSetupFilterLine(pending []byte) (line string, rest []byte, ok bool) {
	idx := bytes.IndexByte(pending, '\n')
	if idx < 0 {
		idx = bytes.IndexByte(pending, '\r')
	}
	if idx < 0 {
		return "", pending, false
	}

	line = strings.TrimRight(string(pending[:idx]), "\r")
	rest = pending[idx+1:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	return line, rest, true
}

func (f *setupOutputFilter) filter(p []byte, user, host string) []byte {
	if len(p) == 0 {
		return nil
	}
	f.pending = append(f.pending, p...)
	var out bytes.Buffer
	for {
		line, rest, ok := nextSetupFilterLine(f.pending)
		if !ok {
			break
		}
		f.pending = rest
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
	line := strings.TrimRight(string(f.pending), "\r\n")
	f.pending = nil
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	return []byte(line)
}

func filterSetupLeakBytes(p []byte, user, host string) []byte {
	var f setupOutputFilter
	out := f.filter(p, user, host)
	if tail := f.flush(user, host); len(tail) > 0 {
		out = append(out, tail...)
	}
	return out
}

type setupStdoutStripper struct {
	pending []byte
}

// nextSetupStdoutLine splits one complete line from pending, returning raw bytes (content + original separator).
func nextSetupStdoutLine(pending []byte) (line string, kept []byte, rest []byte, ok bool) {
	idxNL := bytes.IndexByte(pending, '\n')
	idxCR := bytes.IndexByte(pending, '\r')
	if idxNL < 0 && idxCR < 0 {
		return "", nil, pending, false
	}

	var end int
	if idxNL >= 0 && (idxCR < 0 || idxNL <= idxCR) {
		end = idxNL + 1
		if idxCR >= 0 && idxCR == idxNL-1 {
			line = string(pending[:idxCR])
		} else {
			line = string(pending[:idxNL])
		}
	} else {
		line = string(pending[:idxCR])
		end = idxCR + 1
		if idxNL >= 0 && idxNL == idxCR+1 {
			end = idxNL + 1
		}
	}

	kept = pending[:end]
	rest = pending[end:]
	return line, kept, rest, true
}

func (s *setupStdoutStripper) strip(user, host string) []byte {
	if len(s.pending) == 0 {
		return nil
	}

	var out bytes.Buffer
	for {
		line, kept, rest, ok := nextSetupStdoutLine(s.pending)
		if !ok {
			break
		}
		s.pending = rest
		if !isSetupLeakLine(line, user, host) {
			out.Write(kept)
		}
	}
	return out.Bytes()
}

func (s *setupStdoutStripper) flushStdout(user, host string) []byte {
	if len(s.pending) == 0 {
		return nil
	}
	line := strings.TrimRight(string(s.pending), "\r\n")
	tail := append([]byte(nil), s.pending...)
	s.pending = nil
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	return tail
}

// stripSetupLeakStdout removes setup leak lines from stdout-bound PTY output while preserving original line endings.
func stripSetupLeakStdout(p []byte, user, host string) []byte {
	var s setupStdoutStripper
	s.pending = append(s.pending, p...)
	out := s.strip(user, host)
	if tail := s.flushStdout(user, host); len(tail) > 0 {
		out = append(out, tail...)
	}
	return out
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
		if c.setupInjectDone.Load() && strings.Contains(string(chunk), promptMarker) {
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
