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

package args

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

// TransferSpec holds parsed argv for ait get or ait put.
type TransferSpec struct {
	Conn       Spec
	Session    string
	LocalPath  string
	RemotePath string
}

// ParseTransfer parses argv for ait get or ait put (OpenSSH-compatible flags + AiT extensions).
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
			return TransferSpec{}, fmt.Errorf("-- is not supported for ait %s", cmdLabel)
		}
		if strings.HasPrefix(arg, "-") {
			if arg == "--session" {
				i++
				if i >= len(argv) {
					return TransferSpec{}, sessionFlagRequiresValueErr()
				}
				session = argv[i]
				continue
			}
			next, err := applyFlag(&s, argv, i, cmdLabel)
			if err != nil {
				return TransferSpec{}, err
			}
			i = next
			continue
		}
		positionals = append(positionals, arg)
	}

	if err := applyLegacy(&s); err != nil {
		return TransferSpec{}, err
	}

	local, remote, dest, err := classifyTransferPaths(positionals, session, cmdLabel)
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

func classifyTransferPaths(positionals []string, session, cmdLabel string) (local, remote, dest string, err error) {
	switch len(positionals) {
	case 2:
		return classifyTwoPaths(positionals, session)
	case 3:
		return classifyThreePaths(positionals)
	default:
		return "", "", "", transferPathCountErr(cmdLabel)
	}
}

func classifyTwoPaths(positionals []string, session string) (local, remote, dest string, err error) {
	a, b := positionals[0], positionals[1]

	if user, host, path, perr := transfer.ParseRemoteDest(a); perr == nil {
		return b, path, destinationFromUserHost(user, host), nil
	}
	if user, host, path, perr := transfer.ParseRemoteDest(b); perr == nil {
		return a, path, destinationFromUserHost(user, host), nil
	}

	if session == "" {
		return "", "", "", fmt.Errorf("remote path must be user@host:/path, or use three arguments (destination remote-path local-path), or --session")
	}

	aLocal := isLocalFilePath(a)
	bLocal := isLocalFilePath(b)
	switch {
	case aLocal && !bLocal:
		return a, b, "", nil
	case bLocal && !aLocal:
		return b, a, "", nil
	default:
		return "", "", "", fmt.Errorf("cannot tell local vs remote path with --session; use user@host:/path or a distinct local path (e.g. ./file)")
	}
}

func classifyThreePaths(positionals []string) (local, remote, dest string, err error) {
	destIdx := -1
	for i, p := range positionals {
		if !isDestinationToken(p) {
			continue
		}
		if destIdx >= 0 {
			return "", "", "", fmt.Errorf("ambiguous destination in arguments")
		}
		destIdx = i
	}
	if destIdx < 0 {
		return "", "", "", fmt.Errorf("three arguments require a destination (e.g. user@host or Host alias)")
	}

	dest = positionals[destIdx]
	var paths []string
	for i, p := range positionals {
		if i != destIdx {
			paths = append(paths, p)
		}
	}
	if len(paths) != 2 {
		return "", "", "", fmt.Errorf("expected local and remote file paths")
	}

	local, remote, err = pairLocalRemote(paths[0], paths[1])
	return local, remote, dest, err
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

func destinationFromUserHost(user, host string) string {
	if user != "" {
		return user + "@" + host
	}
	return host
}
