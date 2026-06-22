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

package session

import "github.com/os-artificer/AiT/internal/pkg/clihints"

// Re-export Try-block lines from clihints (avoids import cycle with ssh/client).
const (
	TrySessionList            = clihints.TrySessionList
	TrySessionLoadAlias       = clihints.TrySessionLoadAlias
	TrySessionLoad            = clihints.TrySessionLoad
	TrySessionLoadOptional    = clihints.TrySessionLoadOptional
	TrySessionLoadFullID      = clihints.TrySessionLoadFullID
	TrySessionLoadPick        = clihints.TrySessionLoadPick
	TrySessionSwitchID        = clihints.TrySessionSwitchID
	TrySessionSwitch          = clihints.TrySessionSwitch
	TrySessionSwitchForce     = clihints.TrySessionSwitchForce
	TrySessionHelp            = clihints.TrySessionHelp
	TryConnectSave            = clihints.TryConnectSave
	TryConnectAlias           = clihints.TryConnectAlias
	TryConnectHelp            = clihints.TryConnectHelp
	TryConnectTmuxUserHost    = clihints.TryConnectTmuxUserHost
	TryConnectTmuxDevhost     = clihints.TryConnectTmuxDevhost
	TryConnectTmuxSaveDevhost = clihints.TryConnectTmuxSaveDevhost
)

// Private aliases for in-package hint helpers.
const (
	tryLoadByAlias     = clihints.TrySessionLoadAlias
	tryLoadByFullID    = clihints.TrySessionLoadFullID
	tryLoadInteractive = clihints.TrySessionLoadPick
	trySessionHelp     = clihints.TrySessionHelp
	tryLoadFromLocal   = clihints.TrySessionLoadOptional
	trySwitchForce     = clihints.TrySessionSwitchForce
	tryConnectSave     = clihints.TryConnectSave
	tryConnectAlias    = clihints.TryConnectAlias
	tryConnectHelp     = clihints.TryConnectHelp
)
