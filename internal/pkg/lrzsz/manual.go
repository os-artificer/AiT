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

// ManualTryLines returns errhint Try commands for manual lrzsz installation on target.
func ManualTryLines(target Target) []string {
	cmd := InstallCommand(target)
	if cmd != "" {
		return []string{cmd, "sudo -v   # if install needs a password"}
	}
	switch target {
	case TargetDarwinNoBrew:
		return []string{
			"brew install lrzsz",
			"Install Homebrew from https://brew.sh then: brew install lrzsz",
		}
	case TargetWindowsNone:
		return []string{
			"choco install lrzsz -y",
			"winget install lrzsz",
		}
	default:
		return []string{
			"apt install lrzsz / dnf install lrzsz / brew install lrzsz",
			"sudo -v   # if install needs a password",
		}
	}
}
