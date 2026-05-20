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

package sshclient

import "fmt"

// BasePrompt is the AiT interactive shell prompt when not connected to a remote host.
const BasePrompt = "AiT> "

// LocalPrompt is shown in AiT local mode (!ait local) within an SSH session.
const LocalPrompt = "AiT Local> "

// RemotePrompt returns the display prompt for a remote SSH session.
func RemotePrompt(user, host string) string {
	return fmt.Sprintf("AiT %s@%s> ", user, host)
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

// remotePromptExportLine returns a POSIX one-liner to set PS1/PROMPT on the remote shell.
func remotePromptExportLine(user, host string) string {
	display := RemotePrompt(user, host)
	ps1 := remotePromptPS1(user, host)
	return "export PS1='" + ps1 + "'; [ -n \"$ZSH_VERSION\" ] && export PROMPT='" + display + "'"
}

// RemotePromptInitialCommands returns commands to inject after the remote shell starts.
func RemotePromptInitialCommands(user, host string) []string {
	if user == "" || host == "" {
		return nil
	}
	return []string{remotePromptExportLine(user, host)}
}
