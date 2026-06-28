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

package eino

import (
	"github.com/os-artificer/AiT/internal/pkg/ai/agent"

	"github.com/cloudwego/eino/schema"
)

// AgentToolInfos returns Eino tool schema for AiT Agent mode.
func AgentToolInfos() ([]*schema.ToolInfo, error) {
	out := make([]*schema.ToolInfo, 0, len(agent.ToolDefinitions()))
	for _, def := range agent.ToolDefinitions() {
		out = append(out, &schema.ToolInfo{
			Name: def.Name,
			Desc: def.Desc,
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				def.ParamName: {Type: schema.String, Required: true},
			}),
		})
	}
	return out, nil
}
