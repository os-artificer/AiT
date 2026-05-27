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

package cli

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var embeddedVersion string

// Build metadata injected via -ldflags at make build/install time.
// When ldflags are not provided (e.g. plain go install), the code falls
// back to the embedded VERSION file and runtime/debug.ReadBuildInfo().
var (
	version   = ""
	gitCommit = ""
	buildDate = ""
)

// resolveVersion returns the semantic version string using a priority chain.
// The bool reports whether the value came from a real source (ldflags or
// embedded VERSION file) rather than a fallback.
//
// Priority:
//  1. ldflags injection  (make build/install)
//  2. embedded VERSION file
//  3. "" fallback (not real)
func resolveVersion() (string, bool) {
	if version != "" {
		return version, true
	}
	if v := strings.TrimSpace(embeddedVersion); v != "" {
		return v, true
	}
	return "", false
}

// resolveGitCommit returns the git commit hash using a priority chain.
// The bool reports whether the value came from a real source rather than a fallback.
//
// Priority:
//  1. ldflags injection
//  2. debug.ReadBuildInfo vcs.revision
//  3. "" fallback (not real)
func resolveGitCommit() (string, bool) {
	if gitCommit != "" {
		return gitCommit, true
	}
	if info, ok := readBuildInfo(); ok {
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

// resolveBuildDate returns the build date using a priority chain.
// The bool reports whether the value came from a real source rather than a fallback.
//
// Priority:
//  1. ldflags injection
//  2. debug.ReadBuildInfo vcs.time
//  3. "" fallback (not real)
func resolveBuildDate() (string, bool) {
	if buildDate != "" {
		return buildDate, true
	}
	if info, ok := readBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.time" {
				return s.Value, true
			}
		}
	}
	return "", false
}

// readBuildInfo is a small wrapper for testability.
func readBuildInfo() (*debug.BuildInfo, bool) {
	info, ok := debug.ReadBuildInfo()
	return info, ok
}

// Version returns a human-readable build version string.
// Fields are only included when resolved from a real source
// (ldflags, embedded VERSION file, or debug.BuildInfo);
// fallback values are omitted.
func Version() string {
	var b strings.Builder

	v, ok := resolveVersion()
	if ok {
		fmt.Fprintf(&b, "Version:   %s\n", v)
	}

	g, ok := resolveGitCommit()
	if ok {
		fmt.Fprintf(&b, "GitHash:   %s\n", g)
	}

	d, ok := resolveBuildDate()
	if ok {
		fmt.Fprintf(&b, "Build:     %s\n", d)
	}

	// Trim trailing newline so the caller (newVersionCmd) uses Fprintln.
	return strings.TrimSuffix(b.String(), "\n")
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print build version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), Version())
			return err
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteVersionHelp)
	return cmd
}
