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

// metaPrefixes is the exhaustive set of strict meta openers. The scanner
// only commits to meta mode once a candidate buffer matches one exactly.
var metaPrefixes = []string{
	"!ait ai ",
	"!ait ai\n",
	"!ait local ",
	"!ait local\n",
	"!ait remote\n",
	metaPrefixAitSpace,
	"!! ",
	"!!\n",
}

// Meta — in-session !ait / !! prefix strings.
const (
	metaPrefixAitSpace = "!ait "
	metaPrefixAitAI    = "!ait ai"
	metaLineSuspend    = "!ait suspend"
	metaLineLocal      = "!ait local"
	metaLineRemote     = "!ait remote"
	metaLineBare       = "!ait"
	metaLineHelp       = "!ait help"
)

// IO — CRLF line endings for remote TTY output.
const msgCRLF = "\r\n"
