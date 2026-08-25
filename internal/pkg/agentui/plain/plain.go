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

// Package plain renders agentcore events as structured text (Phase 1 UI).
package plain

import (
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// Renderer prints agent events to w with CRLF-safe output.
type Renderer struct {
	Out io.Writer
}

// NewEmitter returns an agentcore.FuncEmitter that renders to r.Out.
func (r Renderer) NewEmitter() agentcore.Emitter {
	return agentcore.FuncEmitter(func(e agentcore.Event) {
		r.render(e)
	})
}

func (r Renderer) render(e agentcore.Event) {
	if r.Out == nil {
		return
	}
	switch e.Kind {
	case agentcore.EventToolStarted:
		terminal.WriteTTY(r.Out, fmt.Sprintf("AiT Agent: tool %s\n", e.Tool))
	case agentcore.EventToolOutput:
		if e.Output != "" {
			terminal.WriteTTY(r.Out, indentBlock(e.Output))
		}
	case agentcore.EventToolFinished:
		if e.Detail != "" {
			terminal.WriteTTY(r.Out, fmt.Sprintf("AiT Agent: done %s (%s)\n", e.Tool, e.Detail))
		}
	case agentcore.EventApprovalRequired:
		terminal.WriteTTY(r.Out, fmt.Sprintf("AiT Agent: approval required for %s\n", e.Tool))
	case agentcore.EventStepLimitReached:
		terminal.WriteTTY(r.Out, fmt.Sprintf("AiT Agent: step limit %s\n", e.Detail))
	case agentcore.EventWarning:
		terminal.WriteTTY(r.Out, fmt.Sprintf("AiT Agent: warning: %s\n", e.Detail))
	case agentcore.EventMessageFinal:
		// final message written by loop
	default:
	}
}

func indentBlock(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, ln := range lines {
		lines[i] = "  " + ln
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}
