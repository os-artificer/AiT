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

package sshclient

import (
	"time"

	"golang.org/x/term"
)

// watchResize polls window size on Windows (no SIGWINCH) and pushes changes
// to the remote PTY and local shell.
func watchResize(h resizeHandler, fd int) func() {
	stop := make(chan struct{})
	go func() {
		lastCols, lastRows, _ := term.GetSize(fd)
		t := time.NewTicker(750 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				cols, rows, err := term.GetSize(fd)
				if err != nil {
					continue
				}
				if cols != lastCols || rows != lastRows {
					h.onResize(cols, rows)
					lastCols, lastRows = cols, rows
				}
			}
		}
	}()
	return func() { close(stop) }
}
