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

// Package meta is the byte-level state machine that detects in-session
// AiT meta commands (!ait <subcmd>, !!) on a remote PTY stream. It is
// intentionally free of side effects so the parent ssh/client package can
// drive it from any input source. The state machine itself (Scanner) is the
// only design pattern here; it predates this refactor and stays as-is.
package meta

// PrefixAitSpace is the "!ait " prefix used for in-session AiT meta lines.
const PrefixAitSpace = "!ait "

// PrefixAitAI is the "!ait ai" prefix used to detect the AI subcommand.
const PrefixAitAI = "!ait ai"

// Meta line constants — exact matches the scanner recognizes.
const (
	LineSuspend = "!ait suspend"
	LineLocal   = "!ait local"
	LineRemote  = "!ait remote"
	LineBare    = "!ait"
	LineHelp    = "!ait help"
)

// prefixes is the exhaustive set of strict meta openers. The scanner
// only commits to meta mode once a candidate buffer matches one exactly.
var prefixes = []string{
	"!ait ai ",
	"!ait ai\n",
	"!ait local ",
	"!ait local\n",
	"!ait remote\n",
	PrefixAitSpace,
	"!! ",
	"!!\n",
}
