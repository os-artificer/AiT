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

package prompt

const (
	historyOffBashBegin = "set +o history 2>/dev/null"
	historyOffBashEnd   = "set -o history 2>/dev/null"

	historyOffZshBegin = `[ -n "$ZSH_VERSION" ] && _AIT_HIST=$HISTFILE HISTFILE=/dev/null`
	historyOffZshEnd   = `[ -n "$ZSH_VERSION" ] && HISTFILE=$_AIT_HIST unset _AIT_HIST`
)

// wrapHistoryOff wraps a shell compound command so it is not written to bash/zsh history.
// Caller must run the result via Inject (stty echo suppression); this helper only toggles history.
func wrapHistoryOff(body string) string {
	return historyOffBashBegin +
		"; " + historyOffZshBegin +
		"; " + body +
		"; " + historyOffZshEnd +
		"; " + historyOffBashEnd
}
