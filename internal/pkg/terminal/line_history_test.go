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

import "testing"

func TestShouldRecordLine(t *testing.T) {
	if shouldRecordLine("") || shouldRecordLine("   ") {
		t.Fatal("empty should not record")
	}
	if !shouldRecordLine("list") {
		t.Fatal("list should record")
	}
	if shouldRecordLine("export AIT_LOCAL_MODE=1") {
		t.Fatal("AIT env export should not record")
	}
	if shouldRecordLine("PS1=$ ") {
		t.Fatal("PS1 should not record")
	}
}
