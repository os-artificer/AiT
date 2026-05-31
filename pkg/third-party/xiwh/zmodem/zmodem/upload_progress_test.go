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

import "testing"

// TestOnUploadProgress_ChunkLoop mirrors sender.go ZRPOS WriteFunc progress reporting (U31).
func TestOnUploadProgress_ChunkLoop(t *testing.T) {
	size := 20000
	writeCount := 0
	var calls []int64

	consumer := ZModemConsumer{
		OnUploadProgress: func(written, total int64) {
			calls = append(calls, written)
			if total != int64(size) {
				t.Fatalf("total = %d, want %d", total, size)
			}
		},
	}

	chunks := []int{8096, 8096, 3808}
	for _, chunk := range chunks {
		n := chunk
		writeCount += n
		if consumer.OnUploadProgress != nil && size > 0 {
			consumer.OnUploadProgress(int64(writeCount), int64(size))
		}
	}

	if len(calls) != 3 {
		t.Fatalf("expected 3 progress calls, got %d", len(calls))
	}
	if calls[len(calls)-1] != int64(size) {
		t.Fatalf("final written = %d, want %d", calls[len(calls)-1], size)
	}
}
