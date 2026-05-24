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

func TestScrollDelta_Prefix(t *testing.T) {
	before := "line1\n"
	after := "line1\nline2\n"
	got := scrollDelta(before, after)
	if got != "line2" {
		t.Fatalf("got %q", got)
	}
}

func TestScrollDelta_Unchanged(t *testing.T) {
	if scrollDelta("same", "same") != "" {
		t.Fatal("expected empty delta")
	}
}

func TestScrollDelta_ShorterAfter(t *testing.T) {
	got := scrollDelta("longer before", "x")
	if got != "x" {
		t.Fatalf("got %q", got)
	}
}
