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
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// execScanBufMax bounds one stdout document (journal messages can be large).
const execScanBufMax = 10 * 1024 * 1024

// LineParser converts one stdout document into a Record. Returning an error
// skips the document.
type LineParser func(doc []byte) (Record, error)

// LineFramer joins consecutive stdout lines into one document. It returns the
// completed document and true once a record is ready, and false while more
// lines are needed. Backends whose output is one record per line omit it.
type LineFramer func(line []byte) ([]byte, bool)

// ExecCursorOptions configures an exec-backed Cursor.
type ExecCursorOptions struct {
	Path   string
	Args   []string
	Parse  LineParser
	Frame  LineFramer
	MapErr func(stderr string, waitErr error) error
}

type execCursor struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr bytes.Buffer
	sc     *bufio.Scanner
	parse  LineParser
	frame  LineFramer
	mapErr func(string, error) error
	err    error
	closed bool
	waited bool
	mu     sync.Mutex
	once   sync.Once
}

// OpenExecCursor starts opt.Path with opt.Args and returns a Cursor over its
// stdout. Returns ErrToolMissing when opt.Path is empty or the binary cannot
// be executed, or the underlying pipe/start error.
func OpenExecCursor(ctx context.Context, opt ExecCursorOptions) (Cursor, error) {
	if opt.Path == "" {
		return nil, ErrToolMissing
	}
	frame := opt.Frame
	if frame == nil {
		frame = frameSingleLine
	}
	c := &execCursor{
		cmd:    exec.CommandContext(ctx, opt.Path, opt.Args...),
		parse:  opt.Parse,
		frame:  frame,
		mapErr: opt.MapErr,
	}
	c.cmd.Stderr = &c.stderr

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	c.stdout = stdout

	if err := c.cmd.Start(); err != nil {
		_ = stdout.Close()
		return nil, mapStartErr(err)
	}
	c.sc = bufio.NewScanner(stdout)
	c.sc.Buffer(make([]byte, 0, 64*1024), execScanBufMax)
	return c, nil
}

// Next reports whether another record was read into rec.
func (c *execCursor) Next(rec *Record) bool {
	for c.sc.Scan() {
		doc, ready := c.frame(append([]byte(nil), c.sc.Bytes()...))
		if !ready {
			continue
		}
		parsed, err := c.parse(doc)
		if err != nil {
			continue
		}
		*rec = parsed
		return true
	}
	if err := c.sc.Err(); err != nil {
		c.err = err
	}
	c.finish()
	return false
}

// Err returns the first error that stopped iteration, or nil on clean EOF.
func (c *execCursor) Err() error { return c.err }

// Close kills the child process if it is still running and reaps it. Stopping
// early (a satisfied -n, or the user quitting the pager) is not an error, so
// the resulting wait status is discarded.
func (c *execCursor) Close() error {
	c.once.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		if c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		c.finish()
	})
	return c.err
}

// finish reaps the process once and maps its exit status to a domain error.
func (c *execCursor) finish() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.waited {
		return
	}
	c.waited = true

	_ = c.stdout.Close()
	waitErr := c.cmd.Wait()
	if waitErr == nil || c.err != nil || c.closed {
		return
	}

	stderr := strings.TrimSpace(c.stderr.String())
	if c.mapErr != nil {
		c.err = c.mapErr(stderr, waitErr)
		return
	}
	c.err = MapExecWaitErr(stderr, waitErr)
}
