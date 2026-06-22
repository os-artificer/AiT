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

// ToolDef describes one Agent-mode tool for OpenAI-compatible and Eino schema generation.
type ToolDef struct {
	Name      string
	Desc      string
	ParamName string
	ParamJSON string
}

// ToolDefinitions returns the SSOT Agent tool metadata for AiT.
func ToolDefinitions() []ToolDef {
	return []ToolDef{
		{
			Name:      ToolRunShell,
			Desc:      "Run one shell command on the active host",
			ParamName: "command",
			ParamJSON: `{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`,
		},
		{
			Name:      ToolAISshCLI,
			Desc:      "Run ai-ssh CLI subcommand on the local machine",
			ParamName: "args",
			ParamJSON: `{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`,
		},
		{
			Name:      ToolAitCLI,
			Desc:      "Run ai-ssh CLI subcommand on the local machine (legacy alias)",
			ParamName: "args",
			ParamJSON: `{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`,
		},
		{
			Name:      ToolAnswer,
			Desc:      "Reply to the user and finish the tool chain",
			ParamName: "text",
			ParamJSON: `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`,
		},
	}
}
