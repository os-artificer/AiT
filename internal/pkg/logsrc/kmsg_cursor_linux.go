//go:build linux

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
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// kmsgCursor iterates the kernel ring buffer one record per line.
type kmsgCursor struct {
	ctx    context.Context
	f      *os.File
	sc     *bufio.Scanner
	boot   time.Time
	path   string
	follow bool
	err    error
}

// Next reads the next parsable record into rec, skipping lines the kernel
// format does not cover. It reports false at the end of the buffer, on a read
// error, or once the context is cancelled. Non-follow mode treats EAGAIN after
// a drained ring as a clean EOF.
func (c *kmsgCursor) Next(rec *Record) bool {
	for {
		if c.ctx.Err() != nil {
			c.err = c.ctx.Err()
			return false
		}
		if !c.sc.Scan() {
			c.err = mapKmsgScanErr(c.sc.Err(), c.follow)
			return false
		}

		parsed, ok := ParseKmsgLine(c.sc.Text(), c.boot)
		if !ok {
			continue
		}
		*rec = parsed
		return true
	}
}

// Err returns the read or cancellation error that stopped iteration, or nil at
// the end of the buffer.
func (c *kmsgCursor) Err() error { return c.err }

// Close closes the underlying device.
func (c *kmsgCursor) Close() error {
	if c.f == nil {
		return nil
	}
	return c.f.Close()
}

// mapKmsgScanErr turns a non-blocking EAGAIN into a clean EOF when not
// following; other errors are returned as-is.
func mapKmsgScanErr(err error, follow bool) error {
	if err == nil {
		return nil
	}
	if !follow && (errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK)) {
		return nil
	}
	return err
}

// bootTime derives the wall-clock time the machine started from /proc/uptime.
// It returns the zero time when uptime is unreadable, which leaves kmsg records
// without a timestamp rather than with a wrong one.
func bootTime() time.Time {
	data, err := os.ReadFile(procUptime)
	if err != nil {
		return time.Time{}
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return time.Time{}
	}
	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return time.Time{}
	}
	return time.Now().Add(-time.Duration(secs * float64(time.Second)))
}
