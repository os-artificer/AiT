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

package transfer

import (
	"path"
	"path/filepath"
	"strings"
)

const defaultLocalFileName = "file"

// RemoteBaseName returns the final path element of a remote file path.
func RemoteBaseName(remotePath string) string {
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		return defaultLocalFileName
	}

	base := path.Base(remotePath)
	if base == "." || base == "/" || base == "" {
		base = path.Base(strings.TrimRight(remotePath, "/"))
	}
	if base == "" || base == "." {
		return defaultLocalFileName
	}

	return base
}

// DefaultLocalPathForGet returns the default local destination for a get when omitted.
func DefaultLocalPathForGet(remote string) string {
	return "./" + RemoteBaseName(remote)
}

// LocalPathWithDirDefault appends the remote basename when local names an existing directory.
func LocalPathWithDirDefault(remote, local string, localIsDir func(string) bool) string {
	if localIsDir == nil || !localIsDir(local) {
		return local
	}

	return filepath.Join(local, RemoteBaseName(remote))
}

// LocalBaseName returns the final path element of a local file path.
func LocalBaseName(localPath string) string {
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return defaultLocalFileName
	}

	base := filepath.Base(localPath)
	if base == "." || base == string(filepath.Separator) || base == "" {
		base = filepath.Base(strings.TrimRight(localPath, string(filepath.Separator)))
	}
	if base == "" || base == "." {
		return defaultLocalFileName
	}

	return base
}

// DefaultRemotePathForPut returns the default remote destination for a put when omitted.
func DefaultRemotePathForPut(local string) string {
	return "./" + LocalBaseName(local)
}

// RemotePathWithDirDefault appends the local basename when remote names a directory.
func RemotePathWithDirDefault(local, remote string, remoteIsDir func(string) bool) string {
	if remoteIsDir == nil || !remoteIsDir(remote) {
		return remote
	}

	joined := path.Join(remote, LocalBaseName(local))
	if strings.HasPrefix(remote, "./") && !strings.HasPrefix(joined, "./") {
		return "./" + joined
	}

	return joined
}
