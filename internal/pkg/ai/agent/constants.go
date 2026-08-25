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
)

const (
	XMLInvokeRun   = "run"
	XMLInvokeAISsh = "aish"
	XMLInvokeAit   = "ait" // legacy MiniMax XML invoke name

	JSONFieldAction  = "action"
	JSONFieldCommand = "command"
	JSONFieldArgs    = "args"
	JSONFieldText    = "text"

	ToolResultPrefix  = "tool_result:\n"
	FormatErrorPrefix = "format_error: "

	formatCorrectionIntro = "Your previous response was not valid Agent JSON. " +
		"Reply with a single JSON object only (action answer/run_shell/aish_cli/toolbox).\n\n" +
		"Original question:\n"
	formatCorrectionReplyLabel = "\n\nInvalid response (excerpt):\n"
	formatCorrectionErrLabel   = "\n\nParse errors:\n"

	errUnsupportedToolCall = "unsupported tool call"
	errUnknownToolFmt      = "unknown tool %q"

	promptExcerptMaxRunes = 500
	toolErrExcerptRunes   = 200
)

// ActionNamePattern is the alternation of Agent JSON action names for regex use.
const ActionNamePattern = XMLInvokeRun + "|" + ToolRunShell + "|" + XMLInvokeAISsh + "|" + ToolAISshCLI + "|" + XMLInvokeAit + "|" + ToolAitCLI + "|" + ToolToolbox + "|" + ToolAnswer

// ErrUnsupportedToolCall is returned for non-function tool_call responses.
var ErrUnsupportedToolCall = errors.New(errUnsupportedToolCall)

// ErrUnknownTool returns an error for an unrecognized tool name.
func ErrUnknownTool(name string) error {
	return fmt.Errorf(errUnknownToolFmt, name)
}
