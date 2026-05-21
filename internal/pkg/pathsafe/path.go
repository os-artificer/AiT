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

// Package pathsafe validates strings that will be embedded in remote shell commands.
package pathsafe

import (
	"regexp"
	"strings"
)

// remotePathRe allows absolute paths safe for POSIX cd (no metacharacters).
var remotePathRe = regexp.MustCompile(`^[A-Za-z0-9_./~-]+$`)

// RemotePath reports whether p is safe to embed in a remote cd command: absolute path,
// no "..", and only [A-Za-z0-9_./~-] characters.
func RemotePath(p string) bool {
	if p == "" || !strings.HasPrefix(p, "/") {
		return false
	}

	if strings.Contains(p, "..") {
		return false
	}
	return remotePathRe.MatchString(p)
}

// QuoteShellSingle returns s wrapped in POSIX single quotes, escaping embedded apostrophes.
// Safe for embedding in remote sh -c style command lines.
func QuoteShellSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
