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

package client

import "testing"

// TestFormatTransferDone_GetPut (C19) locks the done: completion line format.
func TestFormatTransferDone_GetPut(t *testing.T) {
	gotGet := formatTransferDone("get", "/remote", "./local", 100)
	wantGet := "done: get /remote → ./local (100 bytes)"
	if gotGet != wantGet {
		t.Fatalf("get=%q want %q", gotGet, wantGet)
	}

	gotPut := formatTransferDone("put", "./local", "/remote", 42)
	wantPut := "done: put ./local → /remote (42 bytes)"
	if gotPut != wantPut {
		t.Fatalf("put=%q want %q", gotPut, wantPut)
	}
}
