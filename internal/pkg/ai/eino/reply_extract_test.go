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

package eino

import (
	"strings"
	"testing"
)

func TestExtractReply_StripsPartialAISshCLI(t *testing.T) {
	raw := `Done. {"action":"aish_cli","args":"list`
	reply, _ := ExtractReply(raw)
	if strings.Contains(reply, `{"action":"aish_cli"`) {
		t.Fatalf("partial aish_cli JSON should be stripped, got %q", reply)
	}
	if reply != "Done." {
		t.Fatalf("reply=%q want Done.", reply)
	}
}
