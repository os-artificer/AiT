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

package client

import (
	"path/filepath"
	"strings"
)

// shellQuoteArg wraps s in single quotes with embedded single-quote escaping.
func shellQuoteArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// ptyAbsLocalPath returns the absolute path for localPath.
func ptyAbsLocalPath(localPath string) (string, error) {
	if filepath.IsAbs(localPath) {
		return localPath, nil
	}
	return filepath.Abs(localPath)
}

// remoteBaseName returns the final path element for ZMODEM file headers.
func remoteBaseName(remotePath string) string {
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		return "file"
	}
	base := filepath.Base(remotePath)
	if base == "." || base == "/" || base == "" {
		return filepath.Base(strings.TrimRight(remotePath, "/"))
	}
	return base
}

// buildSzCmd returns a space-prefixed remote sz command for ZMODEM download (get).
func buildSzCmd(remotePath string) string {
	return injectSpacePrefix + "sz -y " + shellQuoteArg(remotePath)
}

// buildRzCmd returns a space-prefixed remote rz command for ZMODEM upload (put).
func buildRzCmd(remotePath string) string {
	dir := filepath.Dir(remotePath)
	if dir == "." || dir == "" {
		return injectSpacePrefix + "rz -y"
	}
	return injectSpacePrefix + "cd " + shellQuoteArg(dir) + " && rz -y"
}
