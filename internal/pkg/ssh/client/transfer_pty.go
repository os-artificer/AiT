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
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

// PTYTransfer transfers files through the interactive PTY session using base64
// encoding, supporting any number of SSH hops. All injected commands are
// space-prefixed for ignorespace history suppression and wrapped with
// stty -echo/echo so they are invisible to the user.
type PTYTransfer struct {
	client   *Client
	progress *TransferProgress
}

// NewPTYTransfer returns a PTYTransfer that injects commands through c's PTY pipe.
// progress may be nil. c must have an active interactive session.
func NewPTYTransfer(c *Client, progress *TransferProgress) *PTYTransfer {
	return &PTYTransfer{client: c, progress: progress}
}

// Get downloads remotePath to localPath through the PTY pipe using base64.
// ctx cancels the transfer. The remote host must have base64 and sha256sum available.
// Returns ctx.Err on cancellation, base64/decode errors, hash mismatch, or local write errors.
func (t *PTYTransfer) Get(ctx context.Context, stdin io.Writer, remotePath, localPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	localAbs, err := ptyAbsLocalPath(localPath)
	if err != nil {
		return err
	}

	if t.progress != nil {
		t.progress.StartGet(remotePath, localPath, -1)
	}

	if t.progress != nil {
		t.progress.Phase("hashing…")
	}
	remoteHash, err := t.captureRemoteHash(stdin, remotePath)
	if err != nil {
		return fmt.Errorf("pty get: remote hash: %w", err)
	}

	if t.progress != nil {
		t.progress.Phase("downloading…")
	}
	remoteBase64, err := t.captureRemoteBase64(stdin, remotePath)
	if err != nil {
		return fmt.Errorf("pty get: remote base64: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(remoteBase64)
	if err != nil {
		return fmt.Errorf("pty get: base64 decode: %w", err)
	}

	if err := os.WriteFile(localAbs, data, 0o600); err != nil {
		return fmt.Errorf("pty get: local write %q: %w", localAbs, err)
	}

	if remoteHash != "" {
		if t.progress != nil {
			t.progress.Phase("verifying…")
		}
		if err := verifyLocalHash(localAbs, remoteHash); err != nil {
			_ = os.Remove(localAbs)
			return fmt.Errorf("pty get: hash mismatch: %w", err)
		}
	}

	if t.progress != nil {
		t.progress.SetBytes(int64(len(data)))
		t.progress.Done(fmt.Sprintf("done: get %s → %s (%d bytes)", remotePath, localPath, len(data)))
	}

	return nil
}

// Put uploads localPath to remotePath through the PTY pipe using base64.
// ctx cancels the transfer. The remote host must have base64 and sha256sum available.
// Returns ctx.Err on cancellation, local read errors, or remote write verification errors.
func (t *PTYTransfer) Put(ctx context.Context, stdin io.Writer, localPath, remotePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("pty put: local read %q: %w", localPath, err)
	}

	if t.progress != nil {
		t.progress.StartPut(localPath, remotePath, int64(len(data)))
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	if err := t.writeRemoteBase64(stdin, remotePath, encoded); err != nil {
		return fmt.Errorf("pty put: remote write: %w", err)
	}

	localHash := localFileHash(data)

	if t.progress != nil {
		t.progress.Phase("verifying…")
	}
	remoteHash, err := t.captureRemoteHash(stdin, remotePath)
	if err != nil {
		return fmt.Errorf("pty put: remote verify: %w", err)
	}

	if remoteHash != "" && remoteHash != localHash {
		return fmt.Errorf("pty put: hash mismatch: local=%s remote=%s", localHash, remoteHash)
	}

	if t.progress != nil {
		t.progress.Done(fmt.Sprintf("done: put %s → %s (%d bytes)", localPath, remotePath, len(data)))
	}

	return nil
}

// captureRemoteHash runs sha256sum on remotePath via the PTY pipe and returns the hash.
func (t *PTYTransfer) captureRemoteHash(stdin io.Writer, remotePath string) (string, error) {
	output, err := t.client.injectCaptureRetry(stdin, "sha256sum "+shellQuoteArg(remotePath))
	if err != nil {
		return "", err
	}

	return parseHashFromOutput(output), nil
}

// captureRemoteBase64 runs base64 on remotePath via the PTY pipe and returns the encoded output.
func (t *PTYTransfer) captureRemoteBase64(stdin io.Writer, remotePath string) (string, error) {
	output, err := t.client.injectCaptureRetry(stdin, "base64 "+shellQuoteArg(remotePath))
	if err != nil {
		return "", err
	}

	cleaned := strings.ReplaceAll(output, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	return cleaned, nil
}

// writeRemoteBase64 uploads base64-encoded data to remotePath in chunks.
// Each chunk is decoded on the remote side via printf | base64 -d >>.
func (t *PTYTransfer) writeRemoteBase64(stdin io.Writer, remotePath, encoded string) error {
	if err := ptyTransferInjectHook(t.client, stdin, "> "+shellQuoteArg(remotePath), InjectOpts{SuppressEcho: true}); err != nil {
		return err
	}

	chunks := chunkString(encoded, ptyTransferChunkSize)
	for i, chunk := range chunks {
		if t.progress != nil {
			t.progress.Phase(fmt.Sprintf("uploading chunk %d/%d…", i+1, len(chunks)))
		}
		cmd := buildPutChunkCmd(remotePath, chunk)
		if err := ptyTransferInjectHook(t.client, stdin, cmd, InjectOpts{SuppressEcho: true}); err != nil {
			return err
		}
	}

	if t.client.scroll != nil {
		t.client.scroll.StripSubstrings(injectStripFragments()...)
	}

	return ptyTransferSettleHook(t.client, stdin)
}
