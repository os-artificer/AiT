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

package agentcore

import "github.com/os-artificer/AiT/internal/pkg/errhint"

const (
	defaultMaxSteps = 12

	msgAgentStepLimitFmt = "AiT Agent: step limit (%d) reached; type continue to proceed or exit to stop.\r\n"
	msgSkipped           = "skipped.\r\n"
	msgInteractiveSkip   = "interactive TUI commands are not auto-run; run manually.\r\n"
	msgConversationClear = "conversation cleared.\r\n"

	metaExit  = "exit"
	metaStop  = "stop"
	metaBack  = "back"
	metaModel = "model"
)

// ErrSessionEnded is returned when the agent loop exits normally.
var ErrSessionEnded = errhint.NewSentinel("agent session ended")
