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

package client

// CloneHistory returns a shallow copy of chat history so callers can mutate without aliasing the input slice.
func CloneHistory(h []ChatMessage) []ChatMessage {
	return append([]ChatMessage{}, h...)
}

// AppendUserTurn appends one user message without committing an assistant reply.
// h is copied first; content becomes the new last message with role user.
func AppendUserTurn(h []ChatMessage, content string) []ChatMessage {
	return append(CloneHistory(h), ChatMessage{Role: "user", Content: content})
}

// ReplaceLastAssistant copies updated history and replaces the last assistant content with canonical.
// When updated is empty or the last message is not assistant, returns a copy unchanged.
func ReplaceLastAssistant(updated []ChatMessage, canonical string) []ChatMessage {
	out := append([]ChatMessage{}, updated...)
	if len(out) == 0 {
		return out
	}

	last := len(out) - 1
	if out[last].Role == "assistant" {
		out[last].Content = canonical
	}
	return out
}
