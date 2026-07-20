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

package version

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"
)

//go:embed VERSION
var embeddedVersion string

var (
	version   = ""
	gitCommit = ""
	buildDate = ""
)

// Version returns a human-readable build version string.
func Version() string {
	var b strings.Builder

	if v, ok := resolveVersion(); ok {
		fmt.Fprintf(&b, "Version:   %s\n", v)
	}
	if g, ok := resolveGitCommit(); ok {
		fmt.Fprintf(&b, "GitHash:   %s\n", g)
	}
	if d, ok := resolveBuildDate(); ok {
		fmt.Fprintf(&b, "Build:     %s\n", d)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// Short returns a compact version token for User-Agent (ldflags or embedded VERSION).
func Short() string {
	if v, ok := resolveVersion(); ok {
		return v
	}
	return "dev"
}

func resolveVersion() (string, bool) {
	if version != "" {
		return version, true
	}
	if v := strings.TrimSpace(embeddedVersion); v != "" {
		return v, true
	}
	return "", false
}

func resolveGitCommit() (string, bool) {
	if gitCommit != "" {
		return gitCommit, true
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) > 12 {
					return s.Value[:12], true
				}
				return s.Value, true
			}
		}
	}
	return "", false
}

func resolveBuildDate() (string, bool) {
	if buildDate != "" {
		return buildDate, true
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.time" {
				return s.Value, true
			}
		}
	}
	return "", false
}
