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
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Get downloads remotePath from conn to localPath (single regular file).
// ctx cancels the transfer. conn must be a live golang.org/x/crypto/ssh client.
// Errors: ctx.Err; SFTP/stat/open failures on remotePath; directory or non-regular remote paths; local write errors.
func Get(ctx context.Context, conn *ssh.Client, remotePath, localPath string) error {
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

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if _, err := statRegularFile(sftpClient, remotePath); err != nil {
		return wrapRemoteErr(remotePath, err)
	}

	remote, err := sftpClient.Open(remotePath)
	if err != nil {
		return wrapRemoteErr(remotePath, err)
	}
	defer remote.Close()

	local, err := os.Create(localAbs)
	if err != nil {
		return wrapLocalErr(localAbs, err)
	}
	defer func() {
		_ = local.Close()
	}()

	if err := copyWithContext(ctx, local, remote); err != nil {
		_ = os.Remove(localAbs)
		return wrapLocalErr(localAbs, err)
	}

	return wrapLocalErr(localAbs, local.Close())
}

// Put uploads localPath to remotePath on conn (single regular file).
// Creates missing parent directories on the remote host before writing the file.
// ctx cancels the transfer. conn must be a live golang.org/x/crypto/ssh client.
// Errors: ctx.Err; local open/stat errors; SFTP mkdir/create/write errors; directory paths.
func Put(ctx context.Context, conn *ssh.Client, localPath, remotePath string) error {
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

	local, err := os.Open(localAbs)
	if err != nil {
		return wrapLocalErr(localAbs, err)
	}
	defer local.Close()

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if err := ensureRemoteParentDir(sftpClient, remotePath); err != nil {
		return err
	}

	remote, err := sftpClient.Create(remotePath)
	if err != nil {
		return wrapRemoteErr(remotePath, err)
	}
	defer remote.Close()

	if err := copyWithContext(ctx, remote, local); err != nil {
		return wrapRemoteErr(remotePath, err)
	}

	return wrapRemoteErr(remotePath, remote.Close())
}
