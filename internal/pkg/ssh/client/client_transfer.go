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

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	xfer "github.com/os-artificer/AiT/internal/pkg/ssh/client/transfer"
	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

const transferUsageGet = xfer.UsageGet
const transferUsagePut = xfer.UsagePut

func (c *Client) handleTransferGet(ctx context.Context, stdin io.Writer, args []string) error {
	return c.handleTransfer(ctx, stdin, transfer.DirGet, transferUsageGet, args)
}

func (c *Client) handleTransferPut(ctx context.Context, stdin io.Writer, args []string) error {
	return c.handleTransfer(ctx, stdin, transfer.DirPut, transferUsagePut, args)
}

func (c *Client) handleTransfer(ctx context.Context, stdin io.Writer, dir transfer.Direction, usage string, args []string) error {
	if len(args) != 2 {
		err := transferUsageHintErr(usage)
		errhint.Print(os.Stderr, errhint.SurfaceTTY, err)
		c.finishMetaInteraction(stdin)
		return nil
	}

	usePTY, autoEnabled, _ := c.resolveTransferMode(stdin)
	prog := NewTransferProgress(os.Stderr)

	if usePTY && !autoEnabled {
		prog.ChainModeOn()
	}

	var err error
	if usePTY {
		err = c.handleTransferPTY(ctx, stdin, dir, args, prog)
		if err != nil && errors.Is(err, ErrInjectCaptureNoMarkers) {
			err = wrapPTYCaptureErr(err)
		} else if err != nil && errors.Is(err, ErrMissingLrzsz) {
			err = wrapPTYZModemErr(err)
		} else if err != nil && strings.Contains(err.Error(), "zmodem") {
			err = wrapPTYZModemErr(err)
		}
	} else {
		err = c.handleTransferSFTP(ctx, dir, args, prog)
		if err != nil && isSFTPNotFoundErr(err) {
			err = wrapTransferWrongHop(c.opts.Host, err)
		}
	}

	if err != nil {
		errhint.Print(os.Stderr, errhint.SurfaceTTY, err)
	}
	c.finishMetaInteraction(stdin)
	return nil
}

// resolveTransferMode decides SFTP vs PTY for this transfer.
// When chain is off but the active PTY shell is on a different host than c.opts.Host,
// enables chain mode for the session and returns usePTY=true.
func (c *Client) resolveTransferMode(stdin io.Writer) (usePTY bool, autoEnabled bool, err error) {
	if c.ChainMode() {
		return true, false, nil
	}
	if c.scroll == nil {
		return false, false, nil
	}

	activeHost, probeErr := c.probeActiveShellHost(stdin)
	if probeErr != nil {
		return false, false, nil
	}

	if activeShellNeedsChain(c.opts.Host, c.opts.LoginDestination, activeHost) {
		c.SetChainMode(true)
		autoChainNotice(activeHost, c.opts.Host)
		return true, true, nil
	}

	return false, false, nil
}

// handleTransferSFTP uses the direct SFTP connection for file transfer.
func (c *Client) handleTransferSFTP(ctx context.Context, dir transfer.Direction, args []string, prog *TransferProgress) error {
	var copyProgress transfer.CopyProgress
	if prog != nil {
		copyProgress = prog.SFTPProgress
	}

	if dir == transfer.DirGet {
		if prog != nil {
			prog.StartGet(args[0], args[1], -1)
		}
		err := transfer.GetWithProgress(ctx, c.conn, args[0], args[1], copyProgress)
		if err != nil {
			return err
		}
		if prog != nil {
			prog.Done(fmt.Sprintf("done: get %s → %s (%d bytes)", args[0], args[1], prog.Bytes()))
		}
		return nil
	}

	if prog != nil {
		prog.StartPut(args[0], args[1], -1)
	}
	err := transfer.PutWithProgress(ctx, c.conn, args[0], args[1], copyProgress)
	if err != nil {
		return err
	}
	if prog != nil {
		prog.Done(fmt.Sprintf("done: put %s → %s (%d bytes)", args[0], args[1], prog.Bytes()))
	}
	return nil
}

// handleTransferPTY uses PTY-pipe injection for file transfer when ChainMode is active.
func (c *Client) handleTransferPTY(ctx context.Context, stdin io.Writer, dir transfer.Direction, args []string, prog *TransferProgress) error {
	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	pt := NewZModemTransfer(c, prog)
	if dir == transfer.DirGet {
		return pt.Get(ctx, stdin, args[0], args[1])
	}
	return pt.Put(ctx, stdin, args[0], args[1])
}

func transferArgsFromMetaLine(line string) (cmd string, args []string, ok bool) {
	return xfer.ArgsFromMetaLine(line)
}

func transferUsageHintErr(usage string) error {
	return xfer.UsageHintErr(usage)
}
