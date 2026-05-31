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

// ProbeCommand returns a POSIX shell one-liner that prints ok when sz and rz are on PATH.
func ProbeCommand() string {
	return "command -v sz >/dev/null 2>&1 && command -v rz >/dev/null 2>&1 && echo " + probeOKToken
}

// ParseProbeOK reports whether capture output from ProbeCommand indicates lrzsz is present.
func ParseProbeOK(output string) bool {
	return strings.Contains(strings.TrimSpace(output), probeOKToken)
}
