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

package pager

import (
	"context"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/terminal/crlf"

	"golang.org/x/term"
)

// feedBuffer decouples the producing goroutine from redraw latency.
const feedBuffer = 1024

// pageInteractive puts fd in raw mode and runs the key loop. If raw mode is
// unavailable it degrades to plain copying rather than failing.
func pageInteractive(ctx context.Context, w io.Writer, fd int, src LineSource, opt Options) error {
	old, err := term.MakeRaw(fd)
	if err != nil {
		return copyAll(ctx, w, src)
	}
	defer func() { _ = term.Restore(fd, old) }()

	rows, cols := terminalSize(fd, opt)
	input := opt.Input
	if input == nil {
		input = os.Stdin
	}

	feedCtx, cancelFeed := context.WithCancel(ctx)
	defer cancelFeed()

	s := newSession(crlf.NewWriter(w), src, feedLines(feedCtx, src), newKeyReader(input), rows, cols, opt)
	s.stopFeed = cancelFeed
	return s.run(ctx)
}

// feedLines pulls lines from src on its own goroutine so that the key loop
// stays responsive while a slow backend is still producing output. The
// goroutine ends when src is exhausted or ctx is cancelled; src.Err() is
// readable once the channel is closed.
func feedLines(ctx context.Context, src LineSource) <-chan string {
	lines := make(chan string, feedBuffer)
	go func() {
		defer close(lines)
		for {
			line, ok := src.NextLine()
			if !ok {
				return
			}
			select {
			case lines <- line:
			case <-ctx.Done():
				return
			}
		}
	}()
	return lines
}

// terminalSize resolves the screen size, preferring explicit Options over the
// terminal so that tests can pin a viewport, and falling back to a default
// size when neither is available.
func terminalSize(fd int, opt Options) (int, int) {
	rows, cols := opt.Rows, opt.Cols
	if rows <= 0 || cols <= 0 {
		termCols, termRows, err := term.GetSize(fd)
		if err == nil && rows <= 0 {
			rows = termRows
		}
		if err == nil && cols <= 0 {
			cols = termCols
		}
	}

	if rows <= 1 {
		rows = fallbackRows
	}
	if cols <= 0 {
		cols = fallbackCols
	}
	return rows, cols
}
