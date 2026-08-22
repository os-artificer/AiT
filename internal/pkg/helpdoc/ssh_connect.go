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

package helpdoc

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// SSHConnectHelpFull is the canonical SSH connect section for aissh --help.
var SSHConnectHelpFull = sshConnectHelpFull()

// WriteSSHConnectSection prints the default SSH connect usage block to w.
func WriteSSHConnectSection(w io.Writer) error {
	_, err := io.WriteString(w, SSHConnectHelpFull)
	return err
}

// WriteSSHConnectShort prints a compact flag list for shell REPL help.
func WriteSSHConnectShort(w io.Writer) error {
	_, err := io.WriteString(w, sshConnectShortBlock())
	return err
}

// RootUsageLine returns the top-level aissh usage without subcommands.
func RootUsageLine() string {
	return toolname.AISsh + " [flags] destination [command ...]"
}
