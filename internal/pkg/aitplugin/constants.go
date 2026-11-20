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
	CmdAgent:   {},
	CmdShell:   {},
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
	{Binary: toolname.AITSSH, Name: NameSSH},
	{Binary: toolname.AITLog, Name: NameLog},
	{Binary: toolname.AITImg, Name: NameImg},
	{Binary: toolname.AITHost, Name: NameHost},
	{Binary: toolname.AITKeep, Name: NameKeep},
	{Binary: toolname.AITVid, Name: NameVid},
}

// LegacyPluginNames maps retired command/binary tokens to current plugin command names.
var LegacyPluginNames = map[string]string{
	"aissh":          NameSSH,
	"ailog":          NameLog,
	"aiimg":          NameImg,
	"aihost":         NameHost,
	toolname.AITSSH:  NameSSH,
	toolname.AITLog:  NameLog,
	toolname.AITImg:  NameImg,
	toolname.AITHost: NameHost,
	toolname.AITKeep: NameKeep,
	toolname.AITVid:  NameVid,
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
	// CmdAgent is the ait reserved builtin for the v4 agent loop.
	CmdAgent = "agent"
	// CmdShell is the ait reserved builtin for the toolbox REPL.
	CmdShell = "shell"

	// NameSSH is the ait plugin command name for the aitssh binary.
	NameSSH = toolname.SSH
	// NameLog is the ait plugin command name for the aitlog binary.
	NameLog = toolname.Log
	// NameImg is the ait plugin command name for the aitimg binary.
	NameImg = toolname.Img
	// NameHost is the ait plugin command name for the aithost binary.
	NameHost = toolname.Host
	// NameKeep is the ait plugin command name for the aitkeep binary.
	NameKeep = toolname.Keep
	// NameVid is the ait plugin command name for the aitvid binary.
	NameVid = toolname.Vid

	pluginJSONName  = "plugin.json"
	pluginsDirName  = "plugins"
	defaultOrder    = 100
	manifestTimeout = 2 * time.Second
	flagTypeBool    = "bool"
	flagTypeString  = "string"
	flagTypeInt     = "int"
)
