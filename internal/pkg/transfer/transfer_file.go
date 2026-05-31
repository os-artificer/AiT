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
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Direction selects download (Get) or upload (Put) for transferFile.
type Direction int

const (
	// DirGet downloads remote to local.
	DirGet Direction = iota
	// DirPut uploads local to remote.
	DirPut
)

func transferFile(ctx context.Context, conn *ssh.Client, dir Direction, localPath, remotePath string, progress CopyProgress) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if conn == nil {
		return fmt.Errorf("SSH connection is not available")
	}

	localAbs, err := absLocalPath(localPath)
	if err != nil {
		return wrapLocalErr(localPath, err)
	}

	var totalSize int64
	if dir == DirPut {
		if err := validateLocalRegularFile(localAbs); err != nil {
			return err
		}
		if info, statErr := os.Stat(localAbs); statErr == nil {
			totalSize = info.Size()
		}
	}

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	src, dst, cleanup, remoteSize, err := openTransferEnds(ctx, sftpClient, dir, localAbs, remotePath)
	if err != nil {
		return err
	}
	defer cleanup()

	copyTotal := totalSize
	if dir == DirGet && remoteSize > 0 {
		copyTotal = remoteSize
	}

	if err := copyWithContext(ctx, dst, src, copyTotal, progress); err != nil {
		if dir == DirGet {
			_ = os.Remove(localAbs)
			return wrapLocalErr(localAbs, err)
		}
		return wrapRemoteErr(remotePath, err)
	}

	return finishTransfer(dir, localAbs, remotePath, dst)
}

func validateLocalRegularFile(localAbs string) error {
	info, err := os.Stat(localAbs)
	if err != nil {
		return wrapLocalErr(localAbs, err)
	}
	if info.IsDir() {
		return fmt.Errorf("local file %q is a directory (only single files are supported)", localAbs)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("local file %q is not a regular file", localAbs)
	}
	return nil
}

func openTransferEnds(
	ctx context.Context,
	sftpClient *sftp.Client,
	dir Direction,
	localAbs, remotePath string,
) (src io.Reader, dst io.Writer, cleanup func(), fileSize int64, err error) {
	_ = ctx
	if dir == DirGet {
		info, statErr := statRegularFile(sftpClient, remotePath)
		if statErr != nil {
			return nil, nil, nil, 0, wrapRemoteErr(remotePath, statErr)
		}

		remote, err := sftpClient.Open(remotePath)
		if err != nil {
			return nil, nil, nil, 0, wrapRemoteErr(remotePath, err)
		}

		local, err := os.Create(localAbs)
		if err != nil {
			_ = remote.Close()
			return nil, nil, nil, 0, wrapLocalErr(localAbs, err)
		}

		return remote, local, func() {
			_ = remote.Close()
			_ = local.Close()
		}, info.Size(), nil
	}

	if err := ensureRemoteParentDir(sftpClient, remotePath); err != nil {
		return nil, nil, nil, 0, err
	}

	local, err := os.Open(localAbs)
	if err != nil {
		return nil, nil, nil, 0, wrapLocalErr(localAbs, err)
	}

	remote, err := sftpClient.Create(remotePath)
	if err != nil {
		_ = local.Close()
		return nil, nil, nil, 0, wrapRemoteErr(remotePath, err)
	}

	info, statErr := local.Stat()
	var size int64
	if statErr == nil {
		size = info.Size()
	}

	return local, remote, func() {
		_ = local.Close()
		_ = remote.Close()
	}, size, nil
}

func finishTransfer(dir Direction, localAbs, remotePath string, dst io.Writer) error {
	if dir == DirGet {
		if c, ok := dst.(io.Closer); ok {
			return wrapLocalErr(localAbs, c.Close())
		}
		return nil
	}
	if c, ok := dst.(io.Closer); ok {
		return wrapRemoteErr(remotePath, c.Close())
	}
	return nil
}
