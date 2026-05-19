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

package pathsafe

import "testing"

func TestRemotePath(t *testing.T) {
	if !RemotePath("/home/user/proj") {
		t.Fatal("expected valid path")
	}
	if RemotePath("/tmp/foo\nid") {
		t.Fatal("newline rejected")
	}
	if RemotePath("/tmp/foo;rm") {
		t.Fatal("semicolon rejected")
	}
	if RemotePath("/tmp/$(x)") {
		t.Fatal("metachar rejected")
	}
	if RemotePath("relative") {
		t.Fatal("relative rejected")
	}
	if RemotePath("/foo/../bar") {
		t.Fatal(".. rejected")
	}
}
