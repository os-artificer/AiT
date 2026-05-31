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

package lrzsz

import "strings"

// Target identifies the remote OS/package manager for lrzsz installation.
type Target string

const (
	TargetLinuxApt      Target = "linux-apt"
	TargetLinuxDnf      Target = "linux-dnf"
	TargetLinuxYum      Target = "linux-yum"
	TargetLinuxApk      Target = "linux-apk"
	TargetLinuxPacman   Target = "linux-pacman"
	TargetDarwinBrew    Target = "darwin-brew"
	TargetDarwinNoBrew  Target = "darwin-no-brew"
	TargetWindowsChoco  Target = "windows-choco"
	TargetWindowsWinget Target = "windows-winget"
	TargetWindowsNone   Target = "windows-none"
	TargetUnknown       Target = "unknown"
)

// DetectCommand returns a shell script that prints a Target token on stdout.
func DetectCommand() string {
	return `case "$(uname -s 2>/dev/null)" in
Darwin)
  command -v brew >/dev/null 2>&1 && echo darwin-brew || echo darwin-no-brew
  ;;
Linux)
  if command -v apt-get >/dev/null 2>&1; then echo linux-apt
  elif command -v dnf >/dev/null 2>&1; then echo linux-dnf
  elif command -v yum >/dev/null 2>&1; then echo linux-yum
  elif command -v apk >/dev/null 2>&1; then echo linux-apk
  elif command -v pacman >/dev/null 2>&1; then echo linux-pacman
  else echo unknown
  fi
  ;;
MINGW*|MSYS*|CYGWIN*)
  if command -v choco >/dev/null 2>&1; then echo windows-choco
  elif command -v winget >/dev/null 2>&1; then echo windows-winget
  else echo windows-none
  fi
  ;;
*)
  if command -v choco >/dev/null 2>&1; then echo windows-choco
  elif command -v winget >/dev/null 2>&1; then echo windows-winget
  else echo unknown
  fi
  ;;
esac`
}

// ParseTarget maps remote capture output to a Target.
func ParseTarget(output string) Target {
	line := strings.TrimSpace(output)
	if line == "" {
		return TargetUnknown
	}
	// Use last non-empty line when capture includes noise.
	for _, part := range strings.Split(line, "\n") {
		part = strings.TrimSpace(part)
		if part != "" {
			line = part
		}
	}
	return Target(line)
}

// InstallCommand returns the remote shell command to install lrzsz for target.
// Empty string means manual install is required.
func InstallCommand(target Target) string {
	switch target {
	case TargetLinuxApt:
		return "sudo apt-get update && sudo apt-get install -y lrzsz"
	case TargetLinuxDnf:
		return "sudo dnf install -y lrzsz"
	case TargetLinuxYum:
		return "sudo yum install -y lrzsz"
	case TargetLinuxApk:
		return "sudo apk add --no-cache lrzsz"
	case TargetLinuxPacman:
		return "sudo pacman -Sy --noconfirm lrzsz"
	case TargetDarwinBrew:
		return "brew install lrzsz"
	case TargetWindowsChoco:
		return "choco install lrzsz -y"
	case TargetWindowsWinget:
		return "winget install --id lrzsz.lrzsz -e --accept-package-agreements --accept-source-agreements"
	default:
		return ""
	}
}
