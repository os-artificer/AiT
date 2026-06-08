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

// Package agent declares the function-calling tool definitions and the
// adapter from OpenAI tool_calls to the AiT internal Agent-mode JSON action
// envelope (run_shell / ait_cli / answer). The Template-Method-shaped agent
// turn loop (plan -> confirm -> execute -> reply) stays in the parent ai
// package because it is implemented as receiver methods on the unexported
// endpoint type and the Client facade.
package agent

import (
	"encoding/json"

	openai "github.com/sashabaranov/go-openai"
)

// Public action names emitted in the Agent-mode JSON envelope.
const (
	ToolRunShell = "run_shell"
	ToolAitCLI   = "ait_cli"
	ToolAnswer   = "answer"
)

// Tools returns the OpenAI function-calling tool definitions for AiT Agent mode.
func Tools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolRunShell,
				Description: "Run one shell command on the active host",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolAitCLI,
				Description: "Run ait CLI subcommand on the local machine",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolAnswer,
				Description: "Reply to the user and finish the tool chain",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`),
			},
		},
	}
}

// ToolCallToJSON converts an OpenAI tool_call response into the AiT Agent-mode
// JSON action envelope: `{"action":"<tool>","<arg>":"<value>"}`.
// Returns an error for unsupported call types or unknown tool names.
func ToolCallToJSON(tc openai.ToolCall) (string, error) {
	if tc.Type != openai.ToolTypeFunction || tc.Function.Name == "" {
		return "", ErrUnsupportedToolCall
	}
	switch tc.Function.Name {
	case ToolRunShell:
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolRunShell, JSONFieldCommand: args.Command})
		return string(b), nil
	case ToolAitCLI:
		var args struct {
			Args string `json:"args"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolAitCLI, JSONFieldArgs: args.Args})
		return string(b), nil
	case ToolAnswer:
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{JSONFieldAction: ToolAnswer, JSONFieldText: args.Text})
		return string(b), nil
	default:
		return "", ErrUnknownTool(tc.Function.Name)
	}
}
