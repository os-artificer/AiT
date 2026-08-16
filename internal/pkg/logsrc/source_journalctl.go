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

package logsrc

import "context"

// JournalctlSource reads the systemd journal by running journalctl with JSON
// output. Kernel selects `-k` (kernel ring buffer) instead of all units.
type JournalctlSource struct {
	Resolver *Resolver
	Kernel   bool
	Path     string
}

// Name returns the chain name: the kernel source when restricted to kernel
// messages, and the tool name otherwise.
func (s *JournalctlSource) Name() string {
	if s.Kernel {
		return SourceNameKernel
	}
	return ToolJournalctl
}

// SupportsReverse reports true: last-N queries add --reverse so journalctl
// yields newest-first and the scan stops after N matches.
func (s *JournalctlSource) SupportsReverse() bool { return true }

// Open starts journalctl for q. Returns ErrToolMissing when journalctl is not
// on PATH, or a process start error.
func (s *JournalctlSource) Open(ctx context.Context, q Query) (Cursor, error) {
	path, err := s.Resolver.toolPath(s.Path, ToolJournalctl)
	if err != nil {
		return nil, err
	}

	reverse := !q.Head && !q.Follow
	return OpenExecCursor(ctx, ExecCursorOptions{
		Path:  path,
		Args:  BuildJournalctlArgs(q, reverse, s.Kernel),
		Parse: ParseJournalJSON,
	})
}
