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

package aimode

import "fmt"

func confirmPromptRisk(opt Options, cmd string, risk RiskLevel) string {
	label := scopeLabel(opt)

	if risk == RiskHigh {
		return fmt.Sprintf(
			"AiT: [%s] HIGH-RISK: %s\n  [y] run anyway   [n] skip\nChoice [y/N]: ",
			label, cmd,
		)
	}
	return fmt.Sprintf(
		"AiT: [%s] %s\n  [y] run once   [c] trust this command pattern   [s] trust safe commands this session   [n] skip\nChoice [y/c/s/N]: ",
		label, cmd,
	)
}
