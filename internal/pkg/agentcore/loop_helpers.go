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

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// continueYesWords accept the step-limit prompt and resume the previous task.
var continueYesWords = []string{metaContinue, "y", "yes"}

// continueNoWords end the session at the step-limit prompt.
var continueNoWords = []string{metaExit, metaStop, metaBack, "n", "no", "quit"}

// parseContinueChoice maps one step-limit prompt answer to the next goal.
// goal is msgContinueGoal to resume, "" to end the session, or the typed line to
// use as a new goal. decided is false for a blank line, which must re-prompt.
func parseContinueChoice(line string) (goal string, decided bool) {
	line = strings.TrimSpace(line)
	lower := strings.ToLower(line)

	switch {
	case lower == "":
		return "", false
	case containsWord(continueYesWords, lower):
		return msgContinueGoal, true
	case containsWord(continueNoWords, lower):
		return "", true
	default:
		return line, true
	}
}

func containsWord(words []string, word string) bool {
	for _, w := range words {
		if w == word {
			return true
		}
	}
	return false
}

// loopPrompt returns the Agent Loop prompt, tagged with the host when the
// session is attached to a remote scope.
func loopPrompt(scope ai.Scope, host string) string {
	if scope == ai.ScopeRemote && host != "" {
		return fmt.Sprintf(promptAgentHostFmt, host)
	}
	return promptAgent
}
