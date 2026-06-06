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

// Package transfer holds receiver-free helpers for the in-session !ait get
// / !ait put meta commands: meta-line argv parsing and the user-visible
// usage hint. The actual SFTP transfer methods (Client.handleTransfer*)
// stay in the parent ssh/client package because they depend on the live
// SSH Client receiver.
package transfer

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// UsageGet is the user-facing one-line usage for `!ait get`.
const UsageGet = "usage: !ait get <remote-path> [<local-path>]"

// UsagePut is the user-facing one-line usage for `!ait put`.
const UsagePut = "usage: !ait put <local-path> [<remote-path>]"

const metaPrefixAitSpace = "!ait "

// CmdGet is the meta subcommand name for !ait get (must match ait get.Name).
const CmdGet = "get"

// CmdPut is the meta subcommand name for !ait put (must match ait put.Name).
const CmdPut = "put"

// ArgsFromMetaLine extracts the subcommand and trailing args from a `!ait <cmd> ...`
// meta line when cmd is one of get/put. Returns ok=false for other lines.
func ArgsFromMetaLine(line string) (cmd string, args []string, ok bool) {
	if !strings.HasPrefix(line, metaPrefixAitSpace) {
		return "", nil, false
	}
	parts := strings.Fields(strings.TrimPrefix(line, metaPrefixAitSpace))
	if len(parts) == 0 {
		return "", nil, false
	}
	switch parts[0] {
	case CmdGet, CmdPut:
		return parts[0], parts[1:], true
	default:
		return "", nil, false
	}
}

// UsageHintErr returns the standard multi-line errhint for transfer argv errors.
// usage is one of UsageGet / UsagePut.
func UsageHintErr(usage string) error {
	return errhint.WithTryMsg(usage,
		"Provide paths; omitted destination defaults to the source basename.",
		"!ait get <remote-path>",
		"!ait get <remote-path> <local-path>",
		"!ait put <local-path>",
		"!ait put <local-path> <remote-path>",
		"ait get --help",
		"ait put --help",
	)
}
