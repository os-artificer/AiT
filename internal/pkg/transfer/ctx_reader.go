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
	"io"
)

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func newCtxReader(ctx context.Context, r io.Reader) io.Reader {
	if ctx == nil || r == nil {
		return r
	}
	return &ctxReader{ctx: ctx, r: r}
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}

	type readResult struct {
		n   int
		err error
	}
	ch := make(chan readResult, 1)
	go func() {
		n, err := c.r.Read(p)
		ch <- readResult{n: n, err: err}
	}()

	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	case res := <-ch:
		return res.n, res.err
	}
}
