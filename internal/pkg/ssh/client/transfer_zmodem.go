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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

const zmodemSettleDelay = 50 * time.Millisecond

// zmodemTransferWaitHook waits for ZMODEM completion (overridable in tests).
var zmodemTransferWaitHook = func(ctx context.Context, sess *zmodem.Session) error {
	return sess.Wait(ctx)
}

// ZModemTransfer moves files through the interactive PTY using lrzsz-compatible ZMODEM.
type ZModemTransfer struct {
	client   *Client
	progress *TransferProgress
}

// NewZModemTransfer returns a ZModemTransfer for chain-mode file transfer.
func NewZModemTransfer(c *Client, progress *TransferProgress) *ZModemTransfer {
	return &ZModemTransfer{client: c, progress: progress}
}

// Get downloads remotePath to localPath via remote sz + local ZMODEM receive.
func (t *ZModemTransfer) Get(ctx context.Context, stdin io.Writer, remotePath, localPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	localAbs, err := ptyAbsLocalPath(localPath)
	if err != nil {
		return err
	}

	if t.progress != nil {
		t.progress.StartGet(remotePath, localPath, -1)
		t.progress.PhaseInline("handshake…")
	}

	progressFn := t.zmodemProgressFn()

	_, cleanup, err := t.client.BeginZModemSession(true, stdin, localAbs, "", progressFn)
	if err != nil {
		return fmt.Errorf("zmodem get: %w", err)
	}
	defer cleanup()

	cmd := strings.TrimPrefix(buildSzCmd(remotePath), injectSpacePrefix)
	if err := t.client.InjectZModemCommand(stdin, cmd); err != nil {
		return fmt.Errorf("zmodem get: inject sz: %w", err)
	}

	if err := t.waitTransfer(ctx); err != nil {
		return fmt.Errorf("zmodem get: %w", err)
	}

	if t.progress != nil {
		if info, statErr := os.Stat(localAbs); statErr == nil {
			t.progress.SetBytes(info.Size())
		}
		t.progress.Done(formatTransferDone("get", remotePath, localPath, t.progress.Bytes()))
	}

	return nil
}

// Put uploads localPath to remotePath via remote rz + local ZMODEM send.
func (t *ZModemTransfer) Put(ctx context.Context, stdin io.Writer, localPath, remotePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	localAbs, err := ptyAbsLocalPath(localPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(localAbs)
	if err != nil {
		return fmt.Errorf("zmodem put: local read %q: %w", localPath, err)
	}

	fileSize := info.Size()
	if t.progress != nil {
		t.progress.StartPut(localPath, remotePath, fileSize)
		t.progress.PhaseInline("handshake…")
	}

	remoteName := remoteBaseName(remotePath)
	progressFn := t.zmodemProgressFn()

	_, cleanup, err := t.client.BeginZModemSession(false, stdin, localAbs, remoteName, progressFn)
	if err != nil {
		return fmt.Errorf("zmodem put: %w", err)
	}
	defer cleanup()

	cmd := strings.TrimPrefix(buildRzCmd(remotePath), injectSpacePrefix)
	if err := t.client.InjectZModemCommand(stdin, cmd); err != nil {
		return fmt.Errorf("zmodem put: inject rz: %w", err)
	}

	if err := t.waitTransfer(ctx); err != nil {
		return fmt.Errorf("zmodem put: %w", err)
	}

	if t.progress != nil {
		t.progress.Done(formatTransferDone("put", localPath, remotePath, info.Size()))
	}

	return nil
}

func (t *ZModemTransfer) zmodemProgressFn() zmodem.ProgressFunc {
	if t.progress == nil {
		return nil
	}
	return func(written, total int64) {
		t.progress.SFTPProgress(written, total)
	}
}

func (t *ZModemTransfer) waitTransfer(ctx context.Context) error {
	sess := t.client.zmodemSession
	if sess == nil {
		return errors.New("zmodem session not active")
	}

	sess.SetChainTeardown(t.client.ChainMode())

	if t.progress != nil {
		t.progress.PhaseFinishing()
	}

	time.Sleep(zmodemSettleDelay)
	if err := zmodemTransferWaitHook(ctx, sess); err != nil {
		return err
	}
	t.client.drainZModemPostWait()
	return nil
}
