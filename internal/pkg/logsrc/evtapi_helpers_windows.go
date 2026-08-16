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
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// evtRenderFallbackBytes sizes the render buffer when the probing call does not
// report a required size.
const evtRenderFallbackBytes = 64 * 1024

// evtRenderXML renders one event handle as UTF-8 XML. The first call asks
// EvtRender for the buffer size it needs; the second one fills it.
func evtRenderXML(h windows.Handle) ([]byte, error) {
	var used, propCount uint32
	procEvtRender.Call(0, uintptr(h), evtRenderEventXML, 0, 0,
		uintptr(unsafe.Pointer(&used)), uintptr(unsafe.Pointer(&propCount)))
	if used == 0 {
		used = evtRenderFallbackBytes
	}

	buf := make([]uint16, used/2+1)
	ret, _, callErr := procEvtRender.Call(0, uintptr(h), evtRenderEventXML,
		uintptr(len(buf)*2), uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&used)), uintptr(unsafe.Pointer(&propCount)))
	if ret == 0 {
		return nil, mapWindowsEvtErr(callErr)
	}

	n := used / 2
	if n > 0 && buf[n-1] == 0 {
		n--
	}
	return []byte(windows.UTF16ToString(buf[:n])), nil
}

// mapWindowsEvtErr maps a Win32 error to a domain error so that the resolver
// can fall back or hint appropriately.
func mapWindowsEvtErr(err error) error {
	errno, ok := err.(syscall.Errno)
	if !ok || errno == 0 {
		return ErrNoSource
	}
	switch errno {
	case windows.ERROR_ACCESS_DENIED:
		return fmt.Errorf("%w: %v", ErrPermission, err)
	case windows.ERROR_FILE_NOT_FOUND, windows.ERROR_PATH_NOT_FOUND, windows.ERROR_EVT_CHANNEL_NOT_FOUND:
		return fmt.Errorf("%w: %v", ErrNoSource, err)
	default:
		return err
	}
}
