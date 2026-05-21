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

import "fmt"

// BasePrompt is the AiT interactive shell prompt when not connected to a remote host.
const BasePrompt = "AiT> "

// LocalPrompt is shown in AiT local mode (!ait local) within an SSH session.
const LocalPrompt = "AiT Local> "

// RemotePrompt returns the AiT display prompt for user@host (e.g. "AiT user@host> ").
// user and host are shown literally; they may be empty.
func RemotePrompt(user, host string) string {
	return fmt.Sprintf("AiT %s@%s> ", user, host)
}

// remoteSetupPromptMarker returns the substring used to detect AiT remote prompt readiness.
func remoteSetupPromptMarker(user, host string) string {
	return RemotePrompt(user, host)
}

// bashPS1 wraps a display prompt for bash readline width counting.
func bashPS1(display string) string {
	return `\[\]` + display + `\[\]`
}

// localPromptPS1 is bash PS1 for local mode.
func localPromptPS1() string {
	return bashPS1(LocalPrompt)
}

// remotePromptPS1 is bash PS1 for remote mode.
func remotePromptPS1(user, host string) string {
	return bashPS1(RemotePrompt(user, host))
}

// remotePromptSetupLine sets PS1/PROMPT while echo is off, clears the default prompt line,
// then restores stty echo (single injected line; local filter suppresses any leaked echo).
func remotePromptSetupLine(user, host string) string {
	display := RemotePrompt(user, host)
	ps1 := remotePromptPS1(user, host)
	return "stty -echo 2>/dev/null; export PS1='" + ps1 + "'; [ -n \"$ZSH_VERSION\" ] && export PROMPT='" + display + "'" +
		"; printf '\\033[1A\\033[2K\\r' 2>/dev/null; stty echo 2>/dev/null"
}

// RemoteSetupStripFragments returns scrollback/filter fragments for remote PS1 setup leakage.
func RemoteSetupStripFragments(user, host string) []string {
	return []string{
		"stty -echo 2>/dev/null",
		remotePromptSetupLine(user, host),
		"export PS1='" + remotePromptPS1(user, host) + "'",
		"export PROMPT='" + RemotePrompt(user, host) + "'",
	}
}

// RemotePromptInitialCommands returns shell setup lines to set PS1/PROMPT after the remote shell starts.
// Returns nil when user or host is empty.
func RemotePromptInitialCommands(user, host string) []string {
	if user == "" || host == "" {
		return nil
	}
	return []string{remotePromptSetupLine(user, host)}
}
