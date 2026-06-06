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
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

func localPathIsDirectory(p string) bool {
	if strings.HasSuffix(p, "/") {
		return true
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func remotePathIsDirectory(p string) bool {
	return strings.HasSuffix(p, "/")
}

func normalizeMetaTransferArgs(dir transfer.Direction, args []string) ([]string, error) {
	switch dir {
	case transfer.DirGet:
		switch len(args) {
		case 1:
			remote := args[0]
			return []string{remote, transfer.DefaultLocalPathForGet(remote)}, nil
		case 2:
			remote := args[0]
			local := transfer.LocalPathWithDirDefault(remote, args[1], localPathIsDirectory)
			return []string{remote, local}, nil
		default:
			return nil, transferUsageHintErr(transferUsageGet)
		}
	case transfer.DirPut:
		switch len(args) {
		case 1:
			local := args[0]
			return []string{local, transfer.DefaultRemotePathForPut(local)}, nil
		case 2:
			local := args[0]
			remote := transfer.RemotePathWithDirDefault(local, args[1], remotePathIsDirectory)
			return []string{local, remote}, nil
		default:
			return nil, transferUsageHintErr(transferUsagePut)
		}
	default:
		return nil, transferUsageHintErr(transferUsageGet)
	}
}

func resolveTransferArgs(c *Client, stdin io.Writer, dir transfer.Direction, args []string) ([]string, error) {
	resolved := make([]string, len(args))
	copy(resolved, args)

	remoteIdx := 0
	if dir == transfer.DirPut {
		remoteIdx = 1
	}

	remote, err := c.resolveRemotePathForTransfer(stdin, args[remoteIdx])
	if err != nil {
		return nil, err
	}
	resolved[remoteIdx] = remote

	return resolved, nil
}
