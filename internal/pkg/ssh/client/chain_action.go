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

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
)

// ChainOp selects the in-session chain action after suspend.
type ChainOp int

const (
	// ChainOpLogin dials a new host via Manager.Login.
	ChainOpLogin ChainOp = iota
	// ChainOpLoad reconnects via Manager.Load.
	ChainOpLoad
	// ChainOpSwitch moves to another saved session via Manager.Switch.
	ChainOpSwitch
)

// ChainActionError signals RunInteractive should chain to login/load/switch.
// Args holds tokens after "!ait " excluding the subcommand keyword where noted:
// load → load args; switch → args after "--switch"; login → full connect argv.
type ChainActionError struct {
	Op   ChainOp
	Args []string
}

// Error implements error.
func (e *ChainActionError) Error() string {
	return "ait: chain action"
}

// ParseMetaChainLine detects !ait load/session --switch/connect meta lines.
// lookup resolves saved session aliases for bare-word chain connect; may be nil.
// Returns ok false when the line is not a chain action.
func ParseMetaChainLine(line string, lookup cliroute.AliasLookup) (*ChainActionError, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, meta.PrefixAitSpace) {
		return nil, false
	}

	args := strings.Fields(strings.TrimPrefix(line, meta.PrefixAitSpace))
	if len(args) == 0 {
		return nil, false
	}

	switch args[0] {
	case "load":
		return &ChainActionError{Op: ChainOpLoad, Args: append([]string(nil), args[1:]...)}, true
	case "session":
		if len(args) >= 2 && args[1] == "--switch" {
			return &ChainActionError{
				Op:   ChainOpSwitch,
				Args: append([]string(nil), args[2:]...),
			}, true
		}
		return nil, false
	}

	if cliroute.IsMetaCLIFirstWord(args[0]) || cliroute.IsRemovedTopLevelWord(args[0]) {
		return nil, false
	}
	if !cliroute.IsMetaChainConnectArgs(args, lookup) {
		return nil, false
	}
	return &ChainActionError{Op: ChainOpLogin, Args: append([]string(nil), args...)}, true
}
