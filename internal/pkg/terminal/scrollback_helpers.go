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

package terminal

// trimBufferForAppend drops oldest bytes from s so len(s)+add fits within max.
func trimBufferForAppend(s string, add, max int) string {
	if len(s)+add <= max {
		return s
	}
	keep := max - add
	if keep >= len(s) {
		return s
	}
	return trimUTF8Tail(s, keep)
}

// lastNLinesFrom returns the last n newline-delimited lines in s[:end].
func lastNLinesFrom(s string, end, n int) string {
	count := 0
	i := end
	for i > 0 {
		if s[i-1] != '\n' {
			i--
			continue
		}
		count++
		if count == n {
			return s[i:end]
		}
		i--
	}
	return s[:end]
}
