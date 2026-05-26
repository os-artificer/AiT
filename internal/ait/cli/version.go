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

// resolveVersion returns the semantic version string using a priority chain:
//  1. ldflags injection  (make build/install VERSION=vX.Y.Z)
//  2. embedded VERSION file
//  3. "dev" fallback
func resolveVersion() string {
	if version != "" {
		return version
	}
	if v := strings.TrimSpace(embeddedVersion); v != "" {
		return v
	}
	return "dev"
}

// resolveGitCommit returns the git commit hash using a priority chain:
//  1. ldflags injection
//  2. debug.ReadBuildInfo vcs.revision
//  3. "unknown" fallback
func resolveGitCommit() string {
	if gitCommit != "" {
		return gitCommit
	}
	if info, ok := readBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) > 12 {
					return s.Value[:12]
				}
				return s.Value
			}
		}
	}
	return "unknown"
}

// resolveBuildDate returns the build date using a priority chain:
//  1. ldflags injection
//  2. debug.ReadBuildInfo vcs.time
//  3. "unknown" fallback
func resolveBuildDate() string {
	if buildDate != "" {
		return buildDate
	}
	if info, ok := readBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.time" {
				return s.Value
			}
		}
	}
	return "unknown"
}

// readBuildInfo is a small wrapper for testability.
func readBuildInfo() (*debug.BuildInfo, bool) {
	info, ok := debug.ReadBuildInfo()
	return info, ok
}

// Version returns a human-readable build version string containing the
// semantic version, git commit hash, and build timestamp.
//
// Resolution priority per field:
//   - Version:   ldflags > embedded VERSION file > "dev"
//   - GitHash:   ldflags > debug.ReadBuildInfo > "unknown"
//   - Build:     ldflags > debug.ReadBuildInfo > "unknown"
func Version() string {
	return fmt.Sprintf("Version:   %s\nGitHash:   %s\nBuild:     %s",
		resolveVersion(), resolveGitCommit(), resolveBuildDate())
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
