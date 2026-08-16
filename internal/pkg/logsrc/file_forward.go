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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// tailPollInterval is how often a --follow cursor re-stats the tailed file.
const tailPollInterval = 200 * time.Millisecond

// fileForwardCursor walks rotation files oldest-first, optionally tailing the
// newest one.
type fileForwardCursor struct {
	ctx    context.Context
	paths  []string
	idx    int
	sc     *bufio.Scanner
	closer io.Closer
	file   *os.File
	follow bool
	err    error
}

func openFileForward(ctx context.Context, paths []string, follow bool) (Cursor, error) {
	c := &fileForwardCursor{ctx: ctx, paths: paths, follow: follow}
	if err := c.openNext(); err != nil {
		return nil, err
	}
	return c, nil
}

// Next reports whether another record was read into rec.
func (c *fileForwardCursor) Next(rec *Record) bool {
	for {
		if c.ctx.Err() != nil {
			c.err = c.ctx.Err()
			return false
		}
		if c.sc == nil {
			return false
		}
		if c.sc.Scan() {
			parsed, ok := ParseSyslogLine(c.sc.Text())
			if !ok {
				continue
			}
			*rec = parsed
			return true
		}
		if !c.advance() {
			return false
		}
	}
}

// Err returns the first error that stopped iteration, or nil on clean EOF.
func (c *fileForwardCursor) Err() error { return c.err }

// Close releases the file currently open.
func (c *fileForwardCursor) Close() error {
	if c.closer == nil {
		return nil
	}
	closer := c.closer
	c.closer = nil
	c.sc = nil
	return closer.Close()
}

// advance moves to the next rotation file, or waits for more data when
// tailing. It reports whether iteration may continue.
func (c *fileForwardCursor) advance() bool {
	if err := c.sc.Err(); err != nil {
		c.err = err
		return false
	}
	if c.follow && c.file != nil && c.idx >= len(c.paths) {
		return c.tailWait()
	}
	if err := c.openNext(); err != nil {
		if err != io.EOF {
			c.err = err
		}
		return false
	}
	return true
}

// openNext opens the next readable path, skipping ones that vanished.
// Returns io.EOF when the list is exhausted, or ErrPermission when a file
// exists but cannot be read.
func (c *fileForwardCursor) openNext() error {
	_ = c.Close()

	for c.idx < len(c.paths) {
		path := c.paths[c.idx]
		c.idx++
		rc, err := openLogReader(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%w: %s: %v", ErrPermission, path, err)
			}
			continue
		}
		c.closer = rc
		c.file, _ = rc.(*os.File)
		c.sc = newLogScanner(rc)
		return nil
	}
	return io.EOF
}

// tailWait blocks until the tailed file grows, is truncated, or is rotated.
func (c *fileForwardCursor) tailWait() bool {
	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.err = c.ctx.Err()
			return false
		case <-ticker.C:
			if grown, ok := c.pollTail(); ok {
				return grown
			}
		}
	}
}

// pollTail inspects the tailed file once. The second result reports whether a
// decision was reached; the first is what Next should do.
func (c *fileForwardCursor) pollTail() (bool, bool) {
	path := c.paths[len(c.paths)-1]
	pathInfo, pathErr := os.Stat(path)
	info, err := c.file.Stat()
	if err != nil {
		return c.reopenTail(), true
	}
	// logrotate replaces the path with a new inode while this handle still
	// points at the old (often deleted) file; reopen onto the live path.
	if pathErr == nil && !os.SameFile(info, pathInfo) {
		return c.reopenTail(), true
	}

	off, err := c.file.Seek(0, io.SeekCurrent)
	if err != nil {
		c.err = err
		return false, true
	}
	if info.Size() > off {
		c.sc = newLogScanner(c.file)
		return true, true
	}
	if info.Size() < off {
		if _, err := c.file.Seek(0, io.SeekStart); err != nil {
			c.err = err
			return false, true
		}
		c.sc = newLogScanner(c.file)
		return true, true
	}
	return false, false
}

// reopenTail reopens the newest path after a rotation.
func (c *fileForwardCursor) reopenTail() bool {
	path := c.paths[len(c.paths)-1]
	_ = c.Close()

	f, err := os.Open(path)
	if err != nil {
		c.err = err
		return false
	}
	c.file = f
	c.closer = f
	c.sc = newLogScanner(f)
	return true
}
