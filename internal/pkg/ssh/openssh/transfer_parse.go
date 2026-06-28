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

package openssh

import (
	"fmt"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

// TransferOp selects get versus put path pairing semantics.
type TransferOp int

const (
	// OpGet downloads remotePath to localPath.
	OpGet TransferOp = iota
	// OpPut uploads localPath to remotePath.
	OpPut
)

// TransferSpec holds parsed argv for aitssh get or aitssh put.
type TransferSpec struct {
	Conn       Spec
	Session    string
	LocalPath  string
	RemotePath string
}

// ParseTransfer parses argv for aitssh get or aitssh put (OpenSSH-compatible flags + AiT extensions).
// Flags and paths may appear in any order. Returns an error for unknown flags, invalid path sets,
// or missing connection information when --session is not set.
func ParseTransfer(argv []string, op TransferOp) (TransferSpec, error) {
	cmdLabel := transferCmdLabel(op)
	var s Spec
	s.Options = make(map[string]string)
	s.LegacyPort = -1

	var session string
	var positionals []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			return TransferSpec{}, fmt.Errorf("-- is not supported for aitssh %s", cmdLabel)
		}
		if !strings.HasPrefix(arg, "-") {
			positionals = append(positionals, arg)
			continue
		}
		if arg == "--session" {
			var err error
			session, i, err = consumeSessionFlag(argv, i)
			if err != nil {
				return TransferSpec{}, err
			}
			continue
		}
		next, err := applyFlag(&s, argv, i, cmdLabel)
		if err != nil {
			return TransferSpec{}, err
		}
		i = next
	}

	if err := applyLegacy(&s); err != nil {
		return TransferSpec{}, err
	}

	local, remote, dest, err := classifyTransferPaths(positionals, session, cmdLabel, op)
	if err != nil {
		return TransferSpec{}, err
	}
	if dest != "" {
		s.Destination = dest
	}

	return TransferSpec{
		Conn:       s,
		Session:    session,
		LocalPath:  local,
		RemotePath: remote,
	}, nil
}

func transferCmdLabel(op TransferOp) string {
	if op == OpPut {
		return "put"
	}
	return "get"
}

func classifyTransferPaths(positionals []string, session, cmdLabel string, op TransferOp) (local, remote, dest string, err error) {
	switch len(positionals) {
	case 1:
		if op == OpGet {
			return classifyOneGetPath(positionals[0], session)
		}
		if op == OpPut {
			return classifyOnePutPath(positionals[0], session)
		}
		return "", "", "", transferPathCountErr(cmdLabel)
	case 2:
		return classifyTwoPaths(positionals, session, op)
	case 3:
		return classifyThreePaths(positionals, op)
	default:
		return "", "", "", transferPathCountErr(cmdLabel)
	}
}

func classifyOneGetPath(remoteArg, session string) (local, remote, dest string, err error) {
	if user, host, path, perr := transfer.ParseRemoteDest(remoteArg); perr == nil {
		return transfer.DefaultLocalPathForGet(path), path, destinationFromUserHost(user, host), nil
	}
	if session == "" {
		return "", "", "", fmt.Errorf("remote path must be user@host:/path, or use two arguments (remote-path local-path), or --session")
	}
	return transfer.DefaultLocalPathForGet(remoteArg), remoteArg, "", nil
}

func classifyOnePutPath(arg, session string) (local, remote, dest string, err error) {
	if user, host, path, perr := transfer.ParseRemoteDest(arg); perr == nil {
		return transfer.DefaultLocalPathForGet(path), path, destinationFromUserHost(user, host), nil
	}
	if session == "" {
		return "", "", "", fmt.Errorf("local path must be user@host:/path, or use two arguments (local-path remote-path), or --session")
	}
	return arg, transfer.DefaultRemotePathForPut(arg), "", nil
}

func classifyTwoPaths(positionals []string, session string, op TransferOp) (local, remote, dest string, err error) {
	a, b := positionals[0], positionals[1]

	if user, host, path, perr := transfer.ParseRemoteDest(a); perr == nil {
		local = b
		if op == OpGet {
			local = transfer.LocalPathWithDirDefault(path, b, isLocalDirectoryPath)
		} else {
			path = transfer.RemotePathWithDirDefault(b, path, isRemoteDirectoryPath)
		}
		return local, path, destinationFromUserHost(user, host), nil
	}
	if user, host, path, perr := transfer.ParseRemoteDest(b); perr == nil {
		local = a
		if op == OpGet {
			local = transfer.LocalPathWithDirDefault(path, a, isLocalDirectoryPath)
		} else {
			path = transfer.RemotePathWithDirDefault(a, path, isRemoteDirectoryPath)
		}
		return local, path, destinationFromUserHost(user, host), nil
	}

	if session == "" {
		return "", "", "", fmt.Errorf("remote path must be user@host:/path, or use three arguments (destination remote-path local-path), or --session")
	}

	aLocal := isLocalFilePath(a)
	bLocal := isLocalFilePath(b)
	switch {
	case aLocal && !bLocal:
		local := a
		remote := b
		if op == OpGet {
			local = transfer.LocalPathWithDirDefault(b, a, isLocalDirectoryPath)
		} else {
			remote = transfer.RemotePathWithDirDefault(a, b, isRemoteDirectoryPath)
		}
		return local, remote, "", nil
	case bLocal && !aLocal:
		local := b
		remote := a
		if op == OpGet {
			local = transfer.LocalPathWithDirDefault(a, b, isLocalDirectoryPath)
		} else {
			remote = transfer.RemotePathWithDirDefault(b, a, isRemoteDirectoryPath)
		}
		return local, remote, "", nil
	default:
		return "", "", "", fmt.Errorf("cannot tell local vs remote path with --session; use user@host:/path or a distinct local path (e.g. ./file)")
	}
}

func classifyThreePaths(positionals []string, op TransferOp) (local, remote, dest string, err error) {
	destIdx, ambiguous := findDestinationIndex(positionals)
	if ambiguous {
		return "", "", "", fmt.Errorf("ambiguous destination in arguments")
	}
	if destIdx < 0 {
		return "", "", "", fmt.Errorf("three arguments require a destination (e.g. user@host or Host alias)")
	}

	dest = positionals[destIdx]
	paths := pathsExcluding(positionals, destIdx)
	if len(paths) != 2 {
		return "", "", "", fmt.Errorf("expected local and remote file paths")
	}

	local, remote, err = pairLocalRemote(paths[0], paths[1])
	if err != nil {
		return "", "", "", err
	}
	if op == OpGet {
		local = transfer.LocalPathWithDirDefault(remote, local, isLocalDirectoryPath)
	} else {
		remote = transfer.RemotePathWithDirDefault(local, remote, isRemoteDirectoryPath)
	}
	return local, remote, dest, nil
}

func findDestinationIndex(positionals []string) (idx int, ambiguous bool) {
	idx = -1
	for i, p := range positionals {
		if !isDestinationToken(p) {
			continue
		}
		if idx >= 0 {
			return -1, true
		}
		idx = i
	}
	return idx, false
}

func pathsExcluding(positionals []string, excludeIdx int) []string {
	var paths []string
	for i, p := range positionals {
		if i != excludeIdx {
			paths = append(paths, p)
		}
	}
	return paths
}

func pairLocalRemote(a, b string) (local, remote string, err error) {
	aLocal := isLocalFilePath(a)
	bLocal := isLocalFilePath(b)
	switch {
	case aLocal && !bLocal:
		return a, b, nil
	case bLocal && !aLocal:
		return b, a, nil
	case strings.HasPrefix(a, "/") && !strings.HasPrefix(b, "/"):
		return b, a, nil
	case strings.HasPrefix(b, "/") && !strings.HasPrefix(a, "/"):
		return a, b, nil
	default:
		return "", "", fmt.Errorf("cannot tell local vs remote path; use ./local or an absolute remote path")
	}
}

func isDestinationToken(p string) bool {
	if strings.Contains(p, ":/") || strings.Contains(p, "]:") {
		return false
	}
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
		return false
	}
	if strings.Contains(p, "@") {
		_, h, _, err := ParseDestination(p)
		return err == nil && h != ""
	}
	if strings.HasPrefix(p, "/") {
		return false
	}
	_, h, _, err := ParseDestination(p)
	return err == nil && h != "" && !strings.Contains(p, "/")
}

func isLocalFilePath(p string) bool {
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
		return true
	}
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}

func isLocalDirectoryPath(p string) bool {
	if strings.HasSuffix(p, "/") {
		return true
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func isRemoteDirectoryPath(p string) bool {
	return strings.HasSuffix(p, "/")
}

func destinationFromUserHost(user, host string) string {
	if user != "" {
		return user + "@" + host
	}
	return host
}
