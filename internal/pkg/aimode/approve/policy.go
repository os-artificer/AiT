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

// Package approve provides command confirmation, pattern/session trust, and
// risk-aware policy for the AI mode REPL.
package approve

// Policy controls how often run/aissh commands require user confirmation.
type Policy int

const (
	// PolicyPerCmd asks before each run (default).
	PolicyPerCmd Policy = iota
	// PolicySessionOnce asks once per AI REPL session, then trusts low-risk commands until exit.
	PolicySessionOnce
)
