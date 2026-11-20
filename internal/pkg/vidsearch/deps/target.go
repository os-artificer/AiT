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

package deps

import (
	"os"
	"runtime"
)

// Target identifies the local OS/package manager for dependency install.
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

// DetectTarget probes the local machine for a package manager.
// lookPath may be nil to use the default LookPathFunc.
func DetectTarget(lookPath LookPathFunc) Target {
	if lookPath == nil {
		lookPath = DefaultLookPath
	}
	switch runtime.GOOS {
	case "darwin":
		if _, err := lookPath("brew"); err == nil {
			return TargetDarwinBrew
		}
		return TargetDarwinNoBrew
	case "linux":
		for _, pair := range []struct {
			bin string
			t   Target
		}{
			{"apt-get", TargetLinuxApt},
			{"dnf", TargetLinuxDnf},
			{"yum", TargetLinuxYum},
			{"apk", TargetLinuxApk},
			{"pacman", TargetLinuxPacman},
		} {
			if _, err := lookPath(pair.bin); err == nil {
				return pair.t
			}
		}
		return TargetUnknown
	case "windows":
		if _, err := lookPath("choco"); err == nil {
			return TargetWindowsChoco
		}
		if _, err := lookPath("winget"); err == nil {
			return TargetWindowsWinget
		}
		return TargetWindowsNone
	default:
		return TargetUnknown
	}
}

// EnvCameraID is the optional camera id override environment variable.
const EnvCameraID = "AIT_VID_CAMERA_ID"

// CameraIDFromEnv returns AIT_VID_CAMERA_ID when set.
func CameraIDFromEnv() string {
	return os.Getenv(EnvCameraID)
}
