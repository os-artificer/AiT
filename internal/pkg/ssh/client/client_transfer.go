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
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

const transferUsageGet = "usage: !ait get <remote-path> <local-path>"
const transferUsagePut = "usage: !ait put <local-path> <remote-path>"

func (c *Client) handleTransferGet(ctx context.Context, stdin io.Writer, args []string) error {
	if len(args) != 2 {
		termPrintln(os.Stderr, "AiT: "+transferUsageGet)
		c.finishMetaInteraction(stdin)
		return nil
	}

	err := transfer.Get(ctx, c.conn, args[0], args[1])
	if err != nil {
		termPrintf(os.Stderr, "AiT: %v", err)
	}
	c.finishMetaInteraction(stdin)
	return nil
}

func (c *Client) handleTransferPut(ctx context.Context, stdin io.Writer, args []string) error {
	if len(args) != 2 {
		termPrintln(os.Stderr, "AiT: "+transferUsagePut)
		c.finishMetaInteraction(stdin)
		return nil
	}

	err := transfer.Put(ctx, c.conn, args[0], args[1])
	if err != nil {
		termPrintf(os.Stderr, "AiT: %v", err)
	}
	c.finishMetaInteraction(stdin)
	return nil
}

func transferArgsFromMetaLine(line string) (cmd string, args []string, ok bool) {
	if !strings.HasPrefix(line, "!ait ") {
		return "", nil, false
	}
	parts := strings.Fields(strings.TrimPrefix(line, "!ait "))
	if len(parts) == 0 {
		return "", nil, false
	}
	switch parts[0] {
	case "get", "put":
		return parts[0], parts[1:], true
	default:
		return "", nil, false
	}
}
