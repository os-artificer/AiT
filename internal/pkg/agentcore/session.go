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
	"io"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/approve"
	"github.com/os-artificer/AiT/internal/pkg/ai/client"
)

type sessionState struct {
	history []client.ChatMessage
	approve.Trust
	// readLine is the single input path for every session sub-prompt, so raw and
	// cooked terminals never mix a Scanner with the line editor on one fd.
	readLine func(prompt string) (string, error)
}

func newSessionState(cfg Config) *sessionState {
	return &sessionState{
		Trust:    approve.Trust{Policy: approve.PolicyPerCmd},
		readLine: scannerLineReader(cfg.In, cfg.Out),
	}
}

// ReadLine writes prompt and reads one line from the session input reader.
// Returns io.EOF when input is exhausted, or the reader's error.
func (st *sessionState) ReadLine(prompt string) (string, error) {
	if st == nil || st.readLine == nil {
		return "", io.EOF
	}
	return st.readLine(prompt)
}
