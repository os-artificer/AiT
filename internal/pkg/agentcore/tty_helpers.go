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

import "golang.org/x/term"

// suspendRawTTY returns a Config.SuspendTTY func that restores fd to cooked mode
// (state as captured before MakeRaw) for the duration of a child process, so the
// child sees a normal terminal and Ctrl+C is delivered as SIGINT instead of
// being read as a 0x03 byte by the loop's line editor.
// The returned resume re-enters raw mode; it is nil when the handoff failed, in
// which case the caller keeps running in raw mode.
func suspendRawTTY(fd int, state *term.State) func() (resume func()) {
	if state == nil {
		return nil
	}

	return func() func() {
		if err := term.Restore(fd, state); err != nil {
			return nil
		}
		return func() { _, _ = term.MakeRaw(fd) }
	}
}
