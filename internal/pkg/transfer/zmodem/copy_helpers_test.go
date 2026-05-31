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

package zmodem

import (
	"bytes"
	"testing"
)

func TestCopyWithProgress_Incremental(t *testing.T) {
	data := make([]byte, copyBufSize*2+copyBufSize/2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	var calls int
	var lastWritten int64
	progress := ProgressFunc(func(written, total int64) {
		calls++
		lastWritten = written
		if total != int64(len(data)) {
			t.Fatalf("total = %d, want %d", total, len(data))
		}
	})

	dst := &bytes.Buffer{}
	n, err := copyWithProgress(dst, bytes.NewReader(data), int64(len(data)), progress)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(data)) {
		t.Fatalf("copied %d bytes, want %d", n, len(data))
	}
	if calls < 2 {
		t.Fatalf("expected multiple progress callbacks, got %d", calls)
	}
	if lastWritten != int64(len(data)) {
		t.Fatalf("last written = %d, want %d", lastWritten, len(data))
	}
	if !bytes.Equal(dst.Bytes(), data) {
		t.Fatal("destination mismatch")
	}
}
