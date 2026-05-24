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

package ai

import (
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

const (
	agentToolRunShell = "run_shell"
	agentToolAitCLI   = "ait_cli"
	agentToolAnswer   = "answer"
)

func agentTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        agentToolRunShell,
				Description: "Run one shell command on the active host",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        agentToolAitCLI,
				Description: "Run ait CLI subcommand on the local machine",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        agentToolAnswer,
				Description: "Reply to the user and finish the tool chain",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`),
			},
		},
	}
}

func toolCallToAgentJSON(tc openai.ToolCall) (string, error) {
	if tc.Type != openai.ToolTypeFunction || tc.Function.Name == "" {
		return "", fmt.Errorf("unsupported tool call")
	}
	switch tc.Function.Name {
	case agentToolRunShell:
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{"action": "run_shell", "command": args.Command})
		return string(b), nil
	case agentToolAitCLI:
		var args struct {
			Args string `json:"args"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{"action": "ait_cli", "args": args.Args})
		return string(b), nil
	case agentToolAnswer:
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]string{"action": "answer", "text": args.Text})
		return string(b), nil
	default:
		return "", fmt.Errorf("unknown tool %q", tc.Function.Name)
	}
}
