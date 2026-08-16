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

const (
	// AISsh is the aitssh Cobra root command name and installed binary name.
	AISsh = "aitssh"
	// AIImg is the aitimg Cobra root command name and installed binary name.
	AIImg = "aitimg"
	// AILog is the aitlog Cobra root command name and installed binary name.
	AILog = "aitlog"
	// AIHost is the aithost Cobra root command name and installed binary name.
	AIHost = "aithost"
	// MetaPrefix is the in-session SSH meta opener without a trailing space (!ait).
	MetaPrefix = "!ait"
)

// CLI returns "aitssh" or "aitssh <subcmd>" when subcmd is non-empty.
func CLI(subcmd string) string {
	if subcmd == "" {
		return AISsh
	}
	return AISsh + " " + subcmd
}

// MetaLine returns "!ait" or "!ait <rest>" when rest is non-empty.
func MetaLine(rest string) string {
	if rest == "" {
		return MetaPrefix
	}
	return MetaPrefix + " " + rest
}
