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

package ait

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintMetaHelp(t *testing.T) {
	var buf bytes.Buffer
	PrintMetaHelp(&buf)
	out := buf.String()
	for _, want := range []string{"help", "ai <question>", "local", "suspend", "!ait list", "ait shell"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q:\n%s", want, out)
		}
	}
}

func TestExecuteMetaHelp(t *testing.T) {
	if err := ExecuteMeta(nil); err != nil {
		t.Fatalf("ExecuteMeta(nil): %v", err)
	}
	if err := ExecuteMeta([]string{"help"}); err != nil {
		t.Fatalf("ExecuteMeta(help): %v", err)
	}
}

func TestExecuteMetaBlockedLogin(t *testing.T) {
	err := ExecuteMeta([]string{"login", "--host", "h"})
	if err == nil {
		t.Fatal("expected error for login in session")
	}
	if !strings.Contains(err.Error(), "exit the SSH") {
		t.Errorf("unexpected error: %v", err)
	}
}
