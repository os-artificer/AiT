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

import (
	"context"
	"time"
)

// LogShowSource reads the macOS unified log by running `log show`, or
// `log stream` when following. The unified log cannot be read newest-first, so
// this source is a WindowedSource: unbounded last-N queries start from the
// narrowest window in LogShowWindows and widen until they have enough matches.
type LogShowSource struct {
	Resolver *Resolver
	Path     string
	NDJSON   bool
}

// Name returns the chain name used in notices and `ailog sources`.
func (s *LogShowSource) Name() string { return ToolLog }

// SupportsReverse reports false: `log show` only emits oldest-first.
func (s *LogShowSource) SupportsReverse() bool { return false }

// Windows returns the expanding search windows used for unbounded queries.
func (s *LogShowSource) Windows() []time.Duration { return LogShowWindows }

// WindowsFor returns LogShowWindowsFiltered when q has keywords or a regex,
// otherwise LogShowWindows.
func (s *LogShowSource) WindowsFor(q Query) []time.Duration {
	if hasTextFilter(q) {
		return LogShowWindowsFiltered
	}
	return LogShowWindows
}

// Open starts `log show` over the narrowest window (or `log stream` when
// q.Follow is set). An explicit q.Since disables windowing. Returns
// ErrToolMissing when `log` is not on PATH.
func (s *LogShowSource) Open(ctx context.Context, q Query) (Cursor, error) {
	if q.Follow {
		return s.openStream(ctx, q)
	}
	windows := s.WindowsFor(q)
	window := windows[0]
	if !q.Since.IsZero() {
		window = 0
	}
	return s.OpenWindow(ctx, q, window)
}

// OpenWindow starts `log show` bounded to the given window. A zero window
// means "no --last bound", which q.Since is then expected to provide.
// Returns ErrToolMissing when `log` is not on PATH.
func (s *LogShowSource) OpenWindow(ctx context.Context, q Query, window time.Duration) (Cursor, error) {
	path, err := s.Resolver.toolPath(s.Path, ToolLog)
	if err != nil {
		return nil, err
	}
	return OpenExecCursor(ctx, ExecCursorOptions{
		Path:  path,
		Args:  BuildLogShowArgs(q, window, s.NDJSON),
		Parse: s.parser(),
	})
}

// openStream starts `log stream` for --follow, which always uses ndjson.
func (s *LogShowSource) openStream(ctx context.Context, q Query) (Cursor, error) {
	path, err := s.Resolver.toolPath(s.Path, ToolLog)
	if err != nil {
		return nil, err
	}
	return OpenExecCursor(ctx, ExecCursorOptions{
		Path:  path,
		Args:  BuildLogStreamArgs(q),
		Parse: ParseLogShowNDJSON,
	})
}

// parser returns the LineParser matching the configured output style.
func (s *LogShowSource) parser() LineParser {
	if s.NDJSON {
		return ParseLogShowNDJSON
	}
	return func(doc []byte) (Record, error) {
		rec, ok := ParseLogShowSyslog(string(doc))
		if !ok {
			return Record{}, ErrEmptyRecord
		}
		return rec, nil
	}
}
