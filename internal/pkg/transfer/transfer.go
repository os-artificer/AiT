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

	"golang.org/x/crypto/ssh"
)

// Get downloads remotePath from conn to localPath (single regular file).
// ctx cancels the transfer. conn must be a live golang.org/x/crypto/ssh client.
// Errors: ctx.Err; SFTP/stat/open failures on remotePath; directory or non-regular remote paths; local write errors.
func Get(ctx context.Context, conn *ssh.Client, remotePath, localPath string) error {
	return transferFile(ctx, conn, DirGet, localPath, remotePath, nil, nil)
}

// Put uploads localPath to remotePath on conn (single regular file).
// Creates missing parent directories on the remote host before writing the file.
// ctx cancels the transfer. conn must be a live golang.org/x/crypto/ssh client.
// Errors: ctx.Err; local open/stat errors; SFTP mkdir/create/write errors; directory paths.
func Put(ctx context.Context, conn *ssh.Client, localPath, remotePath string) error {
	return transferFile(ctx, conn, DirPut, localPath, remotePath, nil, nil)
}

// GetWithProgress is like Get but reports copy progress via progress when non-nil.
func GetWithProgress(ctx context.Context, conn *ssh.Client, remotePath, localPath string, progress CopyProgress) error {
	return transferFile(ctx, conn, DirGet, localPath, remotePath, progress, nil)
}

// GetWithProgressClosing is like GetWithProgress and invokes onClosing after copy, before file close.
func GetWithProgressClosing(
	ctx context.Context, conn *ssh.Client, remotePath, localPath string,
	progress CopyProgress, onClosing func(),
) error {
	return transferFile(ctx, conn, DirGet, localPath, remotePath, progress, onClosing)
}

// PutWithProgress is like Put but reports copy progress via progress when non-nil.
func PutWithProgress(ctx context.Context, conn *ssh.Client, localPath, remotePath string, progress CopyProgress) error {
	return transferFile(ctx, conn, DirPut, localPath, remotePath, progress, nil)
}

// PutWithProgressClosing is like PutWithProgress and invokes onClosing after copy, before file close.
func PutWithProgressClosing(
	ctx context.Context, conn *ssh.Client, localPath, remotePath string,
	progress CopyProgress, onClosing func(),
) error {
	return transferFile(ctx, conn, DirPut, localPath, remotePath, progress, onClosing)
}
