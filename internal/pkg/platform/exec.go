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

package platform

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// LookPath wraps exec.LookPath.
func LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// OSDescription returns a short OS description for AI prompts.
func OSDescription() string {
	switch runtime.GOOS {
	case "darwin":
		if out, err := exec.Command("sw_vers").CombinedOutput(); err == nil {
			return strings.TrimSpace(string(out))
		}
		return "macOS"
	case "windows":
		if out, err := exec.Command("cmd", "/c", "ver").CombinedOutput(); err == nil {
			return strings.TrimSpace(string(out))
		}
		return "Windows"
	default:
		if b, err := os.ReadFile("/etc/os-release"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
				}
			}
		}
		return runtime.GOOS
	}
}
