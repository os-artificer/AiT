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

package terminal

import (
	"io"
	"os"
	"time"
)

func drainStdinFD(fd int) {
	f := os.NewFile(uintptr(fd), "stdin")
	if f == nil {
		return
	}

	for {
		if err := f.SetReadDeadline(time.Now().Add(2 * time.Millisecond)); err != nil {
			break
		}
		buf := make([]byte, 256)
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			break
		}
		if err == io.EOF {
			break
		}
	}
	_ = f.SetReadDeadline(time.Time{})
}
