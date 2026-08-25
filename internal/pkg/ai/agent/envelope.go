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
	"encoding/json"
	"fmt"
	"strings"
)

// AgentEnvelope is the SSOT structured Agent-mode reply before aissh Action mapping.
type AgentEnvelope struct {
	Action  string
	Command string
	Args    string
	Text    string
}

// IsAnswer reports whether e is the terminal answer action.
func (e AgentEnvelope) IsAnswer() bool { return strings.EqualFold(e.Action, ToolAnswer) }

// IsRun reports whether e is a run-shell-command action.
func (e AgentEnvelope) IsRun() bool {
	a := strings.ToLower(strings.TrimSpace(e.Action))
	return a == XMLInvokeRun || a == ToolRunShell
}

// IsAit reports whether e is an aissh CLI invocation action.
func (e AgentEnvelope) IsAit() bool {
	a := strings.ToLower(strings.TrimSpace(e.Action))
	return a == XMLInvokeAISsh || a == XMLInvokeAit || a == ToolAISshCLI || a == ToolAitCLI
}

// IsToolbox reports whether e is a toolbox CLI invocation action.
func (e AgentEnvelope) IsToolbox() bool {
	return strings.EqualFold(strings.TrimSpace(e.Action), ToolToolbox)
}

// CanonicalJSON returns a single-line JSON envelope suitable for chat history.
// Returns an error when the action verb is unknown or json.Marshal fails.
func (e AgentEnvelope) CanonicalJSON() (string, error) {
	payload := map[string]string{JSONFieldAction: strings.ToLower(strings.TrimSpace(e.Action))}
	switch {
	case e.IsAnswer():
		payload[JSONFieldText] = e.Text
	case e.IsRun():
		payload[JSONFieldCommand] = e.Command
	case e.IsAit(), e.IsToolbox():
		payload[JSONFieldArgs] = e.Args
	default:
		return "", fmt.Errorf("unknown action %q", e.Action)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
