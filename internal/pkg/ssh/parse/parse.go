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

package parse

import "github.com/os-artificer/AiT/internal/pkg/ssh/openssh"

// Spec holds parsed connection argv (OpenSSH + AiT extensions).
type Spec = openssh.Spec

// TransferOp selects get versus put path pairing semantics.
type TransferOp = openssh.TransferOp

// OpGet downloads remotePath to localPath.
const OpGet = openssh.OpGet

// OpPut uploads localPath to remotePath.
const OpPut = openssh.OpPut

// TransferSpec holds parsed argv for ait get or ait put.
type TransferSpec = openssh.TransferSpec

// Parse parses argv for ait login.
func Parse(argv []string) (Spec, error) {
	return openssh.ParseLogin(argv)
}

// ParseTransfer parses argv for ait get or ait put.
func ParseTransfer(argv []string, op TransferOp) (TransferSpec, error) {
	return openssh.ParseTransfer(argv, op)
}

// ParseDestination splits dest into user, host, and port.
func ParseDestination(dest string) (user, host string, port int, err error) {
	return openssh.ParseDestination(dest)
}

// TmuxWithRemoteCommandErr reports invalid --tmux with a one-shot remote command.
func TmuxWithRemoteCommandErr() error {
	return openssh.TmuxWithRemoteCommandErr()
}
