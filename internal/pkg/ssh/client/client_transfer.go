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
	"io"
	"os"

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

	var err error
	if dir == transfer.DirGet {
		err = transfer.Get(ctx, c.conn, args[0], args[1])
	} else {
		err = transfer.Put(ctx, c.conn, args[0], args[1])
	}
	if err != nil {
		errhint.Print(os.Stderr, errhint.SurfaceTTY, err)
	}
	c.finishMetaInteraction(stdin)
	return nil
}

func transferArgsFromMetaLine(line string) (cmd string, args []string, ok bool) {
	return xfer.ArgsFromMetaLine(line)
}

func transferUsageHintErr(usage string) error {
	return xfer.UsageHintErr(usage)
}
