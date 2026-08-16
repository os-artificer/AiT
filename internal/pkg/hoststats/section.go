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

package hoststats

// Section is a display strategy for one metrics category.
type Section interface {
	// ID returns the stable section identifier (e.g. "cpu").
	ID() string
	// Human formats snap as a human-readable block for this section.
	Human(s Snapshot) string
	// JSONValue returns the JSON-serializable value for this section.
	JSONValue(s Snapshot) any
}
