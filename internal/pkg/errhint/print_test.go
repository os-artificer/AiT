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

package errhint

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

func TestPrint_CLI_NoPrefix(t *testing.T) {
	var buf bytes.Buffer
	err := WithTry(store.ErrSessionNotFound, "explain", "ait session --help")
	Print(&buf, SurfaceCLI, err)
	out := buf.String()
	if strings.Contains(out, "AiT:") {
		t.Fatalf("CLI should not prefix AiT: %q", out)
	}
	if !strings.Contains(out, "Try:") {
		t.Fatalf("missing Try: %q", out)
	}
}

func TestPrint_TTY_Multiline(t *testing.T) {
	var buf bytes.Buffer
	err := WithTry(errors.New("session not found"), "The list ID column is for display only.",
		"ait session --load <alias>",
	)
	Print(&buf, SurfaceTTY, err)
	out := buf.String()
	if !strings.HasPrefix(out, "AiT: session not found") {
		t.Fatalf("first line prefix: %q", out)
	}
	if strings.Count(out, "AiT:") != 1 {
		t.Fatalf("duplicate AiT: prefix: %q", out)
	}
	if strings.Contains(out, "display only AiT:") {
		t.Fatalf("collapsed lines: %q", out)
	}
	if !strings.Contains(out, "Try:") {
		t.Fatalf("missing Try: %q", out)
	}
}
