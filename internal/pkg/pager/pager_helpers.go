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
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// writerTerminal reports w's file descriptor and whether it is a terminal.
// Writers that are not *os.File are never terminals.
func writerTerminal(w io.Writer) (int, bool) {
	f, ok := w.(*os.File)
	if !ok {
		return 0, false
	}
	fd := int(f.Fd())
	return fd, term.IsTerminal(fd)
}

// copyAll writes every line of src to w, the passthrough used when paging is
// disabled or w is not a terminal. Returns src.Err(), a write error, or nil
// once ctx is cancelled.
func copyAll(ctx context.Context, w io.Writer, src LineSource) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		line, ok := src.NextLine()
		if !ok {
			return src.Err()
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
}
