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

package aitplugin

import (
	"time"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var reservedNames = map[string]struct{}{
	CmdHelp:    {},
	CmdExit:    {},
	CmdQuit:    {},
	CmdVersion: {},
	CmdHistory: {},
	CmdClear:   {},
	CmdInstall: {},
	CmdUpdate:  {},
	CmdUpgrade: {},
}

var manageCommandNames = map[string]struct{}{
	CmdInstall: {},
	CmdUpdate:  {},
	CmdUpgrade: {},
}

// BuiltinBinaryNames maps standard short names to shipped tool binaries (PATH fill).
var BuiltinBinaryNames = []struct {
	Binary string
	Name   string
}{
	{Binary: toolname.AISsh, Name: NameSSH},
	{Binary: toolname.AILog, Name: NameLog},
	{Binary: toolname.AIImg, Name: NameImg},
	{Binary: toolname.AIHost, Name: NameHost},
}

const (
	apiVersionV1 = 1

	// ManifestArg is the hidden argv that prints plugin.json to stdout.
	ManifestArg = "ait-plugin-manifest"

	// Reserved REPL builtins (must not be plugin short names).
	CmdHelp    = "help"
	CmdExit    = "exit"
	CmdQuit    = "quit"
	CmdVersion = "version"
	CmdHistory = "history"
	CmdClear   = "clear"
	// CmdInstall installs official plugins from GitHub Releases.
	CmdInstall = "install"
	// CmdUpdate updates installed official plugins from GitHub Releases.
	CmdUpdate = "update"
	// CmdUpgrade updates ait and installed official plugins from GitHub Releases.
	CmdUpgrade = "upgrade"

	// NameSSH is the REPL command name for aissh (matches binary).
	NameSSH = "aissh"
	// NameLog is the REPL command name for ailog (matches binary).
	NameLog = "ailog"
	// NameImg is the REPL command name for aiimg (matches binary).
	NameImg = "aiimg"
	// NameHost is the REPL command name for aihost (matches binary).
	NameHost = "aihost"

	pluginJSONName  = "plugin.json"
	pluginsDirName  = "plugins"
	defaultOrder    = 100
	manifestTimeout = 2 * time.Second
	flagTypeBool    = "bool"
	flagTypeString  = "string"
	flagTypeInt     = "int"
)
