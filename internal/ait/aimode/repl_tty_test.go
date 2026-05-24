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

package aimode

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/os-artificer/AiT/internal/ait/shell"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestAIMode_ConfirmNoHistory(t *testing.T) {
	hist := terminal.NewLineHistory()
	hist.Add("run ls")

	br := bufio.NewReader(bytes.NewReader([]byte("y\r\x1b[A\r")))
	_, err := shell.ReadTTYLineNoHistory(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}

	line, err := shell.ReadTTYLine(br, io.Discard, "AiT AI> ", hist)
	if err != nil {
		t.Fatal(err)
	}
	if line != "run ls" {
		t.Fatalf("line = %q, want run ls (confirm y must not pollute history)", line)
	}
}
