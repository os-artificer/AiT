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

import "encoding/json"

// Public action names emitted in the Agent-mode JSON envelope.
const (
	ToolRunShell = "run_shell"
	ToolAISshCLI = "aish_cli"
	ToolAitCLI   = "ait_cli" // legacy alias for ToolAISshCLI
	ToolToolbox  = "toolbox"
	ToolAnswer   = "answer"
)

// FCCall is a provider-neutral function-call payload (name + JSON arguments).
type FCCall struct {
	Name      string
	Arguments string
}

// ToolCallToJSON converts a function call into the AiT Agent-mode JSON action envelope.
func ToolCallToJSON(tc FCCall) (string, error) {
	if tc.Name == "" {
		return "", ErrUnsupportedToolCall
	}
	switch tc.Name {
	case ToolRunShell:
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(NormalizeES6UnicodeJSON(tc.Arguments)), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolRunShell, JSONFieldCommand: args.Command})
		return string(b), nil
	case ToolAISshCLI, ToolAitCLI:
		var args struct {
			Args string `json:"args"`
		}
		if err := json.Unmarshal([]byte(NormalizeES6UnicodeJSON(tc.Arguments)), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolAISshCLI, JSONFieldArgs: args.Args})
		return string(b), nil
	case ToolToolbox:
		var args struct {
			Args string `json:"args"`
		}
		if err := json.Unmarshal([]byte(NormalizeES6UnicodeJSON(tc.Arguments)), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolToolbox, JSONFieldArgs: args.Args})
		return string(b), nil
	case ToolAnswer:
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(NormalizeES6UnicodeJSON(tc.Arguments)), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolAnswer, JSONFieldText: args.Text})
		return string(b), nil
	default:
		return "", ErrUnknownTool(tc.Name)
	}
}
