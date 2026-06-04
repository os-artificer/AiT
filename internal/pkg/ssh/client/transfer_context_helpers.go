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
	"sync"
)

const transferInterruptByte = 0x03 // Ctrl+C in raw TTY

// beginTransfer returns a cancellable context and a watcher that maps Ctrl+C (0x03) to cancel.
// The watcher reads stdinReader.Ch while forwardStdin is blocked inside handleTransfer.
func (c *Client) beginTransfer(parent context.Context) (context.Context, func()) {
	transferCtx, cancel := context.WithCancel(parent)

	c.transferMu.Lock()
	c.transferCancel = cancel
	c.transferMu.Unlock()
	c.transferActive.Store(true)

	stopWatch := make(chan struct{})
	var watchWG sync.WaitGroup
	if c.stdinReader != nil && !c.stdinReader.REPLActive() {
		watchWG.Add(1)
		go func() {
			defer watchWG.Done()
			c.watchTransferStdinInterrupt(transferCtx, stopWatch)
		}()
	}

	end := func() {
		close(stopWatch)
		watchWG.Wait()
		c.transferActive.Store(false)
		c.transferMu.Lock()
		c.transferCancel = nil
		c.transferMu.Unlock()
		cancel()
	}
	return transferCtx, end
}

// cancelTransfer cancels an in-flight transfer (Ctrl+C or explicit abort).
func (c *Client) cancelTransfer() {
	c.transferMu.Lock()
	cancel := c.transferCancel
	c.transferMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (c *Client) watchTransferStdinInterrupt(transferCtx context.Context, stop <-chan struct{}) {
	ch := c.stdinReader.Ch
	for {
		select {
		case <-stop:
			return
		case <-transferCtx.Done():
			return
		case sb, ok := <-ch:
			if !ok {
				return
			}
			if sb.Err != nil {
				continue
			}
			if sb.B != transferInterruptByte {
				continue
			}
			c.cancelTransfer()
			if sess := c.zmodemSession; sess != nil {
				sess.Abort()
			}
			return
		}
	}
}
