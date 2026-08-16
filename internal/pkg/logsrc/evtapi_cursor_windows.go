//go:build windows

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
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// evtNextTimeout is the EvtNext wait in milliseconds. A query (unlike a
// subscription) has all of its events available, so this only bounds a
// pathological wait.
const evtNextTimeout = 5000

// evtAPICursor walks the events of one EvtQuery result handle.
type evtAPICursor struct {
	ctx     context.Context
	result  windows.Handle
	channel string
	err     error
}

// Next renders the next event into rec, and reports false at the end of the
// result set or on the first failure.
func (c *evtAPICursor) Next(rec *Record) bool {
	if c.ctx.Err() != nil {
		c.err = c.ctx.Err()
		return false
	}

	handle, ok := c.nextHandle()
	if !ok {
		return false
	}
	defer procEvtClose.Call(uintptr(handle))

	xml, err := evtRenderXML(handle)
	if err != nil {
		c.err = err
		return false
	}
	parsed, err := ParseEventXML(xml)
	if err != nil {
		return false
	}
	parsed.Source = c.channel
	*rec = parsed
	return true
}

// Err returns the first error that stopped iteration, or nil at the end of the
// result set.
func (c *evtAPICursor) Err() error { return c.err }

// Close releases the query result handle.
func (c *evtAPICursor) Close() error {
	if c.result == 0 {
		return nil
	}
	procEvtClose.Call(uintptr(c.result))
	c.result = 0
	return nil
}

// nextHandle fetches one event handle. It reports false at the end of the
// result set, recording an error only for genuine failures.
func (c *evtAPICursor) nextHandle() (windows.Handle, bool) {
	var handles [1]windows.Handle
	var returned uint32

	ret, _, callErr := procEvtNext.Call(
		uintptr(c.result),
		1,
		uintptr(unsafe.Pointer(&handles[0])),
		evtNextTimeout,
		0,
		uintptr(unsafe.Pointer(&returned)),
	)
	if ret != 0 && returned > 0 {
		return handles[0], true
	}

	errno, ok := callErr.(syscall.Errno)
	if !ok || errno == 0 || errno == windows.ERROR_NO_MORE_ITEMS {
		return 0, false
	}
	c.err = mapWindowsEvtErr(callErr)
	return 0, false
}
