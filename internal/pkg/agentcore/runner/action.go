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

// Package agent represents the AI agent loop's parsed actions and the local
// capture runner used to feed shell output back into the model.
package runner

// Action verbs returned by the model in Agent mode.
const (
	// ActionRun is the canonical verb for "run a shell command".
	ActionRun = "run"
	// ActionRunAlt is the OpenAI function-calling style alias for ActionRun.
	ActionRunAlt = "run_shell"
	// ActionAISsh is the canonical verb for "invoke an aissh CLI subcommand".
	ActionAISsh = "aish"
	// ActionAISshAlt is the OpenAI function-calling style alias for ActionAISsh.
	ActionAISshAlt = "aish_cli"
	// ActionAit is the legacy verb alias for ActionAISsh.
	ActionAit = "ait"
	// ActionAitAlt is the legacy OpenAI function-calling style alias for ActionAISsh.
	ActionAitAlt = "ait_cli"
	// ActionToolbox runs an AiT toolbox binary (aihost/ailog/aissh/aiimg).
	ActionToolbox = "toolbox"
	// ActionAnswer terminates the agent loop with a free-text reply.
	ActionAnswer = "answer"
)

// Action is one structured step returned by the model in Agent mode.
// Action is the verb; Command/Args/Text carry verb-specific payload.
type Action struct {
	Action  string
	Command string
	Args    string
	Text    string
}

// IsAnswer reports whether a is the terminal answer action.
func (a Action) IsAnswer() bool { return a.Action == ActionAnswer }

// IsRun reports whether a is a run-shell-command action (canonical or alt verb).
func (a Action) IsRun() bool {
	return a.Action == ActionRun || a.Action == ActionRunAlt
}

// IsAit reports whether a is an aissh CLI invocation action (canonical, alt, or legacy verbs).
func (a Action) IsAit() bool {
	return a.Action == ActionAISsh || a.Action == ActionAISshAlt ||
		a.Action == ActionAit || a.Action == ActionAitAlt
}

// IsToolbox reports whether a is a toolbox CLI invocation.
func (a Action) IsToolbox() bool {
	return a.Action == ActionToolbox
}
