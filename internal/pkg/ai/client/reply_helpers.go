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

import "github.com/os-artificer/AiT/internal/pkg/ai/eino"

// ExtractReply splits raw LLM response content into a clean reply and optional thinking text.
// Delegates to internal/pkg/ai/eino (E9: thinking is not stored in chat history).
func ExtractReply(raw string) (reply, thinking string) {
	return eino.ExtractReply(raw)
}
