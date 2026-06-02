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
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
)

const copyBufSize = 32 * 1024

// CopyProgress reports bytes copied during SFTP transfer. total is 0 when unknown.
type CopyProgress func(written, total int64)

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, total int64, progress CopyProgress) error {
	buf := make([]byte, copyBufSize)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if progress != nil {
				progress(written, total)
			}
		}
		if err == io.EOF {
			return nil
		}
	}
}

func absLocalPath(localPath string) (string, error) {
	if filepath.IsAbs(localPath) {
		return localPath, nil
	}
	return filepath.Abs(localPath)
}

func wrapLocalErr(localPath string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("local file %q: %w", localPath, err)
}

func wrapRemoteErr(remotePath string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf(
			"remote file %q: %w (file or parent directory may not exist; check the remote path)",
			remotePath, err,
		)
	}
	return fmt.Errorf("remote file %q: %w", remotePath, err)
}

func ensureRemoteParentDir(client *sftp.Client, remotePath string) error {
	dir := path.Dir(remotePath)
	if dir == "" || dir == "/" || dir == "." {
		return nil
	}
	if err := client.MkdirAll(dir); err != nil {
		return fmt.Errorf("remote directory %q: %w", dir, err)
	}
	return nil
}

func statRegularFile(client *sftp.Client, path string) (os.FileInfo, error) {
	info, err := client.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%q is a directory (only single files are supported)", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%q is not a regular file", path)
	}
	return info, nil
}
