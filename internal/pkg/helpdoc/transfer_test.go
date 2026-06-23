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

package helpdoc

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTransferHelp_Get(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTransferHelp(&buf, "get"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:",
		SectionPaths,
		SectionOpenSSHFlagsTitle,
		"Not supported by aitssh get",
		"Examples:",
		"help for get",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("get transfer help missing %q:\n%s", want, out)
		}
	}
}
