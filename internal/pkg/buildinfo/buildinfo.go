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

// Package buildinfo formats tool version strings from ldflags / embed / VCS.
package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Inputs holds optional version metadata.
type Inputs struct {
	Version   string
	GitCommit string
	BuildDate string
	Embedded  string
}

// Format returns a human-readable multi-line version string.
// Fields are only included when resolved from a real source.
func Format(in Inputs) string {
	var b strings.Builder
	if v, ok := resolveVersion(in); ok {
		fmt.Fprintf(&b, "Version:   %s\n", v)
	}
	if g, ok := resolveGitCommit(in); ok {
		fmt.Fprintf(&b, "GitHash:   %s\n", g)
	}
	if d, ok := resolveBuildDate(in); ok {
		fmt.Fprintf(&b, "Build:     %s\n", d)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// Short returns a compact version token for User-Agent (ldflags or embedded).
func Short(in Inputs) string {
	if v, ok := resolveVersion(in); ok {
		return v
	}
	return "dev"
}

func resolveVersion(in Inputs) (string, bool) {
	if in.Version != "" {
		return in.Version, true
	}
	if v := strings.TrimSpace(in.Embedded); v != "" {
		return v, true
	}
	return "", false
}

func resolveGitCommit(in Inputs) (string, bool) {
	if in.GitCommit != "" {
		return in.GitCommit, true
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

func resolveBuildDate(in Inputs) (string, bool) {
	if in.BuildDate != "" {
		return in.BuildDate, true
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
