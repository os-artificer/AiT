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

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WevtutilSource reads a Windows Event Log channel by running
// `wevtutil qe /f:xml`. It backs the in-process wevtapi.dll source and is used
// when that API is unavailable.
type WevtutilSource struct {
	Resolver *Resolver
	Channel  string
	Path     string
}

// Name returns the chain name used in notices and `ailog sources`.
func (s *WevtutilSource) Name() string { return ToolWevtutil }

// SupportsReverse reports true: last-N queries pass /rd:true so the newest
// events arrive first.
func (s *WevtutilSource) SupportsReverse() bool { return true }

// Open starts wevtutil for the channel mapped from q.Source. Returns
// ErrFollowUnsupported when q.Follow is set, ErrToolMissing when wevtutil is
// not on PATH, or a process start error.
func (s *WevtutilSource) Open(ctx context.Context, q Query) (Cursor, error) {
	if q.Follow {
		return nil, errhint.WithTry(
			ErrFollowUnsupported,
			"Follow is not supported for Windows Event Log.",
			toolname.AILog+" -n 50",
			toolname.AILog+" --since 1h",
		)
	}

	path, err := s.Resolver.toolPath(s.Path, ToolWevtutil)
	if err != nil {
		return nil, err
	}

	channel := s.Channel
	if channel == "" {
		channel = ChannelForSource(q.Source)
	}

	reverse := !q.Head && !q.Follow
	return OpenExecCursor(ctx, ExecCursorOptions{
		Path:  path,
		Args:  BuildWevtutilArgs(channel, q, reverse),
		Parse: ParseEventXML,
		Frame: FrameEventXML(),
	})
}
