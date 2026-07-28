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

package logsrc

import (
	"context"
	"fmt"
	"io"
	"os"
)

// fileReverseCursor walks rotation files newest-first, and each file from its
// last line backwards.
type fileReverseCursor struct {
	ctx      context.Context
	paths    []string
	idx      int
	scanner  *reverseScanner
	err      error
	finished bool
}

func openFileReverse(ctx context.Context, pathsNewestFirst []string) (Cursor, error) {
	c := &fileReverseCursor{ctx: ctx, paths: pathsNewestFirst}
	if err := c.openNext(); err != nil {
		if err == io.EOF {
			c.finished = true
			return c, nil
		}
		return nil, err
	}
	return c, nil
}

// Next reports whether another record was read into rec.
func (c *fileReverseCursor) Next(rec *Record) bool {
	for !c.finished {
		if c.ctx.Err() != nil {
			c.err = c.ctx.Err()
			return false
		}

		line, ok, err := c.scanner.next()
		if err != nil {
			c.err = err
			return false
		}
		if !ok {
			if err := c.openNext(); err != nil {
				c.finished = true
				if err != io.EOF {
					c.err = err
				}
				return false
			}
			continue
		}
		parsed, ok := ParseSyslogLine(line)
		if !ok {
			continue
		}
		*rec = parsed
		return true
	}
	return false
}

// Err returns the first error that stopped iteration, or nil on clean EOF.
func (c *fileReverseCursor) Err() error { return c.err }

// Close releases the file currently open.
func (c *fileReverseCursor) Close() error {
	if c.scanner == nil {
		return nil
	}
	s := c.scanner
	c.scanner = nil
	return s.close()
}

// openNext opens the next older file, skipping ones that vanished. Returns
// io.EOF when the list is exhausted, or ErrPermission when a file exists but
// cannot be read.
func (c *fileReverseCursor) openNext() error {
	_ = c.Close()

	for c.idx < len(c.paths) {
		path := c.paths[c.idx]
		c.idx++
		s, err := newReverseScanner(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%w: %s: %v", ErrPermission, path, err)
			}
			continue
		}
		c.scanner = s
		return nil
	}
	return io.EOF
}
