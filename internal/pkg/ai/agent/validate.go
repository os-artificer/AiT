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

package agent

import (
	"errors"
	"fmt"
	"strings"
)

// ParseSource identifies which ResolveAgentReply strategy produced an envelope.
type ParseSource string

const (
	SourceJSON      ParseSource = "json"
	SourceXML       ParseSource = "xml"
	SourceHeuristic ParseSource = "heuristic"
	SourceProse     ParseSource = "prose"
)

// ParseResult holds the outcome of ResolveAgentReply.
type ParseResult struct {
	Envelope AgentEnvelope
	Source   ParseSource
	Errs     []error
}

// OK reports whether the envelope parsed and passed ValidateEnvelope.
func (r ParseResult) OK() bool {
	return r.Source != "" && ValidateEnvelope(r.Envelope) == nil
}

// ValidateEnvelope checks required fields for each Agent action verb.
// Returns nil when e is valid. Answer may have empty Text (displayed as done).
// Returns an error for empty action, unknown verbs, or missing command/args payloads.
func ValidateEnvelope(e AgentEnvelope) error {
	action := strings.ToLower(strings.TrimSpace(e.Action))
	if action == "" {
		return errors.New("empty action")
	}
	switch {
	case action == ToolAnswer:
		return nil
	case action == ToolRunShell || action == XMLInvokeRun:
		if strings.TrimSpace(e.Command) == "" {
			return fmt.Errorf("%s requires command", action)
		}
		return nil
	case action == ToolAISshCLI || action == ToolAitCLI || action == XMLInvokeAISsh || action == XMLInvokeAit:
		if strings.TrimSpace(e.Args) == "" {
			return fmt.Errorf("%s requires args", action)
		}
		return nil
	case action == ToolToolbox:
		if strings.TrimSpace(e.Args) == "" {
			return fmt.Errorf("%s requires args", action)
		}
		return nil
	default:
		return fmt.Errorf("unknown action %q", e.Action)
	}
}
