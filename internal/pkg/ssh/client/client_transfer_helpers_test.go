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

// TestJoinRemotePath_RelativeUnderCwd (TR1) joins relative remote paths under shell cwd.
func TestJoinRemotePath_RelativeUnderCwd(t *testing.T) {
	got := joinRemotePath("/data/dev", "./bin/f")
	if got != "/data/dev/bin/f" {
		t.Fatalf("got %q want /data/dev/bin/f", got)
	}
	if !isRelativeRemotePath("./bin/f") {
		t.Fatal("expected relative")
	}
	if isRelativeRemotePath("/abs") || isRelativeRemotePath("~/x") {
		t.Fatal("expected absolute/tilde paths to be non-relative")
	}
}

// TestWrapTransferWrongHop_SkipsSameHostAfterResolve (TR6) avoids wrong-hop on same active host.
func TestWrapTransferWrongHop_SkipsSameHostAfterResolve(t *testing.T) {
	if shouldWrapTransferWrongHop("dev.example", "devhost", "devhost") {
		t.Fatal("same host should not wrap wrong-hop")
	}
	if !shouldWrapTransferWrongHop("bastion", "", "internal") {
		t.Fatal("nested host should wrap wrong-hop")
	}
}
