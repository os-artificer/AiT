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

package clihints

import "testing"

func TestTryLines_ConnectSaveVariantsDiffer(t *testing.T) {
	if TryConnectSave == "" || TryConnectSaveUserHost == "" {
		t.Fatal("connect save try lines must be non-empty")
	}
	if TryConnectSave == TryConnectSaveUserHost {
		t.Fatalf("session and openssh connect save hints must differ: %q", TryConnectSave)
	}
	if TryConnectSave != "ai-ssh --save <destination>" {
		t.Fatalf("TryConnectSave: got %q", TryConnectSave)
	}
	if TryConnectSaveUserHost != "ai-ssh --save user@host" {
		t.Fatalf("TryConnectSaveUserHost: got %q", TryConnectSaveUserHost)
	}
}

func TestTryLines_SessionAndTransferNonEmpty(t *testing.T) {
	lines := []string{
		TrySessionList,
		TryConnectExample,
		TryGetHelp,
		TryPutHelp,
		TryGetExample,
		TryShellList,
	}
	for _, line := range lines {
		if line == "" {
			t.Fatal("expected non-empty try line")
		}
	}
}
