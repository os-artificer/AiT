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

package store

// AliasOrID returns the session alias when set, otherwise the first 8 characters of ID
// (or the full ID when shorter). Used for tables and user-facing labels.
func (s *Session) AliasOrID() string {
	if s.Alias != "" {
		return s.Alias
	}

	if len(s.ID) > 8 {
		return s.ID[:8]
	}
	return s.ID
}
