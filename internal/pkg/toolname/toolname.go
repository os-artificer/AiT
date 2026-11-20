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

// Package toolname holds SSOT strings for AiT CLI tool names and the SSH meta prefix.
package toolname

import "os"

const (
	// AIT is the ait toolbox REPL binary name.
	AIT = "ait"

	// SSH, Log, Img, Host are ait plugin command names (ait ssh, ait log, …).
	SSH = "ssh"
	Log = "log"
	Img = "img"
	// Host is the ait host monitor plugin command name.
	Host = "host"
	// Keep is the ait keepalive plugin command name.
	Keep = "keep"
	// Vid is the ait video search plugin command name.
	Vid = "vid"

	// Agent is the ait agent admin subcommand (ait agent …).
	Agent = "agent"

	// LegacyAI is the retired standalone ai binary and SSH meta subcommand (!ait ai).
	LegacyAI = "ai"

	// AITSSH, AITLog, AITImg, AITHost are shipped binary basenames (bin/, PATH).
	AITSSH  = "aitssh"
	AITLog  = "aitlog"
	AITImg  = "aitimg"
	AITHost = "aithost"
	// AITKeep is the shipped keepalive binary basename.
	AITKeep = "aitkeep"
	// AITVid is the shipped video search binary basename.
	AITVid = "aitvid"

	// MetaPrefix is the in-session SSH meta opener without a trailing space (!ait).
	MetaPrefix = "!ait"

	// EnvPluginCMD is set by ait dispatch when exec'ing a plugin; value is the manifest command name (log, ssh, …).
	EnvPluginCMD = "AIT_PLUGIN_CMD"
)

// AgentCLI returns "ait agent" or "ait agent <subcmd>" when subcmd is non-empty.
func AgentCLI(subcmd string) string {
	if subcmd == "" {
		return AITCLI("agent")
	}
	return AITCLI("agent " + subcmd)
}

// AITCLI returns "ait" or "ait <subcmd>" when subcmd is non-empty.
func AITCLI(subcmd string) string {
	if subcmd == "" {
		return AIT
	}
	return AIT + " " + subcmd
}

// PluginCLI returns the user-visible CLI prefix for a plugin binary.
// When AIT_PLUGIN_CMD is set (ait dispatch), the command name is used; otherwise the binary name.
func PluginCLI(binary, command, subcmd string) string {
	prefix := binary
	if cmd := os.Getenv(EnvPluginCMD); cmd != "" {
		prefix = cmd
	}
	if subcmd == "" {
		return prefix
	}
	return prefix + " " + subcmd
}

// CLI returns the user-visible aissh prefix: "aitssh" or "ssh" when dispatched via ait, plus optional subcmd.
func CLI(subcmd string) string {
	return PluginCLI(AITSSH, SSH, subcmd)
}

// LogCLI returns the user-visible ailog prefix: "aitlog" or "log" when dispatched via ait, plus optional subcmd.
func LogCLI(subcmd string) string {
	return PluginCLI(AITLog, Log, subcmd)
}

// HostCLI returns the user-visible aihost prefix: "aithost" or "host" when dispatched via ait, plus optional subcmd.
func HostCLI(subcmd string) string {
	return PluginCLI(AITHost, Host, subcmd)
}

// ImgCLI returns the user-visible aiimg prefix: "aitimg" or "img" when dispatched via ait, plus optional subcmd.
func ImgCLI(subcmd string) string {
	return PluginCLI(AITImg, Img, subcmd)
}

// KeepCLI returns the user-visible keepalive prefix: "aitkeep" or "keep" when dispatched via ait, plus optional subcmd.
func KeepCLI(subcmd string) string {
	return PluginCLI(AITKeep, Keep, subcmd)
}

// VidCLI returns the user-visible aitvid prefix: "aitvid" or "vid" when dispatched via ait, plus optional subcmd.
func VidCLI(subcmd string) string {
	return PluginCLI(AITVid, Vid, subcmd)
}

// MetaLine returns the in-session SSH meta prefix "!ait" or "!ait <rest>" when rest is non-empty.
func MetaLine(rest string) string {
	if rest == "" {
		return MetaPrefix
	}
	return MetaPrefix + " " + rest
}
