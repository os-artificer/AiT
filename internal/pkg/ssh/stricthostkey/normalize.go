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

// Package stricthostkey normalizes OpenSSH StrictHostKeyChecking option values.
package stricthostkey

import "strings"

// NormalizeStrictHostKeyChecking maps OpenSSH StrictHostKeyChecking aliases to ask, yes, no, or accept-new.
// Empty or unknown values default to ask.
func NormalizeStrictHostKeyChecking(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "ask":
		return "ask"
	case "yes", "true", "on":
		return "yes"
	case "no", "false", "off":
		return "no"
	case "accept-new":
		return "accept-new"
	default:
		return "ask"
	}
}
