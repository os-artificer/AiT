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

package meta

import "github.com/os-artificer/AiT/internal/pkg/toolname"

// PrefixAitSpace is the "!ait " prefix used for in-session AiT meta lines.
const PrefixAitSpace = toolname.MetaPrefix + " "

// PrefixAitAI is the "!ait ai" prefix used to detect the AI subcommand.
const PrefixAitAI = toolname.MetaPrefix + " ai"

// bangAitTypoPrefix is "!!" + MetaPrefix without leading "!" (!!ait). Keep in sync with toolname.MetaPrefix.
const bangAitTypoPrefix = "!!ait"

// Meta line constants — exact matches the scanner recognizes.
const (
	LineSuspend = toolname.MetaPrefix + " suspend"
	LineLocal   = toolname.MetaPrefix + " local"
	LineRemote  = toolname.MetaPrefix + " remote"
	LineBare    = toolname.MetaPrefix
	LineHelp    = toolname.MetaPrefix + " help"
	LineChain   = toolname.MetaPrefix + " chain"
)

// prefixes is the exhaustive set of strict meta openers. The scanner
// only commits to meta mode once a candidate buffer matches one exactly.
var prefixes = []string{
	toolname.MetaPrefix + " local ",
	toolname.MetaPrefix + " local\n",
	toolname.MetaPrefix + " remote\n",
	toolname.MetaPrefix + " chain ",
	toolname.MetaPrefix + " chain\n",
	PrefixAitSpace,
}
